/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package api

import (
	"net/http"
	"strconv"

	"github.com/IBAX-io/go-ibax/packages/conf"
	"github.com/IBAX-io/go-ibax/packages/conf/syspar"
	"github.com/IBAX-io/go-ibax/packages/converter"
	"github.com/IBAX-io/go-ibax/packages/crypto"
)

type HonorNodeJSON struct {
	TCPAddress string `json:"tcp_address"`
	APIAddress string `json:"api_address"`
	PublicKey  string `json:"public_key"`
	UnbanTime  string `json:"unban_time"`
	Stopped    bool   `json:"stopped"`
			TCPAddress: node.TCPAddress,
			APIAddress: node.APIAddress,
			PublicKey:  crypto.PubToHex(node.PublicKey),
			UnbanTime:  strconv.FormatInt(node.UnbanTime.Unix(), 10),
		})
	}
	return nodes
}

func getNetworkHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, &NetworkResult{
		NetworkID:     converter.Int64ToStr(conf.Config.NetworkID),
		CentrifugoURL: conf.Config.Centrifugo.URL,
		Test:          syspar.IsTestMode(),
		Private:       syspar.IsPrivateBlockchain(),
		HonorNodes:    GetNodesJSON(),
	})
}
