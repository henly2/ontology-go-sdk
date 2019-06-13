package ont_api

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology-go-sdk/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
)

func MarshalTx(tx *types.MutableTransaction) ([]byte, error) {
	txIm, err := tx.IntoImmutable()
	if err != nil {
		return nil, err
	}

	return txIm.ToArray(), nil
}
func UnMarshalTx(data []byte) (*types.MutableTransaction, error) {
	txIm, err := types.TransactionFromRawBytes(data)
	if err != nil {
		return nil, err
	}

	return txIm.IntoMutable()
}

func (oac *OntApiClient) Api_GetTransaction(txHash string) (*TransactionInfo, error) {
	txEvent, err := oac.OntSdk.GetSmartContractEvent(txHash)
	if err != nil {
		return nil, err
	}
	if txEvent == nil {
		return nil, fmt.Errorf("txEvent is nil")
	}

	txInfo := &TransactionInfo{
		TxHash:      txEvent.TxHash,
		State:       txEvent.State,
		GasConsumed: txEvent.GasConsumed,
	}

	for _, evtNotify := range txEvent.Notify {
		tkc, err := oac.Api_GetTokenClient(evtNotify.ContractAddress)
		if err != nil {
			continue
		}

		states, ok := evtNotify.States.([]interface{})
		if !ok {
			return nil, fmt.Errorf("txEvent.Notify is not array")
		}
		if len(states) != 4 {
			return nil, fmt.Errorf("txEvent.Notify len is not 4")
		}

		operation := OperationInfo{
			TokenInfo: tkc.TokenInfo,
		}

		if tkc.IsNative {
			operation.FuncName, ok = states[0].(string)
			if !ok {
				return nil, fmt.Errorf("states[0] is not string")
			}
			operation.From, ok = states[1].(string)
			if !ok {
				return nil, fmt.Errorf("states[1] is not string")
			}
			operation.To, ok = states[2].(string)
			if !ok {
				return nil, fmt.Errorf("states[2] is not string")
			}
			operation.Amount, ok = states[3].(uint64)
			if !ok {
				return nil, fmt.Errorf("states[3] is not uint64")
			}
		} else {
			eventName, ok := states[0].(string)
			if !ok {
				return nil, fmt.Errorf("states[0] is not string")
			}
			from, ok := states[1].(string)
			if !ok {
				return nil, fmt.Errorf("states[1] is not string")
			}
			to, ok := states[2].(string)
			if !ok {
				return nil, fmt.Errorf("states[2] is not string")
			}
			amount, ok := states[3].(string)
			if !ok {
				return nil, fmt.Errorf("states[3] is not uint64")
			}

			evt, err := hex.DecodeString(eventName)
			if err != nil {
				return nil, fmt.Errorf("decode event name failed, err: %s", err)
			}

			from2, err := ReverseHex(from)
			if err != nil {
				return nil, fmt.Errorf("ReverseHex from failed, err: %s", err)
			}
			fromAddr, err := utils.AddressFromHexString(from2)
			if err != nil {
				return nil, fmt.Errorf("decode from failed, err: %s", err)
			}

			to2, err := ReverseHex(to)
			if err != nil {
				return nil, fmt.Errorf("ReverseHex from failed, err: %s", err)
			}
			toAddr, err := utils.AddressFromHexString(to2)
			if err != nil {
				return nil, fmt.Errorf("decode to failed, err: %s", err)
			}
			value, err := hex.DecodeString(amount)
			if err != nil {
				return nil, fmt.Errorf("decode value failed, err: %s", err)
			}

			operation.FuncName = string(evt)
			operation.From = fromAddr.ToBase58()
			operation.To = toAddr.ToBase58()
			operation.Amount = common.BigIntFromNeoBytes(value).Uint64()
		}

		// FIXME: 判断此payment是否是手续费
		if operation.To == gasToAddress &&
			operation.Amount == txInfo.GasConsumed &&
			operation.ScriptHashHex == ontology_go_sdk.ONG_CONTRACT_ADDRESS.ToHexString() {
			operation.IsFee = true
		}

		txInfo.Operations = append(txInfo.Operations, operation)
	}

	return txInfo, nil
}

func (oac *OntApiClient) Api_SendTransaction(wifHex string, toAddr string, scriptHashHex string, amount uint64) (string, error) {
	tkc, err := oac.Api_GetTokenClient(scriptHashHex)
	if err != nil {
		return "", err
	}

	account, err := newAccountFromWifHex(wifHex)
	if err != nil {
		return "", err
	}

	toAddress, err := common.AddressFromBase58(toAddr)
	if err != nil {
		return "", err
	}

	var tx *types.MutableTransaction
	if ontology_go_sdk.ONT_CONTRACT_ADDRESS.ToHexString() == tkc.ScriptHashHex {
		// build
		tx, err = oac.OntSdk.Native.Ont.NewTransferTransaction(gasPrice, gasLimit, account.Address, toAddress, amount)
		if err != nil {
			return "", nil
		}
	} else if ontology_go_sdk.ONG_CONTRACT_ADDRESS.ToHexString() == tkc.ScriptHashHex {
		// build
		tx, err = oac.OntSdk.Native.Ong.NewTransferTransaction(gasPrice, gasLimit, account.Address, toAddress, amount)
		if err != nil {
			return "", nil
		}
	} else {
		// oep4
		oep4Inst := tkc.Oep4Inst
		if oep4Inst == nil {
			return "", fmt.Errorf("oep4 inst is nil")
		}

		tx, err = oac.OntSdk.NeoVM.NewNeoVMInvokeTransaction(gasPrice, gasLimit, oep4Inst.ContractAddress,
			[]interface{}{"transfer", []interface{}{account.Address, toAddress, amount}})
		if err != nil {
			return "", err
		}
	}

	err = oac.OntSdk.SignToTransaction(tx, account)
	if err != nil {
		return "", nil
	}

	res, err := oac.OntSdk.SendTransaction(tx)
	if err != nil {
		return "", nil
	}

	return res.ToHexString(), nil
}

func (oac *OntApiClient) Api_BuildTransaction(fromAddr string, toAddr string, scriptHashHex string, amount uint64) ([]byte, error) {
	tkc, err := oac.Api_GetTokenClient(scriptHashHex)
	if err != nil {
		return nil, err
	}

	fromAddress, err := common.AddressFromBase58(fromAddr)
	if err != nil {
		return nil, err
	}

	toAddress, err := common.AddressFromBase58(toAddr)
	if err != nil {
		return nil, err
	}

	var tx *types.MutableTransaction
	if ontology_go_sdk.ONT_CONTRACT_ADDRESS.ToHexString() == tkc.ScriptHashHex {
		// build
		tx, err = oac.OntSdk.Native.Ont.NewTransferTransaction(gasPrice, gasLimit, fromAddress, toAddress, amount)
		if err != nil {
			return nil, err
		}
	} else if ontology_go_sdk.ONG_CONTRACT_ADDRESS.ToHexString() == tkc.ScriptHashHex {
		// build
		tx, err = oac.OntSdk.Native.Ong.NewTransferTransaction(gasPrice, gasLimit, fromAddress, toAddress, amount)
		if err != nil {
			return nil, err
		}
	} else {
		// oep4
		oep4Inst := tkc.Oep4Inst
		if oep4Inst == nil {
			return nil, fmt.Errorf("oep4 inst is nil")
		}

		tx, err = oac.OntSdk.NeoVM.NewNeoVMInvokeTransaction(gasPrice, gasLimit, oep4Inst.ContractAddress,
			[]interface{}{"transfer", []interface{}{fromAddress, toAddress, amount}})
		if err != nil {
			return nil, err
		}
	}

	// pack
	bytes, err := MarshalTx(tx)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (oac *OntApiClient) Api_SignTransaction(wifHex string, txData []byte) ([]byte, string, error) {
	account, err := newAccountFromWifHex(wifHex)
	if err != nil {
		return nil, "", err
	}

	// unpack
	tx, err := UnMarshalTx(txData)
	if err != nil {
		return nil, "", err
	}

	err = oac.OntSdk.SignToTransaction(tx, account)
	if err != nil {
		return nil, "", err
	}

	// pack
	bytes, err := MarshalTx(tx)
	if err != nil {
		return nil, "", err
	}

	hash := tx.Hash()
	if hash == common.UINT256_EMPTY {
		return nil, "", fmt.Errorf("hash is empty")
	}

	return bytes, hash.ToHexString(), nil
}

func (oac *OntApiClient) Api_PostTransaction(txSignedData []byte) (string, error) {
	// unpack
	tx, err := UnMarshalTx(txSignedData)
	if err != nil {
		return "", err
	}

	res, err := oac.OntSdk.SendTransaction(tx)
	if err != nil {
		return "", nil
	}

	return res.ToHexString(), nil
}
