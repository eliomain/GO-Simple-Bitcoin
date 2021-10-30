package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

const Bits = 8 //小数位移动4位 4*4 2位 2*4

//POW
type ProofOfWork struct {
	block *Block
	target *big.Int
}

func NewProofOfWork(block *Block) *ProofOfWork {
	pow := ProofOfWork{
		block: block,
	}
	//定义难度值 难度值应该是推导出来的为了简化我们先固定 后续再推导
	//targetStr := "0010000000000000000000000000000000000000000000000000000000000000" //64个数

	// SetString : 把string转成big.int类型  SetBytes : 把bytes转成big.int类型 []byte
	//var tmp big.Int
	//tmp.SetString(targetStr,16)
	//pow.target = &tmp

	//程序推导难度值, 推导前导为3个难度值
	// 0001000000000000000000000000000000000000000000000000000000000000
	//初始化 64个数*4=256位
	//  0000000000000000000000000000000000000000000000000000000000000001
	//向左移动, 256位
	//1 0000000000000000000000000000000000000000000000000000000000000000
	//向右移动, 四次，一个16进制位代表4个2进制（f:1111）
	//向右移动16位
	//0 0001000000000000000000000000000000000000000000000000000000000000

	// 创建 0000000000000000000000000000000000000000000000000000000000000001
	bigTemp := big.NewInt(1)

	// 向左移动, 256位
	//1 0000000000000000000000000000000000000000000000000000000000000000
	// 再向右移动 向右移动位置*4位
	bigTemp.Lsh(bigTemp,256-Bits)

	pow.target = bigTemp

	return &pow
}

//这是pow的运算函数，为了获取挖矿的随机数 返回值：Hash,Nonce
func (pow *ProofOfWork) Run() ([]byte,uint64){
	//block := pow.block
	var nonce uint64 //默认为0
	var hash [32]byte
	for{
		fmt.Printf("Pow运算：%x\r",hash)
		hash = sha256.Sum256(pow.prepareData(nonce))

		var tmp big.Int
		tmp.SetBytes(hash[:])

		if tmp.Cmp(pow.target) == -1{
			fmt.Printf("挖矿成功 Nonce：%d Hash：%x\n",nonce,hash)
			break
		}else{
			nonce++
		}
	}

	return hash[:],nonce
}

func (pow *ProofOfWork) prepareData(Nonce uint64) []byte {
	block := pow.block
	tmp := [][]byte{
		uintToByte(block.Version),
		block.PrevHash,
		block.MarKleRoot,
		uintToByte(block.TimeStamp),
		uintToByte(block.Difficult),
		uintToByte(Nonce),
	}

	value := bytes.Join(tmp,[]byte{})
	return value
}

func (pow *ProofOfWork) IsValid() bool{
	//校验工作量证明
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)

	var tmp big.Int
	tmp.SetBytes(hash[:])

	return tmp.Cmp(pow.target) == -1
}