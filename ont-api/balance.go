package ont_api

import (
	"fmt"
	"github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/common"
)

func (oac *OntApiClient) Api_GetBalance(addr, scriptHashHex string) (uint64, error) {
	tkc, err := oac.Api_GetTokenClient(scriptHashHex)
	if err != nil {
		return 0, err
	}

	address, err := common.AddressFromBase58(addr)
	if err != nil {
		return 0, err
	}

	if ontology_go_sdk.ONT_CONTRACT_ADDRESS.ToHexString() == tkc.ScriptHashHex {
		// ONT
		b, err := oac.OntSdk.Native.Ont.BalanceOf(address)
		if err != nil {
			return 0, err
		}
		return b, nil
	} else if ontology_go_sdk.ONG_CONTRACT_ADDRESS.ToHexString() == tkc.ScriptHashHex {
		// ONG
		b, err := oac.OntSdk.Native.Ong.BalanceOf(address)
		if err != nil {
			return 0, err
		}
		return b, nil
	} else {
		// oep4
		oep4Inst := tkc.Oep4Inst
		if oep4Inst == nil {
			return 0, fmt.Errorf("oep4 inst is nil")
		}

		b, err := oep4Inst.BalanceOf(address)
		if err != nil {
			return 0, err
		}

		return b.Uint64(), nil
	}
}
