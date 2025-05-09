package ncrypto

import (
	"crypto/md5"
	"encoding/hex"
)

func MD5HexEncode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
