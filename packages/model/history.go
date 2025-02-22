/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package model

import (
	"github.com/IBAX-io/go-ibax/packages/consts"

	"github.com/shopspring/decimal"
)

// History represent record of history table
type History struct {
	ecosystem        int64
	ID               int64
	SenderID         int64
	RecipientID      int64
	SenderBalance    decimal.Decimal
	RecipientBalance decimal.Decimal
	Amount           decimal.Decimal
	Comment          string `json:"comment,omitempty"`
	BlockID          int64  `json:"block_id,omitempty"`
	TxHash           []byte `gorm:"column:txhash"`
	CreatedAt        int64  `json:"created_at,omitempty"`
	Type             int64
}

// SetTablePrefix is setting table prefix
func (h *History) SetTablePrefix(prefix int64) *History {
	h.ecosystem = prefix
	return h
}

// TableName returns table name
func (h *History) TableName() string {
	if h.ecosystem == 0 {
		h.ecosystem = 1
	}
	return `1_history`
}

// MoneyTransfer from to amount
type MoneyTransfer struct {
	SenderID    int64
	RecipientID int64
	Amount      decimal.Decimal
}

//SenderTxCount struct to scan query result
type SenderTxCount struct {
	SenderID int64
	TxCount  int64
}

// Get is retrieving model from database
		Amount decimal.Decimal
	}

	var res result
	err = db.Table("1_history").Select("SUM(amount) as amount").
		Where("to_timestamp(created_at) > NOW() - interval '24 hours' AND amount > 0").Scan(&res).Error

	return res.Amount, err
}

// GetExcessFromToTokenMovementPerDay returns from to pairs where sum of amount greather than fromToPerDayLimit per 24 hours
func GetExcessFromToTokenMovementPerDay(tx *DbTransaction) (excess []MoneyTransfer, err error) {
	db := GetDB(tx)
	err = db.Table("1_history").
		Select("sender_id, recipient_id, SUM(amount) amount").
		Where("to_timestamp(created_at) > NOW() - interval '24 hours' AND amount > 0").
		Group("sender_id, recipient_id").
		Having("SUM(amount) > ?", consts.FromToPerDayLimit).
		Scan(&excess).Error

	return excess, err
}

// GetExcessTokenMovementQtyPerBlock returns from to pairs where money transactions count greather than tokenMovementQtyPerBlockLimit per 24 hours
func GetExcessTokenMovementQtyPerBlock(tx *DbTransaction, blockID int64) (excess []SenderTxCount, err error) {
	db := GetDB(tx)
	err = db.Table("1_history").
		Select("sender_id, count(*) tx_count").
		Where("block_id = ? AND amount > ?", blockID, 0).
		Group("sender_id").
		Having("count(*) > ?", consts.TokenMovementQtyPerBlockLimit).
		Scan(&excess).Error

	return excess, err
}

func GetWalletRecordHistory(tx *DbTransaction, keyId string, searchType string, limit, offset int) (histories []History, err error) {
	db := GetDB(tx)
	if searchType == "income" {
		err = db.Table("1_history").
			Where("recipient_id = ?", keyId).
			Order("id desc").
			Limit(limit).
			Offset(offset).
			Scan(&histories).Error
	} else if searchType == "outcome" {
		err = db.Table("1_history").
			Where("sender_id = ?", keyId).
			Order("id desc").
			Limit(limit).
			Offset(offset).
			Scan(&histories).Error
	} else {
		err = db.Table("1_history").
			Where("recipient_id = ? OR sender_id = ?", keyId, keyId).
			Order("id desc").
			Limit(limit).
			Offset(offset).
			Scan(&histories).Error
	}
	return histories, err

}
