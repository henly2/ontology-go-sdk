package ont_api

import (
	"fmt"
	"github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology-go-sdk/oep4"
	"github.com/ontio/ontology-go-sdk/ont-api/httpclient"
	"github.com/ontio/ontology-go-sdk/utils"
	"sync"
)

type (
	TokenClient struct {
		*TokenInfo

		IsNative bool
		Oep4Inst *oep4.Oep4
	}

	OntApiClient struct {
		OntSdk *ontology_go_sdk.OntologySdk

		ExplorerSdk *httpclient.HttpClient

		rwmutex      sync.RWMutex
		tokenClients map[string]*TokenClient
	}
)

func NewOntApiClient(rpcUrl, explorerUrl string) *OntApiClient {
	oac := &OntApiClient{
		OntSdk:       ontology_go_sdk.NewOntologySdk(),
		tokenClients: make(map[string]*TokenClient),
	}

	oac.OntSdk.NewRpcClient().SetAddress(rpcUrl)
	oac.ExplorerSdk = httpclient.NewHttpClient(explorerUrl)

	return oac
}

func (oac *OntApiClient) Api_RegistTokenInfo(scriptHashHex string) (*TokenInfo, error) {
	var (
		err      error
		tkClient *TokenClient
	)

	// read
	func() {
		oac.rwmutex.RLock()
		defer oac.rwmutex.RUnlock()

		if tkc, ok := oac.tokenClients[scriptHashHex]; ok {
			tkClient = tkc
		}
	}()
	if tkClient != nil {
		return tkClient.TokenInfo, nil
	}

	// get
	tkClient, err = func() (*TokenClient, error) {
		if ontology_go_sdk.ONT_CONTRACT_ADDRESS.ToHexString() == scriptHashHex {
			// ONT
			inst := oac.OntSdk.Native.Ont

			name, err := inst.Name()
			if err != nil {
				return nil, err
			}
			symbol, err := inst.Symbol()
			if err != nil {
				return nil, err
			}
			decimals, err := inst.Decimals()
			if err != nil {
				return nil, err
			}
			totalSupply, err := inst.TotalSupply()
			if err != nil {
				return nil, err
			}

			tk := &TokenInfo{
				Name:          name,
				Symbol:        symbol,
				Precision:     int(decimals),
				TotalSupply:   totalSupply,
				ScriptHashHex: scriptHashHex,
			}

			tkClient := &TokenClient{
				TokenInfo: tk,
				IsNative:  true,
			}

			return tkClient, nil
		} else if ontology_go_sdk.ONG_CONTRACT_ADDRESS.ToHexString() == scriptHashHex {
			// ONT
			inst := oac.OntSdk.Native.Ong

			name, err := inst.Name()
			if err != nil {
				return nil, err
			}
			symbol, err := inst.Symbol()
			if err != nil {
				return nil, err
			}
			decimals, err := inst.Decimals()
			if err != nil {
				return nil, err
			}
			totalSupply, err := inst.TotalSupply()
			if err != nil {
				return nil, err
			}

			tk := &TokenInfo{
				Name:          name,
				Symbol:        symbol,
				Precision:     int(decimals),
				TotalSupply:   totalSupply,
				ScriptHashHex: scriptHashHex,
			}

			tkClient := &TokenClient{
				TokenInfo: tk,
				IsNative:  true,
			}

			return tkClient, nil
		} else {
			contractAddr, err := utils.AddressFromHexString(scriptHashHex)
			if err != nil {
				return nil, err
			}
			inst := oep4.NewOep4(contractAddr, oac.OntSdk)

			name, err := inst.Name()
			if err != nil {
				return nil, err
			}
			symbol, err := inst.Symbol()
			if err != nil {
				return nil, err
			}
			decimals, err := inst.Decimals()
			if err != nil {
				return nil, err
			}
			if decimals.Uint64() > 20 {
				return nil, fmt.Errorf("%s decimals(%d) really > 20", name, decimals.Uint64())
			}

			totalSupply, err := inst.TotalSupply()
			if err != nil {
				return nil, err
			}

			tk := &TokenInfo{
				Name:          name,
				Symbol:        symbol,
				Precision:     int(decimals.Uint64()),
				TotalSupply:   totalSupply.Uint64(),
				ScriptHashHex: scriptHashHex,
			}

			tkClient := &TokenClient{
				TokenInfo: tk,
				IsNative:  false,
				Oep4Inst:  inst,
			}

			return tkClient, nil
		}

		return nil, fmt.Errorf("unknown error")
	}()
	if err != nil {
		return nil, err
	}

	// write
	func() {
		oac.rwmutex.Lock()
		defer oac.rwmutex.Unlock()

		oac.tokenClients[scriptHashHex] = tkClient
	}()

	return tkClient.TokenInfo, nil
}

func (oac *OntApiClient) Api_GetTokenInfo(tokenKey string) (*TokenInfo, error) {
	oac.rwmutex.RLock()
	defer oac.rwmutex.RUnlock()

	if tkc, ok := oac.tokenClients[tokenKey]; ok {
		return tkc.TokenInfo, nil
	}

	return nil, fmt.Errorf("not find token %s", tokenKey)
}

func (oac *OntApiClient) Api_GetTokenClient(tokenKey string) (*TokenClient, error) {
	oac.rwmutex.RLock()
	defer oac.rwmutex.RUnlock()

	if tkc, ok := oac.tokenClients[tokenKey]; ok {
		return tkc, nil
	}

	return nil, fmt.Errorf("not find token %s", tokenKey)
}

func (oac *OntApiClient) Api_Height() (uint64, error) {
	h, err := oac.OntSdk.GetCurrentBlockHeight()
	if err != nil {
		return 0, err
	}

	return uint64(h), nil
}

func (oac *OntApiClient) Api_GetBlock(index uint32) (*BlockInfo, error) {
	ack, err := oac.OntSdk.GetBlockByHeight(index)
	if err != nil {
		return nil, err
	}

	if ack.Header == nil {
		return nil, fmt.Errorf("block header is nil")
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

	return blockInfo, nil
}
