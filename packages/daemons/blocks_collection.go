/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package daemons

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/IBAX-io/go-ibax/packages/smart"

	"github.com/IBAX-io/go-ibax/packages/block"
	"github.com/IBAX-io/go-ibax/packages/network"
	"github.com/IBAX-io/go-ibax/packages/network/tcpclient"

	"github.com/IBAX-io/go-ibax/packages/conf"
	"github.com/IBAX-io/go-ibax/packages/conf/syspar"
	"github.com/IBAX-io/go-ibax/packages/consts"

	"github.com/IBAX-io/go-ibax/packages/model"
	"github.com/IBAX-io/go-ibax/packages/rollback"
	"github.com/IBAX-io/go-ibax/packages/service"
	"github.com/IBAX-io/go-ibax/packages/transaction"
	"github.com/IBAX-io/go-ibax/packages/utils"

	"github.com/IBAX-io/go-ibax/packages/crypto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// BlocksCollection collects and parses blocks
func BlocksCollection(ctx context.Context, d *daemon) error {
	if ctx.Err() != nil {
		d.logger.WithFields(log.Fields{"type": consts.ContextError, "error": ctx.Err()}).Error("context error")
		return ctx.Err()
	}

	return blocksCollection(ctx, d)
}

func blocksCollection(ctx context.Context, d *daemon) (err error) {
	if !atomic.CompareAndSwapUint32(&d.atomic, 0, 1) {
		return nil
	}
	defer atomic.StoreUint32(&d.atomic, 0)

	//if !NtpDriftFlag {
	//	d.logger.WithFields(log.Fields{"type": consts.Ntpdate}).Error("ntp time not ntpdate")
	//	return nil
	//}

	host, maxBlockID, err := getHostWithMaxID(ctx, d.logger)
	if err != nil {
		d.logger.WithFields(log.Fields{"error": err}).Warn("on checking best host")
		return err
	}

	infoBlock := &model.InfoBlock{}
	found, err := infoBlock.Get()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting cur blockID")
		return err
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound, "error": err}).Error("Info block not found")
		return errors.New("Info block not found")
	}

	if infoBlock.BlockID >= maxBlockID {
		log.WithFields(log.Fields{"blockID": infoBlock.BlockID, "maxBlockID": maxBlockID}).Debug("Max block is already in the host")
		return nil
	}

	DBLock()
	defer func() {
		service.NodeDoneUpdatingBlockchain()
		DBUnlock()
	}()

	// update our chain till maxBlockID from the host
	return UpdateChain(ctx, d, host, maxBlockID)
}

// UpdateChain load from host all blocks from our last block to maxBlockID
func UpdateChain(ctx context.Context, d *daemon, host string, maxBlockID int64) error {
	// get current block id from our blockchain
	curBlock := &model.InfoBlock{}
	if _, err := curBlock.Get(); err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting info block")
		return err
	}

	if ctx.Err() != nil {
		d.logger.WithFields(log.Fields{"type": consts.ContextError, "error": ctx.Err()}).Error("context error")
		return ctx.Err()
	}

	playRawBlock := func(rb []byte) error {
		var lastBlockID, lastBlockTime int64
		var err error
		defer func(err2 *error) {
			if err2 != nil {
				banNodePause(host, lastBlockID, lastBlockTime, *err2)
			}
		}(&err)
		bl, err := block.ProcessBlockWherePrevFromBlockchainTable(rb, true)
		if err != nil {
			d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("processing block")
			return err
		}

		curBlock := &model.InfoBlock{}
		if _, err = curBlock.Get(); err != nil {
			d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting info block")
			return err
		}

		if bl.PrevHeader != nil {
			if curBlock.BlockID != bl.PrevHeader.BlockID {
				d.logger.WithFields(log.Fields{"type": consts.DBError}).Error("Getting info block  err curBlock.BlockID: " + strconv.FormatInt(curBlock.BlockID, 10) + "bl.PrevHeader.BlockID: " + strconv.FormatInt(bl.PrevHeader.BlockID, 10))
				return err
			}
		} else {
			d.logger.WithFields(log.Fields{"type": consts.DBError}).Error("Getting info block PrevHeader nil")
			return err
		}

		lastBlockID = bl.Header.BlockID
		lastBlockTime = bl.Header.Time

		if err = bl.Check(); err != nil {
			var replaceCount int64 = 1
			if err == block.ErrIncorrectRollbackHash {
				replaceCount++
			}
			d.logger.WithFields(log.Fields{"error": err, "from_host": host, "different": fmt.Errorf("not match block %d, prev_position %d, current_position %d", bl.PrevHeader.BlockID, bl.PrevHeader.NodePosition, bl.Header.NodePosition), "type": consts.BlockError, "replaceCount": replaceCount}).Error("checking block hash")
			//it should be fork, replace our previous blocks to ones from the host
			if errReplace := ReplaceBlocksFromHost(ctx, host, bl.PrevHeader.BlockID, replaceCount); errReplace != nil {
				return errReplace
			}
			return err
		}
		return bl.PlaySafe()
	}

	var count int
	st := time.Now()

	//if conf.Config.PoolPub.Enable {
	//	bi := model.BlockID{}
	//	f, err := bi.GetRangeByName(consts.MintMax, consts.ChainMax, 2000)
	//	fmt.Println(f, err)
	//	if err != nil {
	//		return err
	//	}
	//	if f {
	//		time.Sleep(4 * time.Second)
	//		return errors.New("deal mint blockid please wait")
	//	}
	//}

	d.logger.WithFields(log.Fields{"min_block": curBlock.BlockID, "max_block": maxBlockID, "count": maxBlockID - curBlock.BlockID}).Info("starting downloading blocks")
	for blockID := curBlock.BlockID + 1; blockID <= maxBlockID; blockID += int64(network.BlocksPerRequest) {

		if loopErr := func() error {
			ctxDone, cancel := context.WithCancel(ctx)
			defer func() {
				cancel()
				d.logger.WithFields(log.Fields{"count": count, "time": time.Since(st).String()}).Info("blocks downloaded")
			}()

			rawBlocksChan, err := tcpclient.GetBlocksBodies(ctxDone, host, blockID, false)
			if err != nil {
				d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("getting block body")
				return err
			}

			for rawBlock := range rawBlocksChan {

				//if conf.Config.PoolPub.Enable {
				//	bi := model.BlockID{}
				//	f, err := bi.GetRangeByName(consts.MintMax, consts.ChainMax, 2000)
				//	if err != nil {
				//		return err
				//	}
				//	if !f {
				//		if err = playRawBlock(rawBlock); err != nil {
				//			d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("playing raw block")
				//			return err
				//		}
				//	} else {
				//		time.Sleep(4 * time.Second)
				//		return errors.New("deal mint blockid please wait")
				//	}
				//} else {

				if err = playRawBlock(rawBlock); err != nil {
					d.logger.WithFields(log.Fields{"error": err, "type": consts.BlockError}).Error("playing raw block")
					return err
				}
				count++
				//}
			}

			return nil
		}(); loopErr != nil {
			return loopErr
		}
	}
	return nil
}

func banNodePause(host string, blockID, blockTime int64, err error) {
	if err == nil || !utils.IsBanError(err) {
		return
	}

	reason := err.Error()
	}
}

// GetHostWithMaxID returns host with maxBlockID
func getHostWithMaxID(ctx context.Context, logger *log.Entry) (host string, maxBlockID int64, err error) {

	nbs := service.GetNodesBanService()
	hosts, err := nbs.FilterBannedHosts(syspar.GetRemoteHosts())
	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Error("on filtering banned hosts")
	}

	host, maxBlockID, err = tcpclient.HostWithMaxBlock(ctx, hosts)
	if len(hosts) == 0 || err == tcpclient.ErrNodesUnavailable {
		hosts = conf.GetNodesAddr()
		return tcpclient.HostWithMaxBlock(ctx, hosts)
	}

	return
}

// ReplaceBlocksFromHost replaces blockchain received from the host.
// Number (replaceCount) of blocks starting from blockID will be re-played.
func ReplaceBlocksFromHost(ctx context.Context, host string, blockID, replaceCount int64) error {

	blocks, err := getBlocks(ctx, host, blockID, replaceCount)
	if err != nil {
		return err
	}
	transaction.CleanCache()

	// mark all transaction as unverified
	_, err = model.MarkVerifiedAndNotUsedTransactionsUnverified()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"type":  consts.DBError,
		}).Error("marking verified and not used transactions unverified")
		return utils.ErrInfo(err)
	}

	// get starting blockID from slice of blocks
	if len(blocks) > 0 {
		blockID = blocks[len(blocks)-1].Header.BlockID
	}

	// we have the slice of blocks for applying
	// first of all we should rollback old blocks
	b := &model.Block{}
	myRollbackBlocks, err := b.GetBlocksFrom(blockID-1, "desc", 0)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting rollback blocks from blockID")
		return utils.ErrInfo(err)
	}
	for _, b := range myRollbackBlocks {
		err := rollback.RollbackBlock(b.Data)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}

	smart.SavepointSmartVMObjects()
	err = processBlocks(blocks)
	if err != nil {
		smart.RollbackSmartVMObjects()
		return err
	}
	smart.ReleaseSmartVMObjects()
	return err
}

func getBlocks(ctx context.Context, host string, blockID, minCount int64) ([]*block.Block, error) {
	rollback := syspar.GetRbBlocks1()
	blocks := make([]*block.Block, 0)
	nextBlockID := blockID

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// load the block bodies from the host
	blocksCh, err := tcpclient.GetBlocksBodies(ctx, host, blockID, true)
	if err != nil {
		return nil, utils.WithBan(errors.Wrapf(err, "Getting bodies of blocks by id %d", blockID))
	}

	for binaryBlock := range blocksCh {
		if blockID < 2 {
			break
		}

		// if the limit of blocks received from the node was exaggerated
		if len(blocks) >= int(rollback) {
			break
		}

		bl, err := block.ProcessBlockWherePrevFromBlockchainTable(binaryBlock, true)
		if err != nil {
			return nil, err
		}

		if bl.Header.BlockID != nextBlockID {
			log.WithFields(log.Fields{"header_block_id": bl.Header.BlockID, "block_id": blockID, "type": consts.InvalidObject}).Error("block ids does not match")
			return nil, utils.WithBan(errors.New("bad block_data['block_id']"))
		}

		// the public key of the one who has generated this block
		nodePublicKey, err := syspar.GetNodePublicKeyByPosition(bl.Header.NodePosition)
		if err != nil {
			log.WithFields(log.Fields{"header_block_id": bl.Header.BlockID, "block_id": blockID, "type": consts.InvalidObject}).Error("block ids does not match")
			return nil, utils.ErrInfo(err)
		}

		// save the block
		blocks = append(blocks, bl)

		// check the signature
		_, okSignErr := utils.CheckSign([][]byte{nodePublicKey},
			[]byte(bl.Header.ForSign(bl.PrevHeader, bl.MrklRoot)),
			bl.Header.Sign, true)
		if okSignErr == nil && len(blocks) >= int(minCount) {
			break
		}

		nextBlockID--
	}

	return blocks, nil
}

func processBlocks(blocks []*block.Block) error {
	dbTransaction, err := model.StartTransaction()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("starting transaction")
		return utils.ErrInfo(err)
	}

	// go through new blocks from the smallest block_id to the largest block_id
	prevBlocks := make(map[int64]*block.Block, 0)

	for i := len(blocks) - 1; i >= 0; i-- {
		b := blocks[i]

		if prevBlocks[b.Header.BlockID-1] != nil {
			b.PrevHeader.Hash = prevBlocks[b.Header.BlockID-1].Header.Hash
			b.PrevHeader.RollbacksHash = prevBlocks[b.Header.BlockID-1].Header.RollbacksHash
			b.PrevHeader.Time = prevBlocks[b.Header.BlockID-1].Header.Time
			b.PrevHeader.BlockID = prevBlocks[b.Header.BlockID-1].Header.BlockID
			b.PrevHeader.EcosystemID = prevBlocks[b.Header.BlockID-1].Header.EcosystemID
			b.PrevHeader.KeyID = prevBlocks[b.Header.BlockID-1].Header.KeyID
			b.PrevHeader.NodePosition = prevBlocks[b.Header.BlockID-1].Header.NodePosition
		}

		b.Header.Hash = crypto.DoubleHash([]byte(b.Header.ForSha(b.PrevHeader, b.MrklRoot)))

		if err := b.Check(); err != nil {
			dbTransaction.Rollback()
			return err
		}

		if err := b.Play(dbTransaction); err != nil {
			dbTransaction.Rollback()
			return utils.ErrInfo(err)
		}
		prevBlocks[b.Header.BlockID] = b

		// for last block we should update block info
		if i == 0 {
			err := block.UpdBlockInfo(dbTransaction, b)
			if err != nil {
				dbTransaction.Rollback()
				return utils.ErrInfo(err)
			}
		}
		if b.SysUpdate {
			if err := syspar.SysUpdate(dbTransaction); err != nil {
				log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
				return utils.ErrInfo(err)
			}
		}
	}

	// If all right we can delete old blockchain and write new
	for i := len(blocks) - 1; i >= 0; i-- {
		b := blocks[i]
		// insert new blocks into blockchain
		if err := block.InsertIntoBlockchain(dbTransaction, b); err != nil {
			dbTransaction.Rollback()
			return err
		}
	}

	return dbTransaction.Commit()
}
