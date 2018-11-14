package main

import (
	"time"
	"bytes"
	"encoding/binary"
	"log"
	"encoding/gob"
	"crypto/sha256"
)

//data , prevHash, Hash

const genesisInfo = "2009年1月3日，财政大臣正处于实施第二轮银行紧急援助的边缘"

type Block struct {
	Version       uint64 //版本号
	PrevBlockHash []byte //前区块哈希值

	MerkelRoot []byte //这是一个哈希值，后面v5用到

	TimeStamp uint64 //时间戳，从1970.1.1到现在的秒数

	Difficulty uint64 //通过这个数字，算出一个哈希值：0x00010000000xxx

	Nonce uint64 // 这是我们要找的随机数，挖矿就找证书

	Hash []byte //当前区块哈希值, 正常的区块不存在，我们为了方便放进来

	//Data []byte //数据本身，区块体，先用字符串表示，v4版本的时候会引用真正的交易结构
	Transactions []*Transaction
}

func NewBlock(txs []*Transaction, prevHash []byte) *Block {
	block := Block{
		Version:       00,
		PrevBlockHash: prevHash,
		MerkelRoot:    []byte{}, //先填写为空
		TimeStamp:     uint64(time.Now().Unix()),
		Difficulty:    difficulty,
		Nonce:         0,        //目前不挖矿，随便写一个值
		Hash:          []byte{}, //见SetHash函数
		//Data:          []byte(data),
		Transactions: txs,
	}

	block.setMerkelRoot()

	//pow运算
	pow := NewProofOfWork(block)
	hash, nonce := pow.Run()

	block.Hash = hash
	block.Nonce = nonce
	return &block
}

//对区块进行序列化
func (block *Block) Serialize() []byte {
	var buffer bytes.Buffer

	//1. 定义编码器
	encoder := gob.NewEncoder(&buffer)
	//2. 使用编码器编码

	err := encoder.Encode(block)
	// 一定要记得校验
	if err != nil {
		log.Panic(err)
	}

	return buffer.Bytes()
}

//对区块的字节流进行解码, 返回Block
func Deserialize(data []byte) *Block {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	var block Block

	//2. 对传过来字节流进行解码
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

//将uint转换成[]byte
func Uint2Byte(num uint64) []byte {
	var buffer bytes.Buffer

	//这是一个序列化的过程, 将num转换成buffer字节流
	err := binary.Write(&buffer, binary.BigEndian, &num)
	if err != nil {
		log.Panic(err)
	}

	return buffer.Bytes()
}


//创建一个简单的MerkelRoot, 使用block中的交易作为数据来源
//只是将多个交易的哈希拼接起来，做sha256
func (block *Block)setMerkelRoot()  {
	var info []byte

	for _, tx := range block.Transactions {
		//只是将多个交易的哈希拼接起来，做sha256
		info = append(info, tx.TXID...)
	}

	hash := sha256.Sum256(info)

	block.MerkelRoot = hash[:]
}
