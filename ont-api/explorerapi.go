package ont_api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (oac *OntApiClient) Exp_GetTransaction(txHash string) (*ExpTransactionInfo, error) {
	code, d, err := oac.ExplorerSdk.Get("/api/v1/explorer/transaction/"+txHash, "", nil)
	if err != nil {
		return nil, err
	}

	if code != http.StatusOK {
		return nil, fmt.Errorf("http code:%d", code)
	}

	ack := &ExpTransactionInfo{}
	err = json.Unmarshal(d, ack)
	if err != nil {
		return nil, err
	}

	return ack, nil
}
