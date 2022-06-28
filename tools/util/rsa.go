package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/url"
)

const _privKey string = `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQCaENV6+7hVMpADLhpWhvih1s2irUeSeTiZDb7VLRJVTAswyjO1
LBpmbYTaQ0RRIPDC/ztfApX85Mgi1xuNCgE2x1zDGPDJZFhSMksEwf1QEhI0B/Ml
9tq6iTFM2h5mqPo/i7Ni3HyuzpIoOsm7iR6JPxDARE3ZMqJ0HXncCrsoiwIDAQAB
AoGAWO56rED9SIClTJCiN2w1vQXHMa4gcFZ06zRaAafAu2fn1cQCUQQiQRna5DqM
BuCi2YyG8vMFLTPKqwHML3+k7tqM7KNla9LqH2WILOW4ww6R+MNu9NO7m88s1+pF
Nl0vjJAfl5dtQKZVqbydp8tVWSGQR6t6qidzq2NwjJy3WLECQQDMGSZL+LweBZ3E
WkxkqGn9uer1wSbgFgB1CEtM3CUSxUpOc3CH6JyFgCsTdYFPu3qKg5J47z4pLMH2
WRt+T66XAkEAwT6OYnDLx8BUeV2400/U6LeSKPK/zFc875mMorPz9xlZjqTD2NDQ
AtbKJ54KVVdHjnFs5IOOefUdcdnRojlILQJAbcohlcCJwUSYJ6XDbmpCCeDXCbgL
Z4OuX0ZE62WI8935KNZkdFemyxG1GlSdaPya4KQCSNe5goC3HgO1DG9kpQJAE6oW
+SN7SSdkMTl9TluIUeokQHB7XgLem48nhYMEZ3e36lEP8OdG05Mh3Sgy6v5HtNIL
/7D3daegyG4e7AAiPQJAfnffZYCxcshhdTjxYYjy4NPyijuU84kGdFUZS5LF7zYd
Qqm0RBw6gC4dnEiFi2mCGuZbF1KOqiZ2L9TxFsHsmw==
-----END RSA PRIVATE KEY-----`

const _publicKey string = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCaENV6+7hVMpADLhpWhvih1s2i
rUeSeTiZDb7VLRJVTAswyjO1LBpmbYTaQ0RRIPDC/ztfApX85Mgi1xuNCgE2x1zD
GPDJZFhSMksEwf1QEhI0B/Ml9tq6iTFM2h5mqPo/i7Ni3HyuzpIoOsm7iR6JPxDA
RE3ZMqJ0HXncCrsoiwIDAQAB
-----END PUBLIC KEY-----`

// ReadPrivKeyFile get priv rsa key from file or return default key
func ReadPrivKeyFile(path string) (*rsa.PrivateKey, error) {
	var fileBytes []byte
	var err error
	if len(path) <= 0 {
		fileBytes = []byte(_privKey)
	} else {
		fileBytes, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}
	block, _ := pem.Decode([]byte(fileBytes))
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, err
}

func LoadPublicKey() (*rsa.PublicKey, error) {

	block, _ := pem.Decode([]byte(_publicKey))
	if block == nil {
		return nil, fmt.Errorf("failed to parsing a pem")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break
	}
	return nil, fmt.Errorf("input is not a rsa public key")
}

// Decrypt RSA
func Decrypt(ciphertext []byte, keyPath string) ([]byte, error) {
	privKey, err := ReadPrivKeyFile(keyPath)
	if err != nil {
		return nil, err
	}
	plaintext, err := rsa.DecryptPKCS1v15( // 암호화된 데이터를 개인 키로 복호화
		rand.Reader, privKey, ciphertext,
	)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// Encrypt DSL Encrypt
func Encrypt(plainText string) ([]string, error) {
	pubKey, err := LoadPublicKey()
	if err != nil {
		return nil, err
	}
	// URL encoding
	plainText = url.QueryEscape(plainText)
	// Split -> Ciphering -> Base64 Encode
	var encryptText []string
	for i := 0; i < len(plainText); i += 100 {
		var splitStr string
		if i+100 > len(plainText) {
			splitStr = plainText[i:len(plainText)]
		} else {
			splitStr = plainText[i : i+100]
		}
		encText, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(splitStr))
		if err != nil {
			return nil, err
		}
		baseDecode := base64.StdEncoding.EncodeToString(encText)
		encryptText = append(encryptText, baseDecode)
	}
	return encryptText, nil
}
