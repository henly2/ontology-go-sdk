package example

import (
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology-go-sdk/ont-api"
	"github.com/ontio/ontology-go-sdk"
)

const pubNode = "http://dappnode1.ont.io:20336"
const pubExp = "https://explorer.ont.io"

func main() {
	ontClient := ont_api.NewOntApiClient(pubNode, pubExp)

	// regist native
	ontClient.RegistTokenInfo(ontology_go_sdk.ONT_CONTRACT_ADDRESS.ToHexString())
	ontClient.RegistTokenInfo(ontology_go_sdk.ONG_CONTRACT_ADDRESS.ToHexString())

	// height
	h, err := ontClient.OntSdk.GetCurrentBlockHeight()
	if err != nil {
		fmt.Println("error: ", err)
		return
	}
	fmt.Println("current height=", h)

	// account
	newAccount := func() {
		// build
		account, wifHex, err := ont_api.NewAccount()
		if err != nil {
			fmt.Println("error: ", err)
			return
		}

		fmt.Println("wifhex:", wifHex)
		fmt.Println("address:", account.Address.ToBase58())

		// check
		account2, err := ont_api.ParseWifHex(wifHex)
		if err != nil {
			fmt.Println("error: ", err)
			return
		}
		fmt.Println("address2:", account2.Address.ToBase58())

		if account2.Address.ToBase58() != account.Address.ToBase58() {
			fmt.Println("err!!!!")
			return
		}
	}
	newAccount()

	// tx
	scanTx := func() {
		tx, err := ontClient.GetTransaction("bf720958a1e1e2a3aeac87078581f9c103c4743f0f5c106db82503b6417ec725")
		if err != nil {
			fmt.Println("error: ", err)
			return
		}

		txExp, err := ontClient.Exp_GetTransaction("bf720958a1e1e2a3aeac87078581f9c103c4743f0f5c106db82503b6417ec725")
		if err != nil {
			fmt.Println("error: ", err)
			return
		}

		s, _ := json.MarshalIndent(tx, "", "  ")
		fmt.Println("res:", string(s))

		sExp, _ := json.MarshalIndent(txExp, "", "  ")
		fmt.Println("res exp:", string(sExp))
	}
	scanTx()

	scanBlock := func() {
		block, err := ontClient.OntSdk.GetBlockByHeight(4741217)
		if err != nil {
			fmt.Println("error: ", err)
			return
		}
		fmt.Println("==========")
		fmt.Println("tx count=", len(block.Transactions))
		fmt.Println("tx time=", block.Header.Timestamp)
		for _, tx := range block.Transactions {
			hash := tx.Hash()
			tx, err := ontClient.GetTransaction(hash.ToHexString())
			if err != nil {
				fmt.Println("error: ", err)
				continue
			}

			dd, _ := json.MarshalIndent(tx, "", "  ")
			fmt.Println("res exp:", string(dd))
		}
	}
	scanBlock()
}
