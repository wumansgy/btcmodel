package main

import (
	"fmt"
	"os"
)

func (cli *CLI) CreateBlockChain(address string) {
	if !IsValidAddress(address) {
		fmt.Printf("地址无效，请检查\n")
		return
	}

	//3. 调用真正的添加区块函数
	bc := CreateBlockChain(address)
	defer bc.Db.Close()
}

func (cli *CLI) PrintChain() {
	bc := NewBlockChain()
	//defer bc.Db.Close()
	bc.PrintChain()
}

func (cli *CLI) PrintTx() {
	bc := NewBlockChain()
	//defer bc.Db.Close()
	//bc.PrintChain()
	it := bc.NewIterator()
	for {
		block := it.Next()
		fmt.Printf("++++++++++++++++++++++++++++++++++\n")
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) GetBalance(address string) {
	if !IsValidAddress(address) {
		fmt.Printf("地址无效，请检查\n")
		return
	}

	bc := NewBlockChain()
	bc.GetBalance(address)
}

func (cli *CLI) Send(from, to string, amount float64) {

	if !IsValidAddress(from) {
		fmt.Printf("from : %s 地址无效，请检查\n", from)
		return
	}

	if !IsValidAddress(to) {
		fmt.Printf("to : %s 地址无效，请检查\n", to)
		return
	}

	fmt.Printf("%s 向 %s 转账 %f\n", from, to, amount)

	bc := NewBlockChain()
	if bc == nil {
		return
	}
	//defer bc.Db.Close()

	//txs := []*Transaction{coinbase}

	//2. 创建普通交易
	tx := NewTransaction(from, to, amount, bc)
	if tx != nil {
		//3. 添加交易到交易数组
		gTx = append(gTx, tx)
	} else {
		fmt.Printf("余额不足，创建交易失败\n")
	}
}

func (cli *CLI) Mine(miner, data string) {
	//1. 创建挖矿交易
	coinbase := NewCoinbaseTx(miner, data)
	//先放着最后面
	gTx = append(gTx, coinbase)

	bc := NewBlockChain()
	if bc == nil {
		return
	}
	//defer bc.Db.Close()

	//4. 添加交易到区块链AddBlock
	bc.AddBlock(gTx)

	//应该清楚全局gTx
	gTx = []*Transaction{}

	fmt.Printf("区块创建成功!\n")
	//./blockchian send FROM TO AMOUNT MINER DATA  "转账"
}

func (cli *CLI) CreateWallet() {
	//w := NewWallet()
	//address := w.GetAddress()
	ws := NewWallets()

	address := ws.CreateWallet()

	if address == "" {
		fmt.Printf("创建地址失败\n")
		os.Exit(1)
	}

	fmt.Printf("你的新地址为: %s\n", address)
}

func (cli *CLI) ListAllAddress() {
	//1. 打开钱包
	//2. 遍历里面的map
	//3. 将所有的key（地址）返回来

	ws := NewWallets()
	//返回一个地址的数组
	addresses := ws.GetAddresses()

	for _, address := range addresses {
		fmt.Printf("%s\n", address)
	}
}

func (cli *CLI) Help() {
	fmt.Printf(Usage)
}

func (cli *CLI)Status()  {
	fmt.Printf("待确认交易数量: %d\n", len(gTx))
}
