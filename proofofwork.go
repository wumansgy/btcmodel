package main

import (
	"math/big"
	"crypto/sha256"
	"bytes"
)

//系统调节的难度值
const difficulty  = 16

//1. 定义工作量证明， block， 难度值
type ProofOfWork struct {
	//数据来源
	block Block
	//难度值
	target *big.Int //一个能够处理大数的内置的类型，有比较方法
}

//2. 创建一个工作量证明

func NewProofOfWork(block Block) *ProofOfWork {
	//填充两个字段

	pow := ProofOfWork{
		block: block,
	}
	//使用全局的难度值，推导目标哈希值 , 前面5个0，数字num = 5， 000001000000xxxxxxx
	//构造这个目标值
	// 0000100000000000000000000000000000000000000000000000000000000000

	//步骤
	//初始值1
	// 0000000000000000000000000000000000000000000000000000000000000001

	//整体向左移动256
	//10000000000000000000000000000000000000000000000000000000000000000

	//整体向右移动5个位(20次)
	// 0000100000000000000000000000000000000000000000000000000000000000

	////引用一个临时的big.int来接这个target
	//var targetInt big.Int
	//targetInt.SetString(targetStr, 16)

	//初始值为1
	targetInt := big.NewInt(1)
	////left shift 左移256
	//targetInt.Lsh(targetInt, 256)
	////right shift 右移20
	//targetInt.Rsh(targetInt, difficulty)

	//终极版
	targetInt.Lsh(targetInt, 256 - difficulty)

	pow.target = targetInt
	return &pow
}

func (pow *ProofOfWork) prepareData(nonce uint64) []byte {
	block := pow.block
	//使用Join代替append
	bytesArray := [][]byte{
		Uint2Byte(block.Version),
		block.PrevBlockHash,
		block.MerkelRoot,
		Uint2Byte(block.TimeStamp),
		Uint2Byte(block.Difficulty),
		Uint2Byte(nonce),
		block.Data,
	}

	info := bytes.Join(bytesArray, []byte{})
	return info
}

//3. 实现挖矿计算满足条件哈希值的函数 nonce
//计算完之后返回当前区块的哈希，和nonce
func (pow *ProofOfWork) Run() ([]byte, uint64) {
	//1. 拿到区块数据
	//block := pow.block
	//区块的哈希值
	var currentHash [32]byte
	//挖矿的随机值
	var nonce uint64

	for {
		info := pow.prepareData(nonce)
		//2. 对数据做哈希运算
		currentHash = sha256.Sum256(info)

		//3. 比较
		//引用big.int，将获取的[]byte类型的哈希值转成big.int
		var currentHashInt big.Int
		currentHashInt.SetBytes(currentHash[:])

		//   -1 if x <  y
		//    0 if x == y
		//   +1 if x >  y
		//
		//func (x *Int) Cmp(y *Int) (r int) {
		if currentHashInt.Cmp(pow.target) == -1 {
			//a. 比目标小，成功，返回哈希和nonce
			break
		} else {
			//b. 比目标大,继续nonce++
			nonce++
		}
	}

	return currentHash[:], nonce
}

//4. 验证找到的nonce是否正确

//根据pow里的数据进行反向计算，验证一个下nonce和哈希是否匹配当前系统的难度需求
func (pow *ProofOfWork) IsValid() bool {
	//nonce已经在block中了，我们就是要校验它
	info := pow.prepareData(pow.block.Nonce)
	//做哈希运算
	hash := sha256.Sum256(info[:])

	//引入临时变量
	tmpInt := big.Int{}
	tmpInt.SetBytes(hash[:])
	// 比较

	return tmpInt.Cmp(pow.target) == -1
}
