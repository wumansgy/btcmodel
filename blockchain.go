package main

import (

	"log"
	"fmt"
	"os"
	"time"
	"bytes"
	"crypto/ecdsa"
	"github.com/boltdb/bolt"
	"github.com/base58"
)

//定义一个区块链结构，使用bolt数据库进行保存
type BlockChain struct {
	//数据库的句柄
	Db *bolt.DB

	//最后一个区块的哈希值
	lastHash []byte
}

const blockChainName = "blockChain.db"
const blockBucket = "blockBucket"
const lastHashKey = "lastHashKey"

//定义一个创建区块链的方法
//就是返回一个区块链的实例instance，已经存在直接返回，不存在，创建再返回

//创建一个新的区块链
func CreateBlockChain(address string) *BlockChain {
	if IsFileExist(blockChainName) {
		fmt.Printf("区块链已经存在!\n")
		//os.Exit(1)
		return nil
	}

	var lastHash []byte
	db, err := bolt.Open(blockChainName, 0600, nil)

	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		//2. 找到我们的桶，通过桶的名字
		// Returns nil if the bucket does not exist.
		bucket := tx.Bucket([]byte(blockBucket))

		//如果没有找到，先创建
		if bucket == nil {
			bucket, err = tx.CreateBucket([]byte(blockBucket))
			if err != nil {
				log.Panic(err)
			}

			//3. 写数据
			//在创建区块链的时候，添加一个创世块genesisBlock
			coinbase := NewCoinbaseTx(address, genesisInfo)
			genesisBlock := NewBlock([]*Transaction{coinbase}, []byte{})

			err = bucket.Put(genesisBlock.Hash, genesisBlock.Serialize() /*将区块序列化成字节流*/)
			if err != nil {
				log.Panic(err)
			}

			//一定要记得更新"lastHashKey" 这个key对应的值，最后一个区块的哈希
			err = bucket.Put([]byte(lastHashKey), genesisBlock.Hash)

			//更新内存中最后区块哈希值
			lastHash = genesisBlock.Hash
		}

		return nil
	})

	return &BlockChain{db, lastHash}
}

// 返回一个已经存在实例
func NewBlockChain() *BlockChain {
	//在创建区块链的时候，添加一个创世块genesisBlock
	//genesisBlock := NewBlock(genesisInfo, []byte{})
	//blockChain := BlockChain{blocks: []*Block{genesisBlock}}
	//return &blockChain

	if !IsFileExist(blockChainName) {
		fmt.Printf("请先创建区块链!\n")
		//os.Exit(1)
		return nil
	}

	var lastHash []byte
	db, err := bolt.Open(blockChainName, 0600, nil)

	if err != nil {
		log.Panic(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		//2. 找到我们的桶，通过桶的名字
		// Returns nil if the bucket does not exist.
		bucket := tx.Bucket([]byte(blockBucket))

		//如果没有找到，先创建
		if bucket == nil {
			fmt.Printf("获取区块链实例时bucket不应为空!")
			os.Exit(1)
		}

		lastHash = bucket.Get([]byte(lastHashKey))

		return nil
	})

	return &BlockChain{db, lastHash}
}

func (bc *BlockChain) AddBlock(txs []*Transaction) {
	//在挖矿时进行校验

	//一个包含所有校验成功交易的集合
	validTXs := []*Transaction{}

	for _, tx := range txs {
		if bc.VerifyTransaction(tx) {
			validTXs = append(validTXs, tx)
		} else {
			fmt.Printf("这是一条无效的交易，校验失败!\n")
		}
	}

	//根据数组的下标找到最后一个区块，获取前区块哈希值
	//创建新的区块，并且添加到区块链
	//最后一个区块的哈希值,也就是新区块的前哈希值
	prevBlockHash := bc.lastHash

	// 更新数据库
	//1. 找到bucket
	//2. 判断有没有，
	//   有，写入数据
	//更新区块数据
	//更新lastHashKey对应的值
	//   没有， 直接报错退出

	bc.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			fmt.Printf("添加区块时，bucket不应为空，请检查!")
			os.Exit(1)
		}

		newBlock := NewBlock(validTXs, prevBlockHash)

		//更新数据库
		bucket.Put(newBlock.Hash, newBlock.Serialize())
		bucket.Put([]byte(lastHashKey), newBlock.Hash)

		//更新内存
		bc.lastHash = newBlock.Hash
		return nil
	})
}

//COPY 使用bolt自带迭代器，按照key-byte 进行排序，而非插入的顺序
func (bc *BlockChain) Printchain1() {

	bc.Db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(blockBucket))

		//从第一个key-> value 进行遍历，到最后一个固定的key时直接返回
		b.ForEach(func(k, v []byte) error {
			fmt.Printf("key : %x\n", k)
			return nil
		})
		return nil
	})
}

func (bc *BlockChain) PrintChain() {
	it := bc.NewIterator()

	for ; ; {

		block := it.Next()

		fmt.Printf("===============================\n")
		fmt.Printf("Version :%d\n", block.Version)
		fmt.Printf("PrevBlockHash :%x\n", block.PrevBlockHash)
		fmt.Printf("MerkeRoot :%x\n", block.MerkelRoot)
		timeFormat := time.Unix(int64(block.TimeStamp), 0).Format("2006-01-02 15:04:05")
		fmt.Printf("TimeStamp: %s\n", timeFormat)
		//fmt.Printf("TimeStamp :%d\n", block.TimeStamp)
		fmt.Printf("Difficulty :%d\n", block.Difficulty)
		fmt.Printf("Nonce :%d\n", block.Nonce)
		fmt.Printf("Hash :%x\n", block.Hash)
		fmt.Printf("Data :%s\n", block.Transactions[0].TXInputs[0].Sig)
		pow := NewProofOfWork(*block)
		fmt.Printf("IsValid : %v\n\n", pow.IsValid())

		if len(block.PrevBlockHash) == 0 {
			fmt.Printf("打印结束!\n")
			break
		}
	}
}

//1.定义一个属于blockchain的迭代器，里面包含两个东西
//a. db : 迭代谁
//b. 哈希指针：一个会移动指针，总是会只想当前的区块

type Iterator struct {
	Db          *bolt.DB //来自于区块链
	currentHash []byte   //随着移动改变
}

//创建一个迭代器, 最初指向最后一个区块
func (bc *BlockChain) NewIterator() *Iterator {
	return &Iterator{Db: bc.Db, currentHash: bc.lastHash}
}

func (it *Iterator) Next() *Block {
	var block *Block
	it.Db.View(func(tx *bolt.Tx) error {
		//找到bucket
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			fmt.Printf("遍历区块时，bucket不应为空，请检查!")
			os.Exit(1)
		}

		//读取数据：currentHash
		blockTmp := bucket.Get(it.currentHash)
		block = Deserialize(blockTmp)

		//currentHash左移
		it.currentHash = block.PrevBlockHash

		return nil
	})

	return block
}

type UTXOInfo struct {
	//output， id， index

	TXID  []byte //为了给定位这个output便于转成input
	Index int64

	Output TXOutput //为了找到里面的余额
}

//1. 遍历整个账本
//2. 匹配自己的地址（自己能够解锁的utxo）
//3. 把所有utxo返回
//4. 遍历utxo，把所有的value加起来

func (bc *BlockChain) FindMyUtxos(pubKeyHash []byte) []UTXOInfo {
	fmt.Printf("FindMyUtxos\n")

	//1. 遍历账本
	//2. 遍历区块
	//3. 遍历交易
	//4. 遍历output
	//a. 找到和address地址相同output
	//b. 过滤掉已经消耗过的

	// 最终要返回的结构
	//var UTXOs []TXOutput
	var UTXOInfos []UTXOInfo
	spentUTXOs := make(map[string][]int64)

	it := bc.NewIterator()

	for {

		//2. 遍历区块
		block := it.Next()

		//3. 遍历交易
		for _, tx := range block.Transactions {

		OUTPUT:
		//4. 遍历output
			for i /*0, 1, 2, 3*/ , output := range tx.TXOutputs {
				fmt.Printf("当前索引为：%d\n", i)

				if bytes.Equal(pubKeyHash, output.PubKeyHash) {

					key := string(tx.TXID)
					//检查一下这个output是否已经被用过了
					if len(spentUTXOs[key]) /*[]int64*/ != 0 {
						fmt.Printf("当前交易里面有%x消耗过的output\n", pubKeyHash)
						//spentUTXOs[0x222] = []int64{0}
						//spentUTXOs[0x333] = []int64{0} //中间值
						//spentUTXOs[0x333] = []int64{0, 1}
						for _, j /*0, 1*/ := range spentUTXOs[key] {
							if int64(i) == j {
								fmt.Printf("i==j,这个output被消耗了，不统计\n")
								continue OUTPUT
							}
						}
					}

					fmt.Printf("找到了一个属于%x的output\n", pubKeyHash)
					//UTXOs = append(UTXOs, output)
					utxoinfo := UTXOInfo{[]byte(key), int64(i), output}
					UTXOInfos = append(UTXOInfos, utxoinfo)
				}
			}

			//遍历input，找到这个address已经消耗过得output，标识出来, 在遍历output前检测，过滤
			//333 -> 0, 1
			//222 -> 0
			//key-> 所在交易的哈希值, value-> 是引用索引的数组
			//spentUTXOs := make(map[string][]int64)

			if tx.IsCoinbaseTx() == false {
				for _, input := range tx.TXInputs {

					//比较一下当前花费的utxo是否是自己的
					if bytes.Equal(HashPubKey(input.PubKey), pubKeyHash) {

						fmt.Printf("%x 已经消耗的output : %d\n", pubKeyHash, input.Index)
						key := string(input.TxId)
						spentUTXOs[key] /*[]int64*/ = append(spentUTXOs[key], input.Index)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXOInfos
}

//获取余额函数
func (bc *BlockChain) GetBalance(address string) {
	decodeInfo := base58.Decode(address)

	pubkeyHash := decodeInfo[1:len(decodeInfo)-4]

	//找到所有属于address的utxo的数组
	utxoinfos := bc.FindMyUtxos(pubkeyHash)

	//总额
	total := 0.0

	for _, utxoinfo := range utxoinfos {
		total += utxoinfo.Output.Value
	}

	fmt.Printf("%s 的余额为: %f\n", address, total)
}

func (bc *BlockChain) FindNeedUTXOs(pubKeyHash []byte, amount float64) (map[string][]int64, float64) {
	fmt.Printf("FindNeedUTXOs\n")

	needUTXOs := make(map[string][]int64)
	calc := 0.0 //10 + 2 + 3

	//这个过程类似于FindMyUtxos
	//只不过不需要全部返回，找到满足金额的utxo我就直接退出

	//1. 调用FindMyUtxos返回所有的UTXOInfos， 拿到所有的钱I
	//2. 从这个返回值中挑取所需要的utxo，直接返回即可

	utxoinfos := bc.FindMyUtxos(pubKeyHash)

	for _, utxoinfo := range utxoinfos {

		calc += utxoinfo.Output.Value
		key := string(utxoinfo.TXID)
		needUTXOs[key] = append(needUTXOs[key], utxoinfo.Index)

		if calc >= amount {
			return needUTXOs, calc
		}
	}
	return needUTXOs, calc
}

func (bc *BlockChain) FindTransctionById(txid []byte) *Transaction {

	//1. 遍历区块链
	//2. 遍历交易
	//3. 比较交易id与txid，相同则返回， 没找到，最后返回nil

	it := bc.NewIterator()
	for {
		block := it.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.TXID, txid) {
				fmt.Printf("找到引用的交易!\n")
				return tx
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return nil
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privateKey *ecdsa.PrivateKey) bool {
	//inputs所有引用的交易
	prevTXs := make(map[string]*Transaction)

	//1. 遍历inputs
	//2. 根据每一个input的id找到交易本身
	//3. 把交易存储到prevTXs

	for _, input := range tx.TXInputs {
		//指定id找到交易本身
		tx := bc.FindTransctionById(input.TxId)

		if tx == nil {
			return false
		}
		prevTXs[string(input.TxId)] = tx
	}

	//签名
	return tx.Sign(privateKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {

	if tx.IsCoinbaseTx() {
		return true
	}

	fmt.Printf("VerifyTransaction...\n")
	prevTXs := make(map[string]*Transaction)

	for _, input := range tx.TXInputs {

		tx := bc.FindTransctionById(input.TxId)

		if tx == nil {
			return false
		}

		prevTXs[string(input.TxId)] = tx
	}

	return tx.Verify(prevTXs)
}
