package main

import (
	"fmt"
	"strings"
)

type Test struct {
	str string
}

//给结构添加一个String()

func (test *Test) String() string {
	res := fmt.Sprintf("hello world:%s\n",test.str)
	return res
}

func main()  {
	//案例1
	t1 := &Test{"您好"}
	fmt.Printf("%v\n",t1)
	//打印结果 hello world:您好

	//连接案例2
	res2 := strings.Join([]string{"1","2","3"},"+") //error
	fmt.Printf("%v",res2)
	//打印结果 1+2+3
}
