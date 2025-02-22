/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
package smart

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/IBAX-io/go-ibax/packages/script"
)

type TestSmart struct {
	Input  string
	Output string
}

func TestNewContract(t *testing.T) {
	test := []TestSmart{
		{`contract NewCitizen {
			data {
//				DBInsert(Sprintf( "%d_citizens", $ecosystem_id), "public_key,block_id", $PublicKey, $block)
			}
}			
		`, ``},
	}
	owner := script.OwnerInfo{
		StateID:  1,
		Active:   false,
		TableID:  1,
		WalletID: 0,
		TokenID:  0,
	}
	for _, item := range test {
		if err := Compile(item.Input, &owner); err != nil {
			t.Error(err)
		}
	}
	cnt := GetContract(`NewCitizen`, 1)
	cfunc := cnt.GetFunc(`conditions`)
	_, err := Run(cfunc, nil, &map[string]interface{}{})
	if err != nil {
		t.Error(err)
	}
}

func TestCheckAppend(t *testing.T) {
	appendTestContract := `contract AppendTest {
		action {
			var list array
			list = Append(list, "naw_value")
			Println(list)
		}
	}`

	owner := script.OwnerInfo{
		StateID:  1,
		Active:   false,
		TableID:  1,
		WalletID: 0,
		TokenID:  0,
	}

	require.NoError(t, Compile(appendTestContract, &owner))

	cnt := GetContract("AppendTest", 1)
	cfunc := cnt.GetFunc("action")

	_, err := Run(cfunc, nil, &map[string]interface{}{})
	require.NoError(t, err)
}
