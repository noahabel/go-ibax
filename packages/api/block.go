/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package api

import (
	"bytes"
	"net/http"

	"github.com/IBAX-io/go-ibax/packages/block"
	"github.com/IBAX-io/go-ibax/packages/consts"
	"github.com/IBAX-io/go-ibax/packages/converter"
	"github.com/IBAX-io/go-ibax/packages/model"

	"errors"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type maxBlockResult struct {
	MaxBlockID int64 `json:"max_block_id"`
}

func getMaxBlockHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)

	block := &model.Block{}
	found, err := block.GetMaxBlock()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting max block")
		errorResponse(w, err)
		return
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound}).Error("last block not found")
		errorResponse(w, errNotFound)
		return
	}

	jsonResponse(w, &maxBlockResult{block.ID})
}

type blockInfoResult struct {
	Hash          []byte `json:"hash"`
	EcosystemID   int64  `json:"ecosystem_id"`
	KeyID         int64  `json:"key_id"`
	Time          int64  `json:"time"`
	Tx            int32  `json:"tx_count"`
	RollbacksHash []byte `json:"rollbacks_hash"`
	NodePosition  int64  `json:"node_position"`
}

func getBlockInfoHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)
	params := mux.Vars(r)

	blockID := converter.StrToInt64(params["id"])
	block := model.Block{}
	found, err := block.Get(blockID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block")
		errorResponse(w, err)
		return
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "id": blockID}).Error("block with id not found")
		errorResponse(w, errNotFound)
		return
	}

	jsonResponse(w, &blockInfoResult{
		Hash:          block.Hash,
		EcosystemID:   block.EcosystemID,
		KeyID:         block.KeyID,
		Time:          block.Time,
		Tx:            block.Tx,
		RollbacksHash: block.RollbacksHash,
		NodePosition:  block.NodePosition,
	})
}

type TxInfo struct {
	Hash         []byte                 `json:"hash"`
	ContractName string                 `json:"contract_name"`
	Params       map[string]interface{} `json:"params"`
	KeyID        int64                  `json:"key_id"`
}

type blocksTxInfoForm struct {
	BlockID int64 `schema:"block_id"`
	Count   int64 `schema:"count"`
}

func (f *blocksTxInfoForm) Validate(r *http.Request) error {
	if f.BlockID > 0 {
		f.BlockID--
	}
	return nil
}

func getBlocksTxInfoHandler(w http.ResponseWriter, r *http.Request) {
	form := &blocksTxInfoForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	if form.BlockID < 0 || form.Count < 0 {
		err := errors.New("parameter is invalid")
		errorResponse(w, err, http.StatusBadRequest)
		return
	}
	logger := getLogger(r)

	blocks, err := model.GetBlockchain(form.BlockID, form.BlockID+form.Count, model.OrderASC)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting blocks range")
			logger.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err, "bolck_id": blockModel.ID}).Error("on unmarshalling block")
			errorResponse(w, err)
			return
		}

		txInfoCollection := make([]TxInfo, 0, len(blck.Transactions))
		for _, tx := range blck.Transactions {
			txInfo := TxInfo{
				Hash: tx.TxHash,
			}

			if tx.TxContract != nil {
				txInfo.ContractName = tx.TxContract.Name
				txInfo.Params = tx.TxData
			}

			if blck.IsGenesis() {
				txInfo.KeyID = blck.Header.KeyID
			} else {
				txInfo.KeyID = tx.TxHeader.KeyID
			}

			txInfoCollection = append(txInfoCollection, txInfo)

			logger.WithFields(log.Fields{"block_id": blockModel.ID, "tx hash": txInfo.Hash, "contract_name": txInfo.ContractName, "key_id": txInfo.KeyID, "params": txInfoCollection}).Debug("Block Transactions Information")
		}

		result[blockModel.ID] = txInfoCollection
	}

	jsonResponse(w, &result)
}

type TxDetailedInfo struct {
	Hash         []byte                 `json:"hash"`
	ContractName string                 `json:"contract_name"`
	Params       map[string]interface{} `json:"params"`
	KeyID        int64                  `json:"key_id"`
	Time         int64                  `json:"time"`
	Type         int64                  `json:"type"`
}

type BlockHeaderInfo struct {
	BlockID      int64  `json:"block_id"`
	Time         int64  `json:"time"`
	EcosystemID  int64  `json:"-"`
	KeyID        int64  `json:"key_id"`
	NodePosition int64  `json:"node_position"`
	Sign         []byte `json:"-"`
	Hash         []byte `json:"-"`
	Version      int    `json:"version"`
}

type BlockDetailedInfo struct {
	Header        BlockHeaderInfo  `json:"header"`
	Hash          []byte           `json:"hash"`
	EcosystemID   int64            `json:"-"`
	NodePosition  int64            `json:"node_position"`
	KeyID         int64            `json:"key_id"`
	Time          int64            `json:"time"`
	Tx            int32            `json:"tx_count"`
	RollbacksHash []byte           `json:"rollbacks_hash"`
	MrklRoot      []byte           `json:"mrkl_root"`
	BinData       []byte           `json:"bin_data"`
	SysUpdate     bool             `json:"-"`
	GenBlock      bool             `json:"-"`
	StopCount     int              `json:"stop_count"`
	Transactions  []TxDetailedInfo `json:"transactions"`
}

func getBlocksDetailedInfoHandler(w http.ResponseWriter, r *http.Request) {
	form := &blocksTxInfoForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}
	if form.BlockID < 0 || form.Count < 0 {
		err := errors.New("parameter is invalid")
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	logger := getLogger(r)

	blocks, err := model.GetBlockchain(form.BlockID, form.BlockID+form.Count, model.OrderASC)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting blocks range")
		errorResponse(w, err)
		return
	}

	if len(blocks) == 0 {
		errorResponse(w, errNotFound)
		return
	}

	result := map[int64]BlockDetailedInfo{}
	for _, blockModel := range blocks {
		blck, err := block.UnmarshallBlock(bytes.NewBuffer(blockModel.Data), false)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err, "bolck_id": blockModel.ID}).Error("on unmarshalling block")
			errorResponse(w, err)
			return
		}

		txDetailedInfoCollection := make([]TxDetailedInfo, 0, len(blck.Transactions))
		for _, tx := range blck.Transactions {
			txDetailedInfo := TxDetailedInfo{
				Hash: tx.TxHash,
			}

			if tx.TxContract != nil {
				txDetailedInfo.ContractName = tx.TxContract.Name
				txDetailedInfo.Params = tx.TxData
				txDetailedInfo.KeyID = tx.TxKeyID
				txDetailedInfo.Time = tx.TxTime
				txDetailedInfo.Type = tx.TxType
			}

			txDetailedInfoCollection = append(txDetailedInfoCollection, txDetailedInfo)

			logger.WithFields(log.Fields{"block_id": blockModel.ID, "tx hash": txDetailedInfo.Hash, "contract_name": txDetailedInfo.ContractName, "key_id": txDetailedInfo.KeyID, "time": txDetailedInfo.Time, "type": txDetailedInfo.Type, "params": txDetailedInfoCollection}).Debug("Block Transactions Information")
		}

		header := BlockHeaderInfo{
			BlockID:      blck.Header.BlockID,
			Time:         blck.Header.Time,
			EcosystemID:  blck.Header.EcosystemID,
			KeyID:        blck.Header.KeyID,
			NodePosition: blck.Header.NodePosition,
			Sign:         blck.Header.Sign,
			Hash:         blck.Header.Hash,
			Version:      blck.Header.Version,
		}

		bdi := BlockDetailedInfo{
			Header:        header,
			Hash:          blockModel.Hash,
			EcosystemID:   blockModel.EcosystemID,
			NodePosition:  blockModel.NodePosition,
			KeyID:         blockModel.KeyID,
			Time:          blockModel.Time,
			Tx:            blockModel.Tx,
			RollbacksHash: blockModel.RollbacksHash,
			MrklRoot:      blck.MrklRoot,
			BinData:       blck.BinData,
			SysUpdate:     blck.SysUpdate,
			GenBlock:      blck.GenBlock,
			Transactions:  txDetailedInfoCollection,
		}
		result[blockModel.ID] = bdi
	}

	jsonResponse(w, &result)
}
