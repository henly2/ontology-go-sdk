package ont_api

import (
	"encoding/hex"
)

func ReverseBytes(b []byte) []byte {
	for i := 0; i < len(b)/2; i++ {
		j := len(b) - i - 1
		b[i], b[j] = b[j], b[i]
	}
	return b
}

func ReverseHex(data string) (string, error) {
	d, err := hex.DecodeString(data)
	if err != nil {
		return "", nil
	}

	dr := ReverseBytes(d)
	return hex.EncodeToString(dr), nil
}
