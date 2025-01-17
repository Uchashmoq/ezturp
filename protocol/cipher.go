package protocol

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// AesEncrypt 原数据	32字节秘钥，16字节初始化向量
func AesEncrypt(plaintext, key, iv []byte) []byte {
	// 创建一个AES加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	// 创建一个填充块
	padding := block.BlockSize() - len(plaintext)%block.BlockSize()
	paddedPlaintext := append(plaintext, bytes.Repeat([]byte{byte(padding)}, padding)...)
	// 创建一个CBC模式的加密器
	mode := cipher.NewCBCEncrypter(block, iv)
	// 加密数据
	ciphertext := make([]byte, len(paddedPlaintext))
	mode.CryptBlocks(ciphertext, paddedPlaintext)
	return ciphertext
}

func AesDecrypt(ciphertext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypter := cipher.NewCBCDecrypter(block, iv)
	// 解密数据
	decryptedText := make([]byte, len(ciphertext))
	decrypter.CryptBlocks(decryptedText, ciphertext)
	// 去除填充
	return removePadding(decryptedText)
}

func removePadding(data []byte) ([]byte, error) {
	padding := int(data[len(data)-1])
	if len(data)-padding < 0 || len(data)-padding >= len(data) {
		return nil, errors.New("failed to decrypt")
	}
	return data[:len(data)-padding], nil
}
