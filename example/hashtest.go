package main

import (
	"crypto/sha256"
	"fmt"
)

func main()  {
	data := "123"

	fmt.Printf("%x",sha256.Sum256([]byte(data)))
}
