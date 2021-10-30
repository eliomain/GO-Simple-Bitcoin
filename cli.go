package main

import (
	"fmt"
	"os"
	"strconv"
)

const Usage = `
./blockchain createBlockChain Address 创建区块链
./blockchain printChain 打印区块链
./blockchain getBalance Address 查询余额
./blockchain send FROM TO AMOUNT MINER DATA 转账命令
./blockchain createWallet 创建钱包
./blockchain ListAddresses 打印钱包地址列表
./blockchain printTx 打印所有交易
`

type CLI struct {
	//bc *BlockChain //cli中不需要保存区块链实例了，所有的名字在自己调用之前，自己获取区块链实例
}

func (cli *CLI) Run() {
	cmds := os.Args

	//没有填写参数
	if len(cmds) < 2{
		fmt.Printf(Usage)
		os.Exit(1)
	}

	switch cmds[1] {
		case "createBlockChain":
			if len(cmds) != 3 {
				fmt.Printf(Usage)
				os.Exit(1)
			}

			fmt.Printf("创建区块链命令被调用!\n")

			addr := cmds[2]
			cli.CreateBlockChain(addr)
		case "printChain":
			cli.printChain()
		case "getBalance":
			fmt.Printf("获取余额命令被调用\n")
			cli.GetBalance(cmds[2])
		case "send":
			fmt.Printf("转账命令被调用\n")
			//./blockchain send FROM TO AMOUNT MINER DATA 转账命令
			if len(cmds) != 7{
				fmt.Printf("send命令发现无效参数，请检查\n")
				fmt.Printf(Usage)
				os.Exit(1)
			}
			from := cmds[2]
			to := cmds[3]
			amount,_ := strconv.ParseFloat(cmds[4],64)
			miner := cmds[5]
			data := cmds[6]
			cli.Send(from,to,amount,miner,data)
		case "createWallet":
			fmt.Printf("创建钱包命令被调用\n")
			cli.createWallet()
		case "ListAddresses":
			fmt.Printf("打印钱包地址命令被调用\n")
			cli.ListAddresses()
		case "printTx":
			fmt.Printf("打印交易命令被调用\n")
			cli.PrintTx()
		default:
			fmt.Printf("参数错误\n")
			fmt.Printf(Usage)
	}
}
