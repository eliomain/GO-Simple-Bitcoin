package main

import (
	"bytes"
	"fmt"
	"time"
)

//添加区块链
func (cli *CLI) addBlock (txs []*Transaction) {
	bc := newBlockchain()
	if bc == nil {
		return
	}
	defer bc.db.Close()

	bc.addBlock(txs)
}

func (cli *CLI) CreateBlockChain(addr string)  {
	if !IsValidAddress(addr){
		fmt.Printf("%s 是无效地址\n",addr)
		return
	}

	bc := CreateBlockChain(addr)
	if bc == nil {
		return
	}

	defer bc.db.Close()
	fmt.Printf("创建区块链成功!\n")
}

func (cli *CLI) GetBalance (addr string){
	if !IsValidAddress(addr){
		fmt.Printf("%s 是无效地址\n",addr)
		return
	}

	bc := newBlockchain() //bc从这里实例化了
	if bc == nil {
		return
	}
	defer bc.db.Close()
	bc.GetBalance(addr)
}

//查看区块链
func (cli *CLI) printChain() {
	bc := newBlockchain()
	if bc == nil {
		return
	}
	defer bc.db.Close()
	//使用迭代器
	it := bc.NewIterator()
	for {
		block := it.Next()
		fmt.Printf("+++++++++++++++++++++++++++++++++++++++\n")

		fmt.Printf("Version ：%d\n", block.Version)
		fmt.Printf("PrevBlockHash ：%x\n", block.PrevHash)
		fmt.Printf("MerKleRoot ：%x\n", block.MarKleRoot)
		timeFormat := time.Unix(int64(block.TimeStamp), 0).Format("2006-01-02 15:04:05")
		fmt.Printf("TimeStamp ：%s\n", timeFormat)
		fmt.Printf("Difficulity ：%d\n", block.Difficult)
		fmt.Printf("Nonce ：%d\n", block.Nonce)
		fmt.Printf("Hash ：%x\n", block.Hash)
		//创世块没有input 本程序设置Address就是第一个data
		fmt.Printf("Data ：%s\n", block.Transactions[0].TXInputs[0].PubKey)

		//校验区块
		pow := NewProofOfWork(block)
		fmt.Printf("IsValid ：%v\n", pow.IsValid())

		//检测到创始块结束
		if bytes.Equal(block.PrevHash, []byte{}) {
			fmt.Printf("区块链遍历结束\n")
			break
		}
	}
}

//发送交易cli.Send(from,to,amount,miner,data)
func (cli *CLI) Send (from,to string,amount float64,miner string,data string)  {

	if !IsValidAddress(from){
		fmt.Printf("%s from是无效地址\n",from)
		return
	}
	if !IsValidAddress(to){
		fmt.Printf("%s to是无效地址\n",to)
		return
	}
	if !IsValidAddress(miner){
		fmt.Printf("%s miner是无效地址\n",miner)
		return
	}

	bc := newBlockchain()
	if bc == nil {
		return
	}
	defer bc.db.Close()

	//1、创建挖矿交易
	coinbase := NewCoinbaseTx(miner,data)

	txs := []*Transaction{coinbase}

	//2、创建普通交易
	tx := NewTransaction(from,to,amount,bc)
	if tx != nil{
		txs = append(txs, tx)
	}else{
		fmt.Printf("发现无效交易，过滤！\n")
	}

	//3、添加到区块
	bc.addBlock(txs)

	fmt.Printf("挖矿成功!")
}

func (CLI *CLI) createWallet() {
	ws := NewWallets()
	address := ws.CreateWallet()
	fmt.Printf("新的钱包地址为：%s\n",address)
}

func (cli *CLI) ListAddresses()  {
	ws := NewWallets()

	addresses := ws.ListAddress()
	for _,address := range addresses{
		fmt.Printf("address: %s\n",address)
	}
}

func (cli *CLI) PrintTx()  {
	bc := newBlockchain()
	if bc == nil{
		return
	}

	defer bc.db.Close()

	it := bc.NewIterator()

	for{
		block := it.Next()
		fmt.Printf("\n+++++++++++++++ 新的区块 +++++++++++++++++\n")
		for _,tx := range block.Transactions{
			fmt.Printf("tx : %v\n", tx)
		}

		if len(block.PrevHash) == 0{
			break
		}
	}
}

