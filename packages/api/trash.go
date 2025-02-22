/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package api

import (
	"net/http"

	"github.com/IBAX-io/go-ibax/packages/script"
	"github.com/IBAX-io/go-ibax/packages/smart"
)

func getContract(r *http.Request, name string) *smart.Contract {
	vm := smart.GetVM()
	if vm == nil {
		return nil
	}
	client := getClient(r)
	contract := smart.VMGetContract(vm, name, uint32(client.EcosystemID))
