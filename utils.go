package main

import (
	"bytes"
	"encoding/binary"
	"os"
)

//工具函数
func uintToByte(num uint64) []byte {
	//使用binary.Write进行编码
	var buffer bytes.Buffer
	//编码进行错误检查 一定要做
	err := binary.Write(&buffer,binary.BigEndian,num)
	if err != nil{
		panic(err)
	}
	return buffer.Bytes()
}

//判断文件是否存在
func IsFileExist(fileName string) bool {
	//使用os.Stat来判断
	//func Stat(name string) (FileInfo, error) {
	_,err := os.Stat(fileName)

	if os.IsNotExist(err){ //不存在返回false
		return false
	}

	return true
}
