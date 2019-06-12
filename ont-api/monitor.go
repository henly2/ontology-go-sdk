package ont_api

import (
	"context"
	"fmt"
	"github.com/alecthomas/log4go"
	"github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/core/types"
	"sync"
	"time"
)

type RetryFuncImpl func() (interface{}, error)

func RetryFunc(cnt int, f RetryFuncImpl) (interface{}, error) {
	errCnt := 0

	for {
		ack, err := f()
		if err != nil {
			errCnt++
			if errCnt >= cnt {
				return nil, err
			}

			continue
		}

		return ack, nil
	}

	return nil, fmt.Errorf("RetryFunc unknown error")
}

type (
	Monitor struct {
		ontSdk *ontology_go_sdk.OntologySdk
		logger log4go.Logger

		wg sync.WaitGroup

		mux   sync.Mutex
		start bool

		startBlock uint32
		data       chan *BlockInfo

		dataHeight chan uint32
	}
)

func NewMonitor(logger log4go.Logger, ontSdk *ontology_go_sdk.OntologySdk, startBlock uint32) *Monitor {
	m := &Monitor{
		ontSdk:     ontSdk,
		logger:     logger,
		start:      false,
		startBlock: startBlock,
		data:       make(chan *BlockInfo, 256),
		dataHeight: make(chan uint32),
	}

	if m.logger == nil {
		m.logger = log4go.Global
	}

	return m
}

func (m *Monitor) Start(ctx context.Context) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.start {
		return
	}

	m.wg.Add(1)
	go m.loop(ctx)

	m.start = true
}

func (m *Monitor) Stop() {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.start == false {
		return
	}

	m.wg.Wait()
	m.start = false
}

func (m *Monitor) Data() chan *BlockInfo {
	return m.data
}

func (m *Monitor) DataHeight() chan uint32 {
	return m.dataHeight
}

func (m *Monitor) retryGetBlock(index uint32) (*types.Block, error) {
	ack, err := RetryFunc(2, func() (interface{}, error) {
		return m.ontSdk.GetBlockByHeight(index)
	})
	if err != nil {
		return nil, err
	}

	data, ok := ack.(*types.Block)
	if !ok {
		return nil, fmt.Errorf("data type not *readable.Block")
	}

	return data, nil
}

func (m *Monitor) loop(ctx context.Context) {
	m.logger.Info("http timer monitor begin...")
	defer m.wg.Done()
	defer close(m.data)
	defer close(m.dataHeight)

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	if m.startBlock != 0 {
		m.startBlock = m.startBlock - 1
	}

	for {
		select {
		case <-ticker.C:
			curHeight, err := m.ontSdk.GetCurrentBlockHeight()
			if err != nil {
				m.logger.Error("monitor BlockchainMetadata err=%s", err.Error())
				continue
			}

			m.dataHeight <- curHeight

			if m.startBlock == 0 {
				m.startBlock = curHeight - 1
			}

			nowHeight := curHeight
			m.logger.Info("monitor nowheight=%d, lastheight=%d", nowHeight, m.startBlock)

			if nowHeight > m.startBlock {
				for index := m.startBlock + 1; index <= nowHeight; index++ {
					ack, err := m.retryGetBlock(index)
					if err != nil {
						m.logger.Error("monitor retryGetBlock(%d) err=%s", index, err.Error())
						continue
					}

					if ack.Header == nil {
						m.logger.Warn("monitor retryGetBlock(%d) err=header is nil", index)
						continue
					}

					blkHash := ack.Header.Hash()
					blockInfo := &BlockInfo{
						Version:   ack.Header.Version,
						Timestamp: ack.Header.Timestamp,
						Height:    ack.Header.Height,
						Hash:      blkHash.ToHexString(),
					}

					for _, tx := range ack.Transactions {
						txHash := tx.Hash()
						blockInfo.Txs = append(blockInfo.Txs, txHash.ToHexString())
					}

					m.data <- blockInfo
				}

				m.startBlock = nowHeight
			}
		case <-ctx.Done():
			m.logger.Info("user interrupt routine...")
			goto quit
		}
	}

quit:
	m.logger.Info("http timer monitor quit...")
}
