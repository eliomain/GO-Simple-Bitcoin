package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
)

type Person struct {
	Name string
	Age uint64
}

func main() {
	Jim := Person{
		Name : "Jim",
		Age: 20,
	}

	//编码
	var buffer bytes.Buffer
	//定义编码器
	buf := gob.NewEncoder(&buffer)
	//编码器对结构进行编码，一定要进行校验
	err := buf.Encode(&Jim)
	if err != nil{
		log.Panic(err)
	}

	fmt.Printf("编码后的数据: %x\n", buffer.Bytes())

	//...传输

	//解码，将字节流转换Person结构
	//解码时，先要创建解码器，解码器进行解码
	var p1 Person
	//创建解码器
	De := gob.NewDecoder(bytes.NewReader(buffer.Bytes()))
	err = De.Decode(&p1)
	if err != nil{
		log.Panic(err)
	}

	fmt.Printf("解码后的数据%v",p1)
}
