package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
)

type Wallets struct {
	WalletsMap map[string]*WalletKeyPair
}

//创建Wallets，返回Wallets实例
func NewWallets() *Wallets {
	var ws Wallets

	ws.WalletsMap = make(map[string]*WalletKeyPair)
	//1、把所有的钱包从本地加载出来
	ws.LoadFromFile()

	//2、把实例返回
	return &ws
}

const WalletName = "wallet.dat"

//这个Wallets是对外的，WalletKeyPair是对内的
//wallets调用WalletKeyPair
func (ws *Wallets) CreateWallet() string {
	//调用NewWalletKeyPair
	wallet := NewWalletKeyPair()
	//将返回的WalletKeyPair添加到WalletMap种
	address := wallet.GetAddress()

	ws.WalletsMap[address] = wallet
	//保存到本地文件
	res := ws.SaveToFile()
	if !res{
		fmt.Printf("创建钱包失败！\n")
		return ""
	}

	return address
}

//保存钱包到文件
func (ws *Wallets)SaveToFile() bool {
	//gob编码
	var buffer bytes.Buffer

	//将接口类型明确注册一下，否则gob编码失败 因为是interface类型
	gob.Register((elliptic.P256()))

	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(ws)

	if err != nil{
		fmt.Printf("钱包序列化失败！，err:%v\n",err)
		return false
	}

	content := buffer.Bytes()

	//func WriteFile(filename string, data []byte, perm os.FileMode) error {
	err = ioutil.WriteFile(WalletName,content,0600)
	if err != nil{
		fmt.Printf("钱包创建失败！\n")
		return false
	}
	return true
}

func (ws *Wallets) LoadFromFile() bool {
	//判断文件是否存在
	if !IsFileExist(WalletName){
		fmt.Printf("钱包文件不存在，准备创建！\n")
		return true
	}

	//读取文件
	//func ReadFile(filename string) ([]byte, error) {
	content,err := ioutil.ReadFile(WalletName)

	if err != nil{
		return false
	}

	//解码时也要注册interface与编码时一样
	gob.Register(elliptic.P256())

	//gob解码
	decoder := gob.NewDecoder(bytes.NewReader(content))

	var wallets Wallets //也要创建一样的类

	err = decoder.Decode(&wallets)

	if err != nil{
		fmt.Printf("err : %v\n",err)
		return false
	}

	//赋值给ws
	ws.WalletsMap = wallets.WalletsMap

	return true
}

//遍历钱包，打印所有地址
func (ws *Wallets) ListAddress() []string {
	//遍历ws.WalletsMap结构返回key即可

	var addresses []string

	for address,_ := range ws.WalletsMap{
		addresses = append(addresses,address)
	}

	return addresses
}
