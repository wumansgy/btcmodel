package main

import (
	"fmt"
	"strconv"
)

const Usage = `
Usage:
	createbc <ADDRESS> "创建区块链"
	print "打印区块链"
	printtx "打印交易"
	balc <ADDRESS> "获取指定地址余额"
	send <FROM> <TO> <AMOUNT> "转账"
	mine [MINER] [DATA] "挖矿"，默认:  
	createwt "创建钱包地址"
	list "打印钱包中的所有地址"
	status "查看当前待确认交易数量"
`

//定义一个CLI，所有细节工作交给它，命令的解析工作交给CLI
type CLI struct {
	//bc *BlockChain
}

func checkArgs(cmds []string, count int) bool {
	if len(cmds) != count {
		fmt.Println("参数无效!")
		//os.Exit(1)
		return false
	}
	return true
}

//定义一个run函数，负责接收命令行的数据，然后根据命令进行解析，并完成最终的调用
func (cli *CLI) Run(cmds []string) {
	//args := os.Args
	args := cmds

	//if len(args) < 2 {
	//	fmt.Println(Usage)
	//	//os.Exit(1)
	//	return
	//}

	cmd := args[0]

	switch cmd {

	case "createbc":
		fmt.Printf("创建区块链命令被调用!\n")
		if !checkArgs(cmds, 2) {
			return
		}
		address := args[1]
		cli.CreateBlockChain(address)

	case "print":
		fmt.Printf("打印区块命令被调用\n")
		if !checkArgs(cmds, 1) {
			return
		}
		cli.PrintChain()

	case "printtx":
		fmt.Printf("打印交易命令被调用\n")
		if !checkArgs(cmds, 1) {
			return
		}
		cli.PrintTx()

	case "balc":
		fmt.Printf("获取余额命令被调用\n")
		if !checkArgs(cmds, 2) {
			return
		}
		address := args[1]
		cli.GetBalance(address)

	case "send":
		fmt.Printf("转账send命令被调用\n")
		if !checkArgs(cmds, 4) {
			return
		}

		from := args[1]
		to := args[2]
		amount, _ := strconv.ParseFloat(args[3], 64) //string
		cli.Send(from, to, amount)

	case "mine":
		fmt.Printf("mine命令被调用\n")
		//if !checkArgs(cmds, 4) {
		//	return
		//}

		//mine address data
		var miner string
		var data string

		if len(cmds) == 3 {
			miner = cmds[1]
			data = cmds[2]
		} else {
			data = "奖励xxxx 50.0"
			miner = "1NVwrN4yZVV3hW1PkXCg38sGcsXMKcYaw7"
		}

		cli.Mine(miner, data)

	case "createwt":
		fmt.Printf("createWallet命令被调用\n")
		if !checkArgs(cmds, 1) {
			return
		}

		cli.CreateWallet()
	case "list":
		fmt.Printf("listAllAddress命令被调用\n")
		if !checkArgs(cmds, 1) {
			return
		}
		cli.ListAllAddress()
	case "status":
		cli.Status()

	default:
		fmt.Printf("无效的命令，请检查!\n")
		cli.Help()
	}
}
