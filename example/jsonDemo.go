package main

import (
	"bytes"
	"fmt"
	"strings"
)

func main()  {
	//字符串连接
	str := []string{"hello","world","demo"}
	str2 := strings.Join(str,"=")
	fmt.Println(str2)

	//byte连接 用二维数组的方式演示做连接
	b := [][]byte{[]byte("hello"),[]byte("world"),[]byte("demo")}
	res := bytes.Join(b,[]byte("="))

	fmt.Printf("%s",res)
}
