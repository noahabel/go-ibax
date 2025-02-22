/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/IBAX-io/go-ibax/packages/consts"
	"github.com/IBAX-io/go-ibax/packages/converter"
	"github.com/IBAX-io/go-ibax/packages/model"

	log "github.com/sirupsen/logrus"
)

type txstatusError struct {
	Type  string `json:"type,omitempty"`
	Error string `json:"error,omitempty"`
	Id    string `json:"id,omitempty"`
}

type txstatusResult struct {
	BlockID string         `json:"blockid"`
	Message *txstatusError `json:"errmsg,omitempty"`
	Result  string         `json:"result"`
	Penalty int64          `json:"penalty"`
}

func getTxStatus(r *http.Request, hash string) (*txstatusResult, error) {
	logger := getLogger(r)

	var status txstatusResult
	if _, err := hex.DecodeString(hash); err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding tx hash from hex")
				status.Message = &txstatusError{
					Type:  "txError",
					Error: ts.Error,
				}
			}
		}
	}
	if ts.BlockID > 0 {
		status.BlockID = converter.Int64ToStr(ts.BlockID)
		status.Penalty = ts.Penalty
		if ts.Penalty == 1 {
			checkErr()
		} else {
			status.Result = ts.Error
		}
	} else {
		checkErr()
	}
	return &status, nil
}

type multiTxStatusResult struct {
	Results map[string]*txstatusResult `json:"results"`
}

type txstatusRequest struct {
	Hashes []string `json:"hashes"`
}

func getTxStatusHandler(w http.ResponseWriter, r *http.Request) {
	result := &multiTxStatusResult{}
	result.Results = map[string]*txstatusResult{}

	var request txstatusRequest
	if err := json.Unmarshal([]byte(r.FormValue("data")), &request); err != nil {
		errorResponse(w, errHashWrong)
		return
	}
	for _, hash := range request.Hashes {
		status, err := getTxStatus(r, hash)
		if err != nil {
			errorResponse(w, err)
			return
		}
		result.Results[hash] = status
	}

	jsonResponse(w, result)
}
