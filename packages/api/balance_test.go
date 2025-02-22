/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/IBAX-io/go-ibax/packages/model"
)

func TestBalance(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var ret balanceResult
	err := sendGet(`balance/`+gAddress, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ret.Amount) < 10 {
		t.Error(`too low balance`, ret)
	}
	err = sendGet(`balance/3434341`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ret.Amount) > 0 {
		t.Error(fmt.Errorf(`wrong balance %s`, ret.Amount))
		return
	}
}

func TestAssignBalance(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var ret model.Response
	err := sendGet(`assignbalance/`+gAddress, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}

	data, _ := json.Marshal(ret)
	fmt.Println(string(data))
	//err = sendGet(`assignbalance/3434341`, nil, &ret)
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//if len(ret.Amount) > 0 {
	//	t.Error(fmt.Errorf(`wrong balance %s`, ret.Amount))
	//	return
	//}
}

func TestMoneyMoreSend(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	//	time.Sleep(2 * time.Second)
	//}
	//for i := 0; i < 2; i++ {
	//	form := url.Values{`Amount`: {`-1`}, `Recipient`: {`1088-3972-0775-1704-9008`}, `Comment`: {`Test`}}
	//	if err := postTx(`TokensSend`, &form); err != nil {
	//		t.Error(err)
	//		return
	//	}
	//	time.Sleep(2 * time.Second)
	//}

	form := url.Values{`Amount`: {`-1`}, `Account`: {`0323-3625-0280-2110-5478`}, `Type`: {`1`}}
	if err := postTx(`AddAssignMember`, &form); err != nil {
		t.Error(err)
		return
	}

}
