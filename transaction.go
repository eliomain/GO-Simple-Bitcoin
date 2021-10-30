package main

import (
	"./base58"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	"strings"
)

//定义交易结构
//定义input
//定义output
//设置交易ID

type TXInput struct {
	TXID []byte //是引用output上一个区块的哈希 用于寻找上一个区块交易
	Index int64 //output的索引
	//Address string //解锁脚本，先使用地址模拟

	Signature []byte //交易签名
	PubKey []byte //公钥本身，不是公钥哈希
}

type TXOutput struct {
	Value float64 //转账金额
	//Address string //锁定脚本

	PubKeyHash []byte //是公钥的哈希，不是公钥本身
}

type Transaction struct {
	TXid []byte //交易id
	TXInputs []TXInput //所有的inputs
	TXOutputs []TXOutput //所有的outputs
}

//setTXID函数
func (tx *Transaction) setTXID() {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(tx)
	if err != nil{
		log.Panic(err)
	}
	hash := sha256.Sum256(buffer.Bytes())
	tx.TXid = hash[:]
}

//TXOutput提供Lock方法 给定地址，得到地址的公钥哈希，锁定这个output
func (output *TXOutput) Lock(address string)  {
	//address -> public key hash
	//25字节
	decodeInfo := base58.Decode(address)

	pubKeyHash := decodeInfo[1:len(decodeInfo)-4]

	output.PubKeyHash = pubKeyHash
}

//创建TXOutput
//由于创建过程需要调用Lock方法，所以我们自顶一个NewTXOutput方法
func NewTXOutput(value float64,address string) TXOutput {
	output := TXOutput{Value: value}

	output.Lock(address)

	return output
}

const reward = 12.5

//实现挖矿交易
//特点：只有输出，没有有效的输入（不需要ID索引签名）
func NewCoinbaseTx(miner string,data string) *Transaction {
	//挖矿交易没输入
	inputs := []TXInput{TXInput{nil,-1,nil,[]byte(data)}}

	//输出给矿工奖励
	output := NewTXOutput(reward,miner)
	outputs := []TXOutput{output}

	tx := Transaction{nil,inputs,outputs}
	tx.setTXID()

	return &tx
}

func (tx *Transaction) IsCoinbase() bool {
	//挖矿交易判断
	//特点1、只有一个input 2、id是nil 3、索引是-1
	inputs := tx.TXInputs
	if len(inputs) == 1 && inputs[0].TXID == nil && inputs[0].Index == -1{
		return true
	}
	return false
}

//创建普通交易
func NewTransaction(from,to string,amount float64,bc *BlockChain) *Transaction{
	//1、打开钱包
	ws := NewWallets()
	//获取密钥对
	wallet := ws.WalletsMap[from] // 传入我的地址 获取我的公钥私钥

	if wallet == nil{
		fmt.Printf("%s 的私钥不存在，交易创建失败\n",from)
	}

	//获取公钥、私钥
	privateKey := wallet.PrivateKey //目前使用不到，步骤三签名时需要
	publicKey := wallet.PublicKey
	pubKeyHash := HashPubKey(wallet.PublicKey)

	utxos := make(map[string][]int64) //标识能用的utxo
	var resValue float64 //这些utxo存储的金额
	//例如李四转赵六4，返回的信息为
	//utxos[0x333] = int64{0,1}
	//resValue = 5

	//1、遍历账本，找到属于付款人的合适金额，把这个outputs找到
	utxos,resValue = bc.FindNeedUtxos(pubKeyHash,amount)

	//2、如果找到钱不足以转账，创建交易失败
	if resValue < amount{
		fmt.Printf("余额不足，交易失败\n")
		return nil
	}

	var inputs []TXInput
	var outputs []TXOutput

	//3.将Outputs转成inputs
	for txid,indexes := range utxos{
		for _,i := range indexes{
			//将outputID给新的input
			input := TXInput{[]byte(txid),i,nil,publicKey}
			inputs = append(inputs,input)
		}
	}

	//4、创建输出，创建一个属于收款人的output
	//output := TXOutput{amount,to}
	output := NewTXOutput(amount,to)
	outputs = append(outputs,output)

	//5、如果有找零，创建属于付款人output
	if resValue > amount{
		//output1 := TXOutput{resValue-amount,from}
		output1 := NewTXOutput(resValue-amount,from)
		outputs = append(outputs,output1)
	}

	tx := Transaction{nil,inputs,outputs}
	//6、设置交易ID
	tx.setTXID()

	//把查找引用交易的环节放到BlockChain中去，同时在BlockChain进行调用签名

	//我们付款人在创建交易时，已经得到了所有引用的output详细信息
	//但是我们不去使用，因为矿工在校验的时候，矿工是没用这部分信息的，矿工需要遍历账本找到所有引用交易
	bc.SignTransaction(&tx,privateKey) //将转账这笔交易传进去

	//7、返回交易结构
	return &tx
}

//第一个参数是私钥
//第二个参数是这个交易的input所引用的所有的交易
func (tx *Transaction) Sign(privKey *ecdsa.PrivateKey,prevTXs map[string]Transaction) {
	fmt.Printf("对交易进行签名\n")

	//1、拷贝一份交易txCopy
	// > 做相应裁剪：把每一个input的sig和pubkey设置为nil
	// > output不做改变

	txCopy := tx.TrimmedCopy()
	//2、遍历txCopy.inputs
	//把这个Input所引用的output的公钥哈希拿过来，赋值给pubkey
	for i,input := range txCopy.TXInputs {
		//找到引用的交易
		preTX := prevTXs[string(input.TXID)]
		output := preTX.TXOutputs[input.Index]

		//for循环迭代出来的数据是一个副本，对这个Input进行修改不会影响到原始数据
		//所以我们需要下标方式修改
		txCopy.TXInputs[i].PubKey = output.PubKeyHash

		//签名要对数据的hash进行签名
		//我们的数据都在交易中，我们要求交易的哈希
		//Transaction的SetTXID函数就是对交易的哈希
		//所以我们可以使用交易id作为我们的签名的内容

		//3、生成要签名的数据（哈希）
		txCopy.setTXID()
		signData := txCopy.TXid

		//清理
		txCopy.TXInputs[i].PubKey = nil

		fmt.Printf("要签名的数据：signData:%x\n",signData)

		//4、对数据进行签名r,s
		r,s,err := ecdsa.Sign(rand.Reader,privKey,signData)

		if err != nil{
			fmt.Printf("交易签名失败，err：%v\n",err)
		}

		//5、拼接r,s为字节流
		signature := append(r.Bytes(),s.Bytes()...)

		//6、赋值给原始的交易的Signature字段
		tx.TXInputs[i].Signature = signature
	}


	//5、拼接r,s为字节流，赋值给原始的交易的Signature字段
}

//trim：裁剪
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _,input := range tx.TXInputs{
		input1 := TXInput{input.TXID,input.Index,nil,nil}
		inputs = append(inputs,input1)
	}

	outputs = tx.TXOutputs

	tx1 := Transaction{tx.TXid,inputs,outputs}
	return tx1
}

func (tx *Transaction) Verify (prevTXs map[string]Transaction) bool {
	fmt.Printf("对交易进行校验...\n")

	//1、拷贝修剪的副本
	txCopy := tx.TrimmedCopy()

	//2、遍历原始交易（注意，不是txCopy）
	for i,input := range tx.TXInputs{

		//3、遍历原始交易的input所引用的前交易prevTX
		prevTX := prevTXs[string(input.TXID)]
		output := prevTX.TXOutputs[input.Index]

		//4、找到output的公钥哈希，赋值给txCopy对应的Input
		txCopy.TXInputs[i].PubKey = output.PubKeyHash

		//5、还原签名的数据
		txCopy.setTXID()

		//清理动作，重要!!!
		txCopy.TXInputs[i].PubKey = nil

		verifyData := txCopy.TXid
		fmt.Printf("verifyData:%x\n",verifyData)

		//6、校验
		//还原签名为r,s
		signature := input.Signature

		//公钥字节流
		pubKeyBytes := input.PubKey

		r := big.Int{}
		s := big.Int{}

		rData := signature[:len(signature)/2]
		sData := signature[len(signature)/2:]

		r.SetBytes(rData)
		s.SetBytes(sData)

		//type PublicKey struct {
		//	elliptic.Curve
		//	X, Y *big.Int
		//}

		//还原公钥为curve,x,y
		x := big.Int{}
		y := big.Int{}

		xData := pubKeyBytes[:len(pubKeyBytes)/2]
		yData := pubKeyBytes[len(pubKeyBytes)/2:]

		x.SetBytes(xData)
		y.SetBytes(yData)

		curve := elliptic.P256()

		publicKey := ecdsa.PublicKey{curve,&x,&y}

		//数据、签名、公钥 准备完毕，开始校验
		//func Verify(pub *PublicKey, hash []byte, r, s *big.Int) bool {

		if !ecdsa.Verify(&publicKey,verifyData,&r,&s){ //公钥 数据 r、s签名
			return false
		}
	}
	return true
}

//交易的String函数
func (tx *Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.TXid))

	for i, input := range tx.TXInputs {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.TXID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Index))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.TXOutputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %f", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	//11111, 2222, 3333, 44444, 5555

	//`11111
	//2222
	//3333
	//44444
	//5555`

	return strings.Join(lines, "\n")
}