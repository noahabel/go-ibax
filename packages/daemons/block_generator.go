/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package daemons

import (
	"bytes"
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/IBAX-io/go-ibax/packages/block"
	"github.com/IBAX-io/go-ibax/packages/conf"
	"github.com/IBAX-io/go-ibax/packages/conf/syspar"
	"github.com/IBAX-io/go-ibax/packages/consts"
	"github.com/IBAX-io/go-ibax/packages/model"
	"github.com/IBAX-io/go-ibax/packages/protocols"
	"github.com/IBAX-io/go-ibax/packages/service"
	"github.com/IBAX-io/go-ibax/packages/transaction"
	"github.com/IBAX-io/go-ibax/packages/utils"

	log "github.com/sirupsen/logrus"
)

// BlockGenerator is daemon that generates blocks
func BlockGenerator(ctx context.Context, d *daemon) error {
	if atomic.CompareAndSwapUint32(&d.atomic, 0, 1) {
		defer atomic.StoreUint32(&d.atomic, 0)
	} else {
		return nil
	}
	d.sleepTime = time.Second
	if service.IsNodePaused() {
		return nil
	}

	nodePosition, err := syspar.GetThisNodePosition()
	if err != nil {
		// we are not honor node and can't generate new blocks
		d.sleepTime = 4 * time.Second
		d.logger.WithFields(log.Fields{"type": consts.JustWaiting, "error": err}).Debug("we are not honor node, sleep for 10 seconds")
		return nil
	}

	DBLock()
	defer DBUnlock()

	// wee need fresh myNodePosition after locking
	nodePosition, err = syspar.GetThisNodePosition()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting node position by key id")
		return err
	}

	btc := protocols.NewBlockTimeCounter()
	st := time.Now()

	if exists, err := btc.BlockForTimeExists(st, int(nodePosition)); exists || err != nil {
		return nil
	}

	timeToGenerate, err := btc.TimeToGenerate(st, int(nodePosition))
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.BlockError, "error": err, "position": nodePosition}).Debug("calculating block time")
		return err
	}

	if !timeToGenerate {
		d.logger.WithFields(log.Fields{"type": consts.JustWaiting}).Debug("not my generation time")
		return nil
	}

	//if !NtpDriftFlag {
	//	d.logger.WithFields(log.Fields{"type": consts.Ntpdate}).Error("ntp time not ntpdate")
	//	return nil
	//}

	//var cf model.Confirmation
	//cfg, err := cf.CheckAllowGenBlock()
	//if err != nil {
	//	d.logger.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Debug("confirmation block not allow")
	//	return err
	//}
	//
	//if !cfg {
	//	d.logger.WithFields(log.Fields{"type": consts.JustWaiting}).Debug("not my confirmation time")
	//	return nil
	//}

	_, endTime, err := btc.RangeByTime(st)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.TimeCalcError, "error": err}).Error("on getting end time of generation")
		return err
	}

	done := time.After(endTime.Sub(st))
	prevBlock := &model.InfoBlock{}
	_, err = prevBlock.Get()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting previous block")
		return err
	}

	NodePrivateKey, NodePublicKey := utils.GetNodeKeys()
	if len(NodePrivateKey) < 1 {
		d.logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
		return errors.New(`node private key is empty`)
	}

	dtx := DelayedTx{
		privateKey: NodePrivateKey,
		publicKey:  NodePublicKey,
		logger:     d.logger,
		time:       st.Unix(),
	}

	txs, err := dtx.RunForDelayBlockID(prevBlock.BlockID + 1)
	if err != nil {
		return err
	}

	trs, err := processTransactions(d.logger, txs, done, st.Unix())
	if err != nil {
		return err
	}

	// Block generation will be started only if we have transactions
	if len(trs) == 0 {
		return nil
	}

	header := &utils.BlockData{
		BlockID:      prevBlock.BlockID + 1,
		Time:         st.Unix(),
		EcosystemID:  0,
		log.WithFields(log.Fields{"error": err}).Error("on inserting new block")
		return err
	}
	log.WithFields(log.Fields{"Block": header.String(), "type": consts.SyncProcess}).Debug("Generated block ID")

	//go notificator.CheckTokenMovementLimits(nil, conf.Config.TokenMovement, header.BlockID)
	return nil
}

func generateNextBlock(blockHeader *utils.BlockData, trs []*model.Transaction, key string, prevBlock *utils.BlockData) ([]byte, error) {
	trData := make([][]byte, 0, len(trs))
	for _, tr := range trs {
		trData = append(trData, tr.Data)
	}

	return block.MarshallBlock(blockHeader, trData, prevBlock, key)
}

func processTransactions(logger *log.Entry, txs []*model.Transaction, done <-chan time.Time, st int64) ([]*model.Transaction, error) {
	p := new(transaction.Transaction)

	//verify transactions
	//err := transaction.ProcessTransactionsQueue(p.DbTransaction)
	//if err != nil {
	//	return nil, err
	//}

	trs, err := model.GetAllUnusedTransactions(p.DbTransaction, syspar.GetMaxTxCount())
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all unused transactions")
		return nil, err
	}

	limits := block.NewLimits(nil)

	type badTxStruct struct {
		hash  []byte
		msg   string
		keyID int64
	}

	processBadTx := func(dbTx *model.DbTransaction) chan badTxStruct {
		ch := make(chan badTxStruct)

		go func() {
			for badTxItem := range ch {
				block.BadTxForBan(badTxItem.keyID)
				_ = transaction.MarkTransactionBad(dbTx, badTxItem.hash, badTxItem.msg)
			}
		}()

		return ch
	}

	txBadChan := processBadTx(p.DbTransaction)

	defer func() {
		close(txBadChan)
	}()

	// Checks preprocessing count limits
	txList := make([]*model.Transaction, 0, len(trs))
	txs = append(txs, trs...)
	for i, txItem := range txs {
		select {
		case <-done:
			return txList, nil
		default:
			if txItem.GetTransactionRateStopNetwork() {
				txList = append(txList[:0], txs[i])
				break
			}
			bufTransaction := bytes.NewBuffer(txItem.Data)
			tr, err := transaction.UnmarshallTransaction(bufTransaction, true)
			if err != nil {
				if tr != nil {
					txBadChan <- badTxStruct{hash: tr.TxHash, msg: err.Error(), keyID: tr.TxHeader.KeyID}
				}
				continue
			}

			//if err := p.Check(st, false); err != nil {
			//	txBadChan <- badTxStruct{hash: p.TxHash, msg: err.Error(), keyID: p.TxHeader.KeyID}
			//	continue
			//}

			if tr.TxSmart != nil {
				err = limits.CheckLimit(tr)
				if err == block.ErrLimitStop && i > 0 {
					break
				} else if err != nil {
					if err != block.ErrLimitSkip {
						txBadChan <- badTxStruct{hash: tr.TxHash, msg: err.Error(), keyID: tr.TxHeader.KeyID}
					}
					continue
				}
			}
			txList = append(txList, txs[i])
		}
	}
	return txList, nil
}
