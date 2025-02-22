/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package daemons

import (
	"encoding/hex"

	callDelayedContract = "CallDelayedContract"
	firstEcosystemID    = 1
)

// DelayedTx represents struct which works with delayed contracts
type DelayedTx struct {
	logger     *log.Entry
	privateKey string
	publicKey  string
	time       int64
}

// RunForDelayBlockID creates the transactions that need to be run for blockID
func (dtx *DelayedTx) RunForDelayBlockID(blockID int64) ([]*model.Transaction, error) {

	contracts, err := model.GetAllDelayedContractsForBlockID(blockID)
	if err != nil {
		dtx.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting delayed contracts for block")
		return nil, err
	}
	txList := make([]*model.Transaction, 0, len(contracts))
	for _, c := range contracts {
		params := make(map[string]interface{})
		params["Id"] = c.ID
		tx, err := dtx.createDelayTx(c.KeyID, c.HighRate, params)
		if err != nil {
			dtx.logger.WithFields(log.Fields{"error": err}).Debug("can't create transaction for delayed contract")
			return nil, err
		}
		txList = append(txList, tx)
	}

	return txList, nil
}

func (dtx *DelayedTx) createDelayTx(keyID, highRate int64, params map[string]interface{}) (*model.Transaction, error) {
	vm := smart.GetVM()
	contract := smart.VMGetContract(vm, callDelayedContract, uint32(firstEcosystemID))
	info := contract.Info()

	smartTx := tx.SmartContract{
		Header: tx.Header{
			ID:          int(info.ID),
			Time:        dtx.time,
			EcosystemID: firstEcosystemID,
			KeyID:       keyID,
			NetworkID:   conf.Config.NetworkID,
		},
		SignedBy: smart.PubToID(dtx.publicKey),
		Params:   params,
	}

	privateKey, err := hex.DecodeString(dtx.privateKey)
	if err != nil {
		return nil, err
	}

	txData, txHash, err := tx.NewInternalTransaction(smartTx, privateKey)
	if err != nil {
		return nil, err
	}
	return tx.CreateDelayTransactionHighRate(txData, txHash, keyID, highRate), nil
}
