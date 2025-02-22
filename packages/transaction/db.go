/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
package transaction

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/IBAX-io/go-ibax/packages/conf/syspar"
	"github.com/IBAX-io/go-ibax/packages/consts"
	"github.com/IBAX-io/go-ibax/packages/model"
	"github.com/IBAX-io/go-ibax/packages/utils"

	log "github.com/sirupsen/logrus"
)

var (
	ErrDuplicatedTx = errors.New("Duplicated transaction")
	ErrNotComeTime  = errors.New("Transaction processing time has not come")
	ErrExpiredTime  = errors.New("Transaction processing time is expired")
	ErrEarlyTime    = utils.WithBan(errors.New("Early transaction time"))
	ErrEmptyKey     = utils.WithBan(errors.New("KeyID is empty"))
)

// InsertInLogTx is inserting tx in log
func InsertInLogTx(t *Transaction, blockID int64) error {
	ltx := &model.LogTransaction{Hash: t.TxHash, Block: blockID}
	if err := ltx.Create(t.DbTransaction); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("insert logged transaction")
		return utils.ErrInfo(err)
	}
	return nil
}

// CheckLogTx checks if this transaction exists
// And it would have successfully passed a frontal test
func CheckLogTx(txHash []byte, transactions, txQueue bool) error {
	logTx := &model.LogTransaction{}
	found, err := logTx.GetByHash(txHash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting log transaction by hash")
		return err
	}
	if found {
		log.WithFields(log.Fields{"tx_hash": txHash, "type": consts.DuplicateObject}).Error("double tx in log transactions")
		return ErrDuplicatedTx
	}

	if transactions {
		// check for duplicate transaction
		tx := &model.Transaction{}
		isfound, err := tx.GetVerified(txHash)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting verified transaction")
			return utils.ErrInfo(err)
		}
		if isfound {
			log.WithFields(log.Fields{"tx_hash": tx.Hash, "type": consts.DuplicateObject}).Error("double tx in transactions")
			return ErrDuplicatedTx
		}
	}

	return nil
}

// DeleteQueueTx deletes a transaction from the queue
func DeleteQueueTx(dbTransaction *model.DbTransaction, hash []byte) error {
	delQueueTx := &model.QueueTx{Hash: hash}
	err := delQueueTx.DeleteTx(dbTransaction)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Debug("deleting transaction from queue")
		return err
	}
	// Because we process transactions with verified=0 in queue_parser_tx, after processing we need to delete them
	err = model.DeleteTransactionByHash(dbTransaction, hash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Debug("deleting transaction if unused")
		return err
	}
	//err = model.DeleteTransactionsAttemptsByHash(dbTransaction, hash)
	//if err != nil {
	//	log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Debug("deleting DeleteTransactionsAttemptsByHash")
	//	return err
	//}
	return nil
}

func MarkTransactionBad(dbTransaction *model.DbTransaction, hash []byte, errText string) error {
	if hash == nil {
		return nil
	}
	if len(errText) > 255 {
		errText = errText[:255] + "..."
	}
	log.WithFields(log.Fields{"type": consts.BadTxError, "tx_hash": hash, "error": errText}).Debug("tx marked as bad")

	return model.NewDbTransaction(model.DBConn).Connection().Transaction(func(tx *gorm.DB) error {
		// looks like there is not hash in queue_tx in this moment
		qtx := &model.QueueTx{}
		_, err := qtx.GetByHash(model.NewDbTransaction(tx), hash)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Debug("getting tx by hash from queue")
			return err
		}

		if qtx.FromGate == 0 {
			m := &model.TransactionStatus{}
			err = m.SetError(model.NewDbTransaction(tx), errText, hash)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Debug("setting transaction status error")
				return err
			}
		}
		err = DeleteQueueTx(model.NewDbTransaction(tx), hash)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Debug("deleting transaction from queue")
			return err
		}
		return nil
	})
}

// ProcessQueueTransaction writes transactions into the queue
//func ProcessQueueTransaction(dbTransaction *model.DbTransaction, hash, binaryTx []byte, myTx bool) error {
//	t, err := UnmarshallTransaction(bytes.NewBuffer(binaryTx), true)
//	if err != nil {
//		return err
//	}
//
//	if err = t.Check(time.Now().Unix(), true); err != nil {
//		if err != ErrEarlyTime {
//			return err
//		}
//		return nil
//	}
//
//	if t.TxKeyID == 0 {
//		errStr := "undefined keyID"
//	if found {
//		err = model.DeleteTransactionByHash(dbTransaction, hash)
//		if err != nil {
//			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transaction by hash")
//			return utils.ErrInfo(err)
//		}
//	}
//	// put with verified=1
//	var expedite decimal.Decimal
//	if len(t.TxSmart.Expedite) > 0 {
//		expedite, err = decimal.NewFromString(t.TxSmart.Expedite)
//		if err != nil {
//			return utils.ErrInfo(err)
//		}
//	}
//	newTx := &model.Transaction{
//		Hash:     hash,
//		Data:     binaryTx,
//		Type:     int8(t.TxType),
//		KeyID:    t.TxKeyID,
//		Expedite: expedite,
//		Time:     t.TxTime,
//		Verified: 1,
//	}
//	err = newTx.Create()
//	if err != nil {
//		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating new transaction")
//		return utils.ErrInfo(err)
//	}
//
//	delQueueTx := &model.QueueTx{Hash: hash}
//	if err = delQueueTx.DeleteTx(dbTransaction); err != nil {
//		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transaction from queue")
//		return utils.ErrInfo(err)
//	}
//
//	return nil
//}

// ProcessTransactionsQueue parses new transactions
func ProcessTransactionsQueue(dbTransaction *model.DbTransaction) error {
	all, err := model.GetAllUnverifiedAndUnusedTransactions(dbTransaction, syspar.GetMaxTxCount())
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all unverified and unused transactions")
		return err
	}
	//for i := 0; i < len(all); i++ {
	//	err := ProcessQueueTransaction(dbTransaction, all[i].Hash, all[i].Data, false)
	//	if err != nil {
	//		MarkTransactionBad(dbTransaction, all[i].Hash, err.Error())
	//		return utils.ErrInfo(err)
	//	}
	//	log.Debug("transaction parsed successfully")
	//}
	return ProcessQueueTransactionBatches(dbTransaction, all)
}

// AllTxParser parses new transactions
func ProcessTransactionsAttempt(dbTransaction *model.DbTransaction) error {
	all, err := model.FindTxAttemptCount(dbTransaction, consts.MaxTXAttempt)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all  transactions attempt > consts.MaxTXAttempt")
		return err
	}
	for _, data := range all {
		err := MarkTransactionBad(dbTransaction, data.Hash, fmt.Sprintf("The limit of %d attempts has been reached", consts.MaxTXAttempt))
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("transaction attempt deal successfully")
	}
	return nil
}
