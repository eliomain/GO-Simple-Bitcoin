package main

import (
	"./base58"
	"./bolt"
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"log"
	"os"
)

const blockChainName = "blockChain.db"
const blockBucketName = "blockBucket"
const lastHashKey = "lastHashKey"

//创建区块链
type BlockChain struct {
	//Blocks []*Block /*改写*/
	db *bolt.DB
	tail []byte //最后一个区块哈希值
}

//获取区块链实例函数
func newBlockchain() *BlockChain {
	if !IsFileExist(blockChainName){
		fmt.Printf("区块链不存在，请先创建\n")
		return nil
	}

	//打开数据库
	db, err := bolt.Open(blockChainName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close() 返回db指针 所以先不要关闭

	var tail []byte

	db.View(func(tx *bolt.Tx) error {
		//打开桶
		b := tx.Bucket([]byte(blockBucketName))
		if b == nil{
			fmt.Printf("区块链bucket为空，请检查\n")
			os.Exit(1)
		}

		tail = b.Get([]byte(lastHashKey))
		return nil
	})

	return &BlockChain{db,tail}
}

//创建区块链函数
func CreateBlockChain(miner string) *BlockChain {
	if IsFileExist(blockChainName){
		fmt.Printf("区块链已经存在，不需要重复创建!\n")
		return nil
	}
	//打开数据库
	db, err := bolt.Open(blockChainName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close() 返回dbstr指针 所以先不要关闭

	var tail []byte

	db.Update(func(tx *bolt.Tx) error {
		//打开桶
		b, err := tx.CreateBucket([]byte(blockBucketName))

		if err != nil{
			log.Panic(err)
		}
		//bucket已经创建完成，准备写入数据
		//创世块中只有一个挖矿交易，只有Coinbase
		coinbase := NewCoinbaseTx(miner,firstData)
		firstBlock := newBlock([]*Transaction{coinbase},[]byte{})
		b.Put(firstBlock.Hash, firstBlock.Serialize() /*将区块序列化，转成字节流*/)
		b.Put([]byte(lastHashKey), firstBlock.Hash)
		tail = firstBlock.Hash
		return nil
	})

	return &BlockChain{db,tail}
}



//添加区块
func (bc *BlockChain) addBlock (txs []*Transaction){
	//矿工得到交易时，第一时间对交易进行验证
	//矿工如果不验证 即使挖矿城关，广播区块后，其他的验证矿工，仍然会校验每一笔交易
	validTXs := []*Transaction{}

	for _,tx := range txs{
		if bc.VerifyTransaction(tx){
			fmt.Printf("--- 该交易有效: %x\n", tx.TXid)
			validTXs = append(validTXs,tx)
		}else{
			fmt.Printf("发现无效的交易: %x\n", tx.TXid)
		}
	}
	////获取前区块哈希
	//lastBc := bc.Blocks[len(bc.Blocks)-1]
	////添加区块
	//bc.Blocks = append(bc.Blocks,newBlock(data,lastBc.Hash))
	bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucketName))
		if b == nil{
			fmt.Printf("bucket不存在，请检查\n")
			os.Exit(1)
		}
		block := newBlock(txs,bc.tail)
		b.Put(block.Hash,block.Serialize()) //将区块序列化，转成字节流
		b.Put([]byte(lastHashKey), block.Hash)
		bc.tail = block.Hash
		return nil
	})
}

//定义一个区块链的迭代器，包含db,current
type BlockChainIterator struct {
	db *bolt.DB
	current []byte //当前所指向区块的哈希值
}

//创建迭代器，使用bc进行初始化
func (bc *BlockChain) NewIterator() *BlockChainIterator  {
	return &BlockChainIterator{bc.db,bc.tail}
}

func (it *BlockChainIterator) Next() *Block {
	var block Block

	it.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucketName))
		if b == nil{
			fmt.Printf("bucket不存在，请检查\n")
			os.Exit(1)
		}
		//读取当前current数据
		blockInfo := b.Get(it.current) //字节流
		block = *DeSerialize(blockInfo)
		//赋值前哈希 current向前移动一个区块
		it.current = block.PrevHash
		return nil
	})

	return &block
}

//我们想把FindMyUtxos和FindNeedUTXO进行整合
//1、FindMyUtxos:找到所有utxo（只要output就可以了）
//2、FindNeedUTXO:找到需要的utxo（要output的定位）
//我们可以定义一个结构，同时包含output已经定位信息

type UTXOInfo struct {
	TXID []byte //交易id
	Index int64 //output的索引值
	Output TXOutput //output本身
}

//遍历交易输出
func (bc *BlockChain) FindMyUtxos(pubKeyHash []byte) []UTXOInfo {
	//var myOutput []TXOutput
	var UTXOInfos []UTXOInfo //新的返回结构
	it := bc.NewIterator()


	//这是标识已经消耗过的utxo结构key是交易ID，value是这个id里的output索引的数组
	spentUTXOs := make(map[string][]int64)

	for{ //遍历账本
		block := it.Next()
		for _,tx := range block.Transactions{ //遍历交易
			//遍历交易输入inputs
			if tx.IsCoinbase() == false{
				for _,input := range tx.TXInputs{
					//input中公钥本身要进行哈希 判断使用过的input是否为目标地址所有 （其实还是为了看看时不时一个人）
					if bytes.Equal(HashPubKey(input.PubKey),pubKeyHash){
						fmt.Printf("找到了消耗过的output! index:%d\n",input.Index)
						key := string(input.TXID) // 这里是每一个input里面的ID
						spentUTXOs[key] = append(spentUTXOs[key],input.Index)
					}
				}
			}

			//遍历output
			OUTPUT:
			for i,output := range tx.TXOutputs{
				//这里迭代器由后往前 所以已经花费过的金额 已经在新的区块捕捉
				key := string(tx.TXid) //这里是交易ID output中没有ID参数 只有外面的交易ID
				indexes /*[]int64{0,1}*/ := spentUTXOs[key]

				if len(indexes) != 0{ //花费过了 一次性会花费整个的交易
					fmt.Printf("当前这币交易中又被消耗过的output\n")
					for _,j /*0,1*/ := range indexes{
						if int64(i) == j{
							fmt.Printf("i == j ,当前的output已经被消耗过了，跳过不统计\n")
							continue OUTPUT //跳出两层循环
						}
					}
				}

				//找到属于我的所有output 前面已经排除了消耗过的 这里就剩下未花费的
				if bytes.Equal(pubKeyHash,output.PubKeyHash){
					//fmt.Printf("找到了属于 %s 的output，i：%d\n",address,i)
					//myOutput = append(myOutput, output)
					utxoinfo := UTXOInfo{tx.TXid,int64(i),output}
					UTXOInfos = append(UTXOInfos,utxoinfo)
				}
			}
		}

		//这里一定要跳出
		if len(block.PrevHash) == 0{
			fmt.Printf("遍历区块链结束\n")
			break
		}
	}

	return UTXOInfos
}

func (bc *BlockChain) GetBalance(address string)  {
	//传地址时候需要逆过程 找到公钥哈希 必须用地址反推出
	//这个过程，不要打开钱包，因为有可能查看余额的人不是地址本人
	decodeInfo := base58.Decode(address)
	pubKeyHash := decodeInfo[1:len(decodeInfo)-4]
	utxoinfos := bc.FindMyUtxos(pubKeyHash)

	var total = 0.0
	//所有的output都在utxoinfos内部 获取余额时，遍历utxoinfos获取output即可
	for _,utxoinfo := range utxoinfos{
		total += utxoinfo.Output.Value
	}

	fmt.Printf("%s 的余额为%f \n",address,total)
}

//1、遍历账本，找到属于付款人的合适金额，把这个outputs找到
//utxos,resValue = bc.FindNeedUtxos(from,amount)
func (bc *BlockChain)FindNeedUtxos(pubKeyHash []byte,amount float64) (map[string][]int64,float64) {
	needUtxos := make(map[string][]int64) //标识能用的utxo
	var resValue float64 //返回统计的金额

	//复用FindMyUtxo函数，这个函数已经包含了所有信息
	utxoinfos := bc.FindMyUtxos(pubKeyHash)

	for _,utxoinfo := range utxoinfos{
		key := string(utxoinfo.TXID)
		needUtxos[key] = append(needUtxos[key],int64(utxoinfo.Index)) //这里定位了UTXO->TXID和索引
		resValue += utxoinfo.Output.Value

		//判断一下金额是否足够
		if resValue >= amount{
			//足够就跳出
			break
		}
	}
	return needUtxos,resValue
}

func (bc *BlockChain) SignTransaction(tx *Transaction,privateKey *ecdsa.PrivateKey)  {
	//1、遍历账本找到所有应用交易
	prevTXs := make(map[string]Transaction)

	//遍历tx的inputs，通过id去查找所引用的交易
	for _,input := range tx.TXInputs{
		prevTx := bc.FindTransaction(input.TXID) //传入本次转账交易的 Input的TXID查找Transaction

		if prevTx == nil{
			fmt.Printf("没用找到交易：%x\n",input.TXID)
		}else{
			//把找到的引用交易保存起来
			//0x222
			//0x333
			prevTXs[string(input.TXID)] = *prevTx // TXID -> Transaction
		}
	}

	tx.Sign(privateKey,prevTXs)
}

//矿工校验流程 （要丢到网络中 矿工挖矿校验）
//1、找到交易input所引用的所有的交易prevTXs
//2、对交易进行校验
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	//校验的时候，如果是挖矿交易，直接返回true
	if tx.IsCoinbase(){
		return true
	}

	prevTXs := make(map[string]Transaction)

	//遍历tx的inputs，通过id去查找所引用的交易
	for _,input := range tx.TXInputs{
		prevTx := bc.FindTransaction(input.TXID) //传入本次转账交易的 Input的TXID查找Transaction

		if prevTx == nil{
			fmt.Printf("没用找到交易：%x\n",input.TXID)
		}else{
			//把找到的引用交易保存起来
			//0x222
			//0x333
			prevTXs[string(input.TXID)] = *prevTx // TXID -> Transaction
		}
	}

	return tx.Verify(prevTXs)
}

//因为在本程序中内存中没用记录 交易ID和交易索引 所以每次都需要遍历
func (bc *BlockChain) FindTransaction(txid []byte) *Transaction {
	//遍历区块链的交易
	//通过对比ID来识别
	it := bc.NewIterator()

	for{
		block := it.Next()

		for _,tx := range block.Transactions{

			//如果找到相同ID交易，直接返回交易即可
			if bytes.Equal(tx.TXid,txid){
				fmt.Printf("找到了所引用交易：%x\n",tx.TXid)
				return tx
			}
		}

		if len(block.PrevHash) == 0{
			break
		}
	}

	return nil
}