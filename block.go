package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)


//1. 定义结构（区块头的字段比正常的少）
//>1. 前区块哈希
//>2. 当前区块哈希
//>3. 数据

//2. 创建区块
//3. 生成哈希
//4. 引入区块链
//5. 添加区块
//6. 重构代码

const firstData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type Block struct {
	Version uint64 //版本号
	PrevHash []byte //前区块哈希
	MarKleRoot []byte //梅克尔根
	TimeStamp uint64
	Difficult uint64 //难度值
	Nonce uint64 //随机数
	//Data []byte //数据，目前使用字节流，v4开始使用交易代替
	Transactions []*Transaction
	Hash []byte //当前区块哈希
}

func newBlock(txs []*Transaction,prevHash []byte) *Block {
	block := Block{
		Version: 00,
		PrevHash: prevHash,
		MarKleRoot: []byte{},
		TimeStamp: uint64(time.Now().Unix()),
		Difficult: Bits,
		Nonce: 10,
		Transactions: txs,
		Hash: []byte{}, //先填充为空，后续会填充数据
	}

	//生成梅克尔根
	block.HashTransactions()

	//block.setHash()
	pow := NewProofOfWork(&block)
	hash,nonce := pow.Run()

	block.Hash = hash
	block.Nonce = nonce

	return &block
}

//模拟梅克尔根，简单处理
func (block *Block) HashTransactions() {
	//将交易的ID拼接起来
	var hashes []byte
	for _,tx := range block.Transactions{
		txid := tx.TXid
		hashes = append(hashes,txid...)
	}

	hash := sha256.Sum256(hashes)
	block.MarKleRoot = hash[:]
}

//序列化 将区块转换为字节流
func (block *Block) Serialize() []byte {
	//fmt.Printf("编码开始\n")
	//编码
	var buffer bytes.Buffer
	//定义编码器
	buf := gob.NewEncoder(&buffer)
	//编码器对结构进行编码，一定要进行校验
	err := buf.Encode(block)
	if err != nil{
		log.Panic(err)
	}
	return buffer.Bytes()
}

func DeSerialize(data []byte) *Block {
	//fmt.Printf("解码开始\n")
	var block Block
	//创建解码器
	De := gob.NewDecoder(bytes.NewReader(data))
	err := De.Decode(&block)
	if err != nil{
		log.Panic(err)
	}
	return &block
}
