/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package model

import (
	"time"
)

// StopDaemon is model
type StopDaemon struct {
	StopTime int64 `gorm:"not null"`
}

// TableName returns name of table
func (sd *StopDaemon) TableName() string {
	return "stop_daemons"
}

// Create is creating record of model
func (sd *StopDaemon) Create() error {
	return DBConn.Create(sd).Error
}

// Delete is deleting record
func (sd *StopDaemon) Delete() error {
	return DBConn.Delete(&StopDaemon{}).Error
}
