/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package daemons

import (
	"context"
	"sync"
	"time"

	"github.com/IBAX-io/go-ibax/packages/consts"
	"github.com/IBAX-io/go-ibax/packages/model"
	"github.com/IBAX-io/go-ibax/packages/transaction"

	log "github.com/sirupsen/logrus"
)

var mutex = sync.Mutex{}

// WaitDB waits for the end of the installation
func WaitDB(ctx context.Context) error {
	// There is could be the situation when installation is not over yet.
	// Database could be created but tables are not inserted yet

	if model.DBConn != nil && CheckDB() {
		return nil
	}

	// poll a base with period
	tick := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-tick.C:
			if model.DBConn != nil && CheckDB() {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// CheckDB check if installation complete or not
func CheckDB() bool {
	install := &model.Install{}

	err := install.Get()
func DBLock() {
	mutex.Lock()
}

// DBUnlock unlocks database
func DBUnlock() {
	transaction.CleanCache()
	mutex.Unlock()
}
