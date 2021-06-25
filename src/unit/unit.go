package unit

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	b64 "encoding/base64"
)

func Encrypt(text string, passphrase string) (string) {
	key := passphrase
	data := []byte(passphrase)
	iv := md5.Sum(data)
	iva := iv[:]
	key = key + "\x00\x00\x00\x00\x00\x00\x00\x00"
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}
	//pad := __PKCS7Padding([]byte(text), block.BlockSize())
	byteText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, iva)
	encrypted := make([]byte, len(byteText))
	cfb.XORKeyStream(encrypted, byteText)

	return b64.StdEncoding.EncodeToString([]byte(string(encrypted)))
}

func Decrypt(encrypted string, passphrase string) (string) {
	ct, _ := b64.StdEncoding.DecodeString(encrypted)
	key := passphrase+"\x00\x00\x00\x00\x00\x00\x00\x00"
	data := []byte(passphrase)
	iv := md5.Sum(data)
	iva := iv[:]

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}

	cfb := cipher.NewCFBDecrypter(block, iva)
	dst := make([]byte, len(ct))

	cfb.XORKeyStream(dst, ct)
	return string(dst)
}
