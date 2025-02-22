/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
package rollback

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/IBAX-io/go-ibax/packages/block"
	"github.com/IBAX-io/go-ibax/packages/consts"
	"github.com/IBAX-io/go-ibax/packages/model"
	"github.com/IBAX-io/go-ibax/packages/transaction"
	"github.com/IBAX-io/go-ibax/packages/utils"

	log "github.com/sirupsen/logrus"
)

var (
	ErrLastBlock = errors.New("Block is not the last")
)

// BlockRollback is blocking rollback
func RollbackBlock(data []byte) error {
	bl, err := block.UnmarshallBlock(bytes.NewBuffer(data), true)
	if err != nil {
		return err
	}

	b := &model.Block{}
	if _, err = b.GetMaxBlock(); err != nil {
		return err
	}

	if b.ID != bl.Header.BlockID {
		return ErrLastBlock
	}

		return err
	}

	b = &model.Block{}
	if _, err = b.Get(bl.Header.BlockID - 1); err != nil {
		dbTransaction.Rollback()
		return err
	}

	bl, err = block.UnmarshallBlock(bytes.NewBuffer(b.Data), false)
	if err != nil {
		dbTransaction.Rollback()
		return err
	}

	ib := &model.InfoBlock{
		Hash:           b.Hash,
		RollbacksHash:  b.RollbacksHash,
		BlockID:        b.ID,
		NodePosition:   strconv.Itoa(int(b.NodePosition)),
		KeyID:          b.KeyID,
		Time:           b.Time,
		CurrentVersion: strconv.Itoa(bl.Header.Version),
	}
	err = ib.Update(dbTransaction)
	if err != nil {
		dbTransaction.Rollback()
		return err
	}

	return dbTransaction.Commit()
}

func rollbackBlock(dbTransaction *model.DbTransaction, block *block.Block) error {
	// rollback transactions in reverse order
	logger := block.GetLogger()
	for i := len(block.Transactions) - 1; i >= 0; i-- {
		t := block.Transactions[i]
		t.DbTransaction = dbTransaction

		_, err := model.MarkTransactionUnusedAndUnverified(dbTransaction, t.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("starting transaction")
			return err
		}
		_, err = model.DeleteLogTransactionsByHash(dbTransaction, t.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting log transactions by hash")
			return err
		}

		ts := &model.TransactionStatus{}
		err = ts.UpdateBlockID(dbTransaction, 0, t.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating block id in transaction status")
			return err
		}

		_, err = model.DeleteQueueTxByHash(dbTransaction, t.TxHash)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transacion from queue by hash")
			return err
		}

		if t.TxContract != nil {
			if err = rollbackTransaction(t.TxHash, t.DbTransaction, logger); err != nil {
				return err
			}
		} else {
			MethodName := consts.TxTypes[t.TxType]
			txParser, err := transaction.GetTransaction(t, MethodName)
			if err != nil {
				return utils.ErrInfo(err)
			}
			result := txParser.Init()
			if _, ok := result.(error); ok {
				return utils.ErrInfo(result.(error))
			}
			result = txParser.Rollback()
			if _, ok := result.(error); ok {
				return utils.ErrInfo(result.(error))
			}
		}
	}

	return nil
}
