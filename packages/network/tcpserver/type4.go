/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
package tcpserver

import (
	"github.com/IBAX-io/go-ibax/packages/consts"
	"github.com/IBAX-io/go-ibax/packages/model"
	"github.com/IBAX-io/go-ibax/packages/network"

	log "github.com/sirupsen/logrus"
)

// Type4 writes the hash of the specified block
// The request is sent by 'confirmations' daemon
func Type4(r *network.ConfirmRequest) (*network.ConfirmResponse, error) {
	resp := &network.ConfirmResponse{}
	block := &model.Block{}
	found, err := block.Get(int64(r.BlockID))
	if err != nil || !found {
		hash := [32]byte{}
		resp.Hash = hash[:]
	} else {
		resp.Hash = block.Hash // can we send binary data ?
	}
	if err != nil {
