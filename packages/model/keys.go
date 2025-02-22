/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package model

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/IBAX-io/go-ibax/packages/converter"
)

// Key is model
type Key struct {
	ecosystem    int64
	accountKeyID int64 `gorm:"-"`

	ID          int64  `gorm:"primary_key;not null"`
	AccountID   string `gorm:"column:account;not null"`
	PublicKey   []byte `gorm:"column:pub;not null"`
	Amount      string `gorm:"not null"`
	Mintsurplus string `gorm:"not null"`
	Maxpay      string `gorm:"not null"`
	Deleted     int64  `gorm:"not null"`
	Blocked     int64  `gorm:"not null"`
}

// SetTablePrefix is setting table prefix
func (m *Key) SetTablePrefix(prefix int64) *Key {
	m.ecosystem = prefix
	return m
}

// TableName returns name of table
func (m Key) TableName() string {
	if m.ecosystem == 0 {
		m.ecosystem = 1
	}
	return `1_keys`
}
func (m *Key) Disable() bool {
	return m.Deleted != 0 || m.Blocked != 0
}
func (m *Key) CapableAmount() decimal.Decimal {
	amount := decimal.New(0, 0)
	if len(m.Amount) > 0 {
		amount, _ = decimal.NewFromString(m.Amount)
	}
	maxpay := decimal.New(0, 0)
	if len(m.Maxpay) > 0 {
		maxpay, _ = decimal.NewFromString(m.Maxpay)
	}
	if maxpay.GreaterThan(decimal.New(0, 0)) && maxpay.LessThan(amount) {
		amount = maxpay
	}
	return amount
}

// Get is retrieving model from database
func (m *Key) Get(db *DbTransaction, wallet int64) (bool, error) {
	return isFound(GetDB(db).Where("id = ? and ecosystem = ?", wallet, m.ecosystem).First(m))
}

// GetTr is retrieving model from database
func (m *Key) GetTr(db *DbTransaction, wallet int64) (bool, error) {
	return isFound(GetDB(db).Where("id = ? and ecosystem = ?", wallet, m.ecosystem).First(m))
}

func (m *Key) AccountKeyID() int64 {
	if m.accountKeyID == 0 {
		m.accountKeyID = converter.StringToAddress(m.AccountID)
	}
	return m.accountKeyID
}

// KeyTableName returns name of key table
