/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package daemons

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	vde_api "github.com/IBAX-io/go-ibax/packages/vde_sdk"

	"path/filepath"

	"github.com/IBAX-io/go-ibax/packages/conf"
	"github.com/IBAX-io/go-ibax/packages/model"

	log "github.com/sirupsen/logrus"
)

//Install task contract
func VDEDestTaskInstallContractDest(ctx context.Context, d *daemon) error {
	var (
		blockchain_http      string
		blockchain_ecosystem string
		err                  error
	)

	m := &model.VDEDestTaskFromSrc{}
	DestTask, err := m.GetAllByContractStateDest(0) //0
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("getting all untreated task data")
		time.Sleep(time.Millisecond * 2)
		return err
	}
	if len(DestTask) == 0 {
		//log.Info("Dest task not found")
		time.Sleep(time.Millisecond * 2)
		return nil
	}

	// deal with task data
	for _, item := range DestTask {
		//fmt.Println("DestTask:", item.TaskUUID)
		blockchain_http = item.ContractRunHttp
		blockchain_ecosystem = item.ContractRunEcosystem
		//fmt.Println("ContractRunHttp and ContractRunEcosystem:", blockchain_http, blockchain_ecosystem)
		ecosystemID, err := strconv.Atoi(blockchain_ecosystem)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("VDEDestTaskInstallContractDest encode error")
			time.Sleep(time.Millisecond * 2)
			continue
		}
		//api.ApiAddress = blockchain_http
		//api.ApiEcosystemID = int64(ecosystemID)
		vde_dest_apiAddress := blockchain_http
		vde_dest_apiEcosystemID := int64(ecosystemID)

		src := filepath.Join(conf.Config.KeysDir, "PrivateKey")
		// Login
		//err := api.KeyLogin(src, api.ApiEcosystemID)
		gAuth_dest, _, gPrivate_dest, _, _, err := vde_api.KeyLogin(vde_dest_apiAddress, src, vde_dest_apiEcosystemID)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Login chain failure")
			time.Sleep(time.Millisecond * 2)
			continue
		}
		//fmt.Println("Login OK!")

		ContractDest := item.ContractDestGet

		form := url.Values{
			`Value`:         {ContractDest},

		item.UpdateTime = time.Now().Unix()
		err = item.Updates()
		if err != nil {
			fmt.Println("Update VDEDestTask table err: ", err)
			log.WithFields(log.Fields{"error": err}).Error("Update VDEDestTask table!")
			time.Sleep(time.Millisecond * 2)
			continue
		}

	} //for

	return nil
}
