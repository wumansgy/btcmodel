package main

import (
	"github.com/boltdb/bolt"
	"log"
	"fmt"
	"os"
	"time"
)

//定义一个区块链结构，使用bolt数据库进行保存
type BlockChain struct {
	//数据库的句柄
	Db *bolt.DB
	//最后一个区块的哈希值
	lastHash []byte    //在内存中的临时值，只保存最后一个区块哈希
}

const blockChainName = "blockChain.db"
const blockBucket = "blockBucket"
const lastHashKey = "lastHashKey"

//定义一个创建区块链的方法
//就是返回一个区块链的实例instance，已经存在直接返回，不存在，创建再返回

//创建一个新的区块链
func CreateBlockChain() *BlockChain {
	if isDbExist() {
		fmt.Printf("区块链已经存在!\n")
		os.Exit(1)
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
			genesisBlock := NewBlock(genesisInfo, []byte{})
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
func GetBlockChain() *BlockChain {
	//在创建区块链的时候，添加一个创世块genesisBlock
	//genesisBlock := NewBlock(genesisInfo, []byte{})
	//blockChain := BlockChain{blocks: []*Block{genesisBlock}}
	//return &blockChain

	if !isDbExist() {
		fmt.Printf("请先创建区块链!\n")
		os.Exit(1)
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

func (bc *BlockChain) AddBlock(data string) {
	//根据数组的下标找到最后一个区块，获取前区块哈希值
	//lastBlock := bc.blocks[len(bc.blocks)-1]
	//prevHash := lastBlock.Hash
	//
	////创建新的区块，并且添加到区块链
	//block := NewBlock(data, prevHash)
	//bc.blocks = append(bc.blocks, block)

	//最后一个区块的哈希值,也就是新区块的前哈希值
	prevBlockHash := bc.lastHash

	// 更新数据库
	//1. 找到bucket
	//2. 判断有没有，
	//   有，写入数据
	//更新区块数据
	//更新lastHashKey对应的值
	//   没有， 直接报错退出
	//

	bc.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			fmt.Printf("添加区块时，bucket不应为空，请检查!")
			os.Exit(1)
		}

		newBlock := NewBlock(data, prevBlockHash)

		//更新数据库
		bucket.Put(newBlock.Hash, newBlock.Serialize())
		bucket.Put([]byte(lastHashKey), newBlock.Hash)

		//更新内存
		bc.lastHash = newBlock.Hash
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
		fmt.Printf("Data :%s\n", block.Data)
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

//判断区块链文件是否存在
func isDbExist() bool {
	if _, err := os.Stat(blockChainName); os.IsNotExist(err) {
		return false
	}

	return true
}
