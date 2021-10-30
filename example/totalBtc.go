package main

import "fmt"

const blockInterval = 21 //万 区块衰减周期 21万区块

func main()  {
	total := 0.0 //总量
	reward := 50.0 //初始奖励
	times := 0 //减半次数
	start := 2009 //起始年份
	for reward > 0 {
		amount := reward*blockInterval //每一个衰减周期生产的数量
		total = total + amount
		reward *= 0.5 //周期减半
		times += 1
		start += 4 //加减半周期
		if reward > 0.00000001{
			fmt.Printf("%d年区块奖励=%.8f\n",start,reward)
		}
	}
	fmt.Println(total)
	fmt.Printf("减半次数%d\n",times) // 2137年后也依然在减半不过奖励将低于0.00000001
}
