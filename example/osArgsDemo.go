package main

import (
	"fmt"
	"os"
)

func main()  {
	cmds := os.Args
	for i,k := range cmds{
		fmt.Printf("cmds[%d]=%s\n",i,k)
	}
}
