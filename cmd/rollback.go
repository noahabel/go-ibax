/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package cmd

import (
	"github.com/IBAX-io/go-ibax/packages/conf"
	"github.com/IBAX-io/go-ibax/packages/conf/syspar"
	"github.com/IBAX-io/go-ibax/packages/model"
	"github.com/IBAX-io/go-ibax/packages/rollback"
	"github.com/IBAX-io/go-ibax/packages/smart"
	"github.com/IBAX-io/go-ibax/packages/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var blockID int64

// rollbackCmd represents the rollback command
var rollbackCmd = &cobra.Command{
	Use:    "rollback",
	Short:  "Rollback blockchain to blockID",
	PreRun: loadConfigWKey,
	Run: func(cmd *cobra.Command, args []string) {
		f := utils.LockOrDie(conf.Config.LockFilePath)
		defer f.Unlock()

		if err := model.GormInit(
			conf.Config.DB.Host,
			conf.Config.DB.Port,
			conf.Config.DB.User,
			conf.Config.DB.Password,
			conf.Config.DB.Name,
		); err != nil {
		if err := syspar.SysTableColType(nil); err != nil {
			log.WithFields(log.Fields{"error": err}).Error("updating sys table col type")
		}

		smart.InitVM()
		if err := smart.LoadContracts(); err != nil {
			log.WithError(err).Fatal("loading contracts")
			return
		}
		err := rollback.ToBlockID(blockID, nil, log.WithFields(log.Fields{}))
		if err != nil {
			log.WithError(err).Fatal("rollback to block id")
			return
		}

		// block id = 1, is a special case for full rollback
		if blockID != 1 {
			log.Info("Not full rollback, finishing work without checking")
			return
		}
	},
}

func init() {
	rollbackCmd.Flags().Int64Var(&blockID, "blockId", 1, "blockID to rollback")
	rollbackCmd.MarkFlagRequired("blockId")
}
