package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"bytes"
)

//定义一个秘钥对结构， KeyPair
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey

	//PubKey ecdsa.PublicKey

	PubKey []byte //为了传输方便，在对端可以还原
}

func NewWallet() *Wallet {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	pubKeyRaw := privateKey.PublicKey

	//真正传递的形式
	pubKey := append(pubKeyRaw.X.Bytes(), pubKeyRaw.Y.Bytes()...)

	return &Wallet{PrivateKey: privateKey, PubKey: pubKey}
}

func (w *Wallet) GetAddress() string {

	ripHashValue := HashPubKey(w.PubKey)

	version := byte(00)

	payload := append([]byte{version}, ripHashValue...)

	checkSum := checksum(payload)

	payload = append(payload, checkSum...)

	address := base58.Encode(payload)

	return address
}

func IsValidAddress(address string) bool {
	//1. 解码base58
	decodeInfo := base58.Decode(address)

	//2. 截取前21字节和 后四个字节
	payload := decodeInfo[0: len(decodeInfo)-4]
	checksum1 := decodeInfo[len(decodeInfo)-4:]

	//3. 对前21字节进行checksum计算
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	checksum2 := secondHash[0:4]

	//4. 比较生成cheksum1和截取cheksum2
	return bytes.Equal(checksum1, checksum2)
}

func HashPubKey(pubkey []byte) []byte {
	hash := sha256.Sum256(pubkey)

	rip160Hasher := ripemd160.New()
	_, err := rip160Hasher.Write(hash[:])
	if err != nil {
		log.Panic(err)
	}

	ripHashValue := rip160Hasher.Sum(nil)

	return ripHashValue
}

func checksum(payload []byte) []byte {

	firsHash := sha256.Sum256(payload)
	sencondHash := sha256.Sum256(firsHash[:])

	checkSum := sencondHash[0:4]
	return checkSum
}
