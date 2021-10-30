package main

import "fmt"

type ClassMate struct {
	Name string
	Age uint64
}

func main()  {
	var arr []*ClassMate

	class1 := ClassMate{"小红",20}
	class2 := ClassMate{"小B",30}

	arr = append(arr, &class1,&class2)

	fmt.Println(*arr[0])
	fmt.Println(*arr[1])

}