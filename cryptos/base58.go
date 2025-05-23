package cryptos

import (
	"encoding/hex"
	"math/big"

	"github.com/btcsuite/btcutil/base58"
)

func Base58ToBigInt(str string) big.Int {
	decoded := base58.Decode(str)
	hexStr := hex.EncodeToString(decoded)
	bigInt := big.Int{}
	bigInt.SetString(hexStr, 16)
	return bigInt
}
