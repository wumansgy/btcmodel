package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"fmt"
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"crypto/elliptic"
	"strings"
)

//1. 定义Input
//2. 定义Output
//3. 定义交易Transaction
//4. 生成交易ID
//5. 创建交易（创建Coinbase交易）
//6. 改写我们区块结构：data-> Transaction

type TXInput struct {
	//1. 引用的交易id
	TxId []byte

	//2. 引用的output的索引
	Index int64
	//3. 解锁脚本

	Sig    []byte //签名
	PubKey []byte //付款人的公钥，字节流
}

type TXOutput struct {
	//1. 金额
	Value float64 //一定要大写

	//2. 锁定脚本
	//PubKeyHash string //也用地址代替，只要比对地址是否相同，就认为可以解锁
	PubKeyHash []byte //收款人公钥的哈希
}

//给TXOutput提供一个方法，给定地址，生成公钥哈希, 赋值给PubKeyHash，整个过程叫Lock
func (output *TXOutput) LockWithHash(address string) {
	decodeInfo := base58.Decode(address)

	pubkeyHash := decodeInfo[1:len(decodeInfo)-4]
	output.PubKeyHash = pubkeyHash
}

//提供一个创建TXOutput的方法
func NewTXOutput(value float64, address string) *TXOutput {
	output := TXOutput{Value: value,}
	output.LockWithHash(address)

	return &output
}

type Transaction struct {
	TXID []byte
	//多个输入
	TXInputs []TXInput
	//多个输出
	TXOutputs []TXOutput
}

func (tx *Transaction) SetTxID() {
	//使用gob编码，生成交易的哈希

	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash := sha256.Sum256(buffer.Bytes())
	tx.TXID = hash[:]
}

const reward = 12.5

//创建挖矿交易CoinbaseTx
//比特币挖矿的人可以有权利填充zhege Sig字段， 中本聪的创世语就是写在这里

func NewCoinbaseTx(miner, data string) *Transaction {
	//挖矿交易的特点， 没有输入， 只有输出
	input := TXInput{nil, -1, nil, []byte(data)}
	//output := TXOutput{12.5, miner}
	output := NewTXOutput(reward, miner)

	tx := Transaction{nil, []TXInput{input}, []TXOutput{*output}}
	tx.SetTxID()

	return &tx
}

//检查是否为挖矿交易
func (tx *Transaction) IsCoinbaseTx() bool {
	if tx.TXInputs[0].TxId == nil && len(tx.TXInputs) == 1 && tx.TXInputs[0].Index == -1 {
		fmt.Printf("找到挖矿交易...\n")
		return true
	}
	return false
}

//1. 找到所有合适utxos的集合
//
//1. 如果找到总额小于转账金额，转账失败
//
//1. 将所有的utxo转成input
//2. 创建output
//3. 如果有剩余，找零

func NewTransaction(from, to string, amount float64, bc *BlockChain) *Transaction {
	fmt.Printf("创建交易中...\n")
	//1. 钱包这就在这里使用，因为要对交易进行签名，要使用到私钥
	ws := NewWallets()

	if ws.WalletsMap[from] == nil {
		fmt.Printf("本地没有 %s 的钱包，无法创建交易\n", from)
		return nil
	}

	wallet := ws.WalletsMap[from]

	privateKey := wallet.PrivateKey
	pubKey := wallet.PubKey

	//私钥赋值签名，目前先不用
	//公钥负责寻找能够使用utxo,
	//由于每一个utxo是使用pubkeyHash锁定的，所以我们传递公钥哈希到FindNeedUTXO函数
	pubKeyHash := HashPubKey(pubKey)

	//spentUTXOs := make(map[string][]int64)

	//1. 找到所有合适utxos的集合, 并且返回来
	//spentUTXOs[0x222] = []int64{0}
	//spentUTXOs[0x333] = []int64{0} //中间值
	//spentUTXOs[0x333] = []int64{0, 1}
	spentUTXOs, calcMoney := bc.FindNeedUTXOs(pubKeyHash, amount)

	if calcMoney < amount {
		return nil
	}

	var inputs []TXInput
	var outputs []TXOutput

	//创建input
	for txid, indexArray := range spentUTXOs {
		//key 是2222， 3333
		//value 0,     0, 1
		for _, i := range indexArray {
			//每一个output都要创建一个input
			input := TXInput{[]byte(txid), i, nil, pubKey}
			inputs = append(inputs, input)
		}
	}

	//创建output
	//output := TXOutput{amount, to}
	output := NewTXOutput(amount, to)
	outputs = append(outputs, *output)

	//找零
	if calcMoney > amount {
		//outputs = append(outputs, TXOutput{calcMoney - amount, from})
		output1 := NewTXOutput(calcMoney-amount, from)
		outputs = append(outputs, *output1)
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetTxID()
	if !bc.SignTransaction(&tx, privateKey) {
		fmt.Printf("签名失败!\n")
		return nil
	}

	fmt.Printf("交易创建成功!\n")
	return &tx
}

//1. privateKey
//2. inputs所引用的交易的集合
//map[0x222] = Transaction222
//map[0x333] = Transaction333

func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey, prevTXs map[string]*Transaction) bool {
	fmt.Printf("对交易进行签名...\n")

	if tx.IsCoinbaseTx() {
		return true
	}

	//1. 生成原始交易的副本
	txCopy := tx.TrimmedCopy()

	//一定要注意，是对copy进行遍历，修改，而不是原始交易，原始交易是为了填充sig
	for i, input := range txCopy.TXInputs {
		//对每一个input进行赋值sig
		//找到引用的交易，
		//找到引用的output，将pubKeyHash拿过来，赋值给input
		//map[0x222] = Transaction222
		prevTX := prevTXs[string(input.TxId)]

		//2. 根据引用交易进行pubKeyHash赋值
		//将引用的output的公钥hash拿过来赋值给input的pubKey字段
		//input.PubKey = prevTX.TXOutputs[input.Index].PubKeyHash
		//一定要注意，input是一个副本，怼他赋值无法修改txCopy里面的对应的input，导致后面SetTxID是错的
		txCopy.TXInputs[i].PubKey = prevTX.TXOutputs[input.Index].PubKeyHash

		//3. 对交易进行哈希运算，得到要签名的数据
		txCopy.SetTxID()

		//将次值设置为空，避免Verify端拼接数据时不一致
		txCopy.TXInputs[i].PubKey = nil
		//input.PubKey = nil

		//TXID里面存储了交易哈希值，所以直接复用
		signDataHash := txCopy.TXID

		//4. 签名，并赋值给原始交易
		// func Sign(rand io.Reader, priv *PrivateKey, hash []byte) (r, s *big.Int, err error) {
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, signDataHash[:])

		if err != nil {
			log.Panic(err)
		}

		//将r，s拼接成字节流
		signature := append(r.Bytes(), s.Bytes()...)

		//对原始的交易的sig进行赋值，而不是copy里面input
		tx.TXInputs[i].Sig = signature
	}

	return true
}

//校验过程需要1. 数据， 2.公钥(交易中含有)，3.签名(交易中含有)
//数据是矿工自己根据所校验交易提供的信息寻找出来的
func (tx *Transaction) Verify(prevTXs map[string]*Transaction) bool {
	//1. 签名的时候是对每一个引用的input进行签名，所以校验的时候，要对每一input的签名进行校验

	//2. 拿到数据，这个数据要根据，都在input里面

	txCopy := tx.TrimmedCopy()

	//遍历原始数据，而不是txCopy
	for i, input := range tx.TXInputs {
		//要拿取引用output的公钥哈希，拼接到txCopy里面，拼接处data
		prevTX := prevTXs[string(input.TxId)]

		txCopy.TXInputs[i].PubKey = prevTX.TXOutputs[input.Index].PubKeyHash

		//做哈希运算, 得到要签名的数据
		txCopy.SetTxID()

		data := txCopy.TXID //这个就是我们要校验的原始数据

		//只要保证签名端和校验端能够得到相同的数据即可，设置nil，语义更加明确
		txCopy.TXInputs[i].PubKey = nil

		pubKey := input.PubKey //X, Y拼接的字节流，我们要转回来
		signature := input.Sig //r, s拼接的字节流

		//根据signature 切出来r1, s1, 一分为二
		r1 := big.Int{}
		s1 := big.Int{}

		r1Data := signature[:len(signature)/2]
		s1Data := signature[len(signature)/2:]

		r1.SetBytes(r1Data)
		s1.SetBytes(s1Data)

		//切pubkey字节流
		x1 := big.Int{}
		y1 := big.Int{}

		x1Data := pubKey[:len(pubKey)/2]
		y1Data := pubKey[len(pubKey)/2:]

		x1.SetBytes(x1Data)
		y1.SetBytes(y1Data)

		curve := elliptic.P256()
		pubKeyOrigin := ecdsa.PublicKey{curve, &x1, &y1}

		if !ecdsa.Verify(&pubKeyOrigin, data, &r1, &s1) {
			fmt.Printf("校验失败!\n")
			return false
		}
	}

	fmt.Printf("恭喜，校验成功！\n")
	return true
}

//trim修剪
//完成交易复制，同时做一些修改
func (tx *Transaction) TrimmedCopy() *Transaction {
	var inputs []TXInput

	//复制input，同时将sig和pubkey设置为nil
	for _, input := range tx.TXInputs {
		inputs = append(inputs, TXInput{input.TxId, input.Index, nil, nil})
	}

	return &Transaction{tx.TXID, inputs, tx.TXOutputs}
}

func (tx *Transaction) String() string {
	//fmt.Sprintf("打印交易细节...\n")
	//return string("hello")
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.TXID))

	for i, input := range tx.TXInputs {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.TxId))
		lines = append(lines, fmt.Sprintf("       Index:       %d", input.Index))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Sig))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.TXOutputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %f", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
