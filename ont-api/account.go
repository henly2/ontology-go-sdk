package ont_api

import (
	"fmt"
	"github.com/ontio/ontology-crypto/ec"
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/common"
)

func newAccount() (*ontology_go_sdk.Account, error) {
	account := ontology_go_sdk.NewAccount(s.SHA256withECDSA)
	if account == nil {
		return nil, fmt.Errorf("account is nil")
	}

	return account, nil
}

func newAccountFromWifHex(wifHex string) (*ontology_go_sdk.Account, error) {
	wif, err := common.HexToBytes(wifHex)
	if err != nil {
		return nil, err
	}

	privkey, err := keypair.WIF2Key(wif)
	if err != nil {
		return nil, err
	}

	v, ok := privkey.(*ec.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key type error")
	}

	keyBytes := v.PrivateKey.D.Bytes()
	account, err := ontology_go_sdk.NewAccountFromPrivateKey(keyBytes, s.SHA256withECDSA)
	if err != nil {
		return nil, err
	}

	return account, nil
}

/////////////////////////////////////////////////////////////////
func NewAccount() (*ontology_go_sdk.Account, string, error) {
	account, err := newAccount()
	if err != nil {
		return account, "", err
	}

	wif, err := keypair.Key2WIF(account.PrivateKey)
	if err != nil {
		return nil, "", err
	}

	return account, common.ToHexString(wif), nil
}

func ParseWifHex(wifHex string) (*ontology_go_sdk.Account, error) {
	account, err := newAccountFromWifHex(wifHex)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func VerifyAddress(addr string) error {
	address, err := common.AddressFromBase58(addr)
	if err != nil {
		return err
	}

	if address.ToBase58() != addr {
		return fmt.Errorf("%s!=%s", address.ToBase58(), addr)
	}

	return nil
}
