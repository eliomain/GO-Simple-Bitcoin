package main

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

func main() {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	db.Update(func(tx *bolt.Tx) error {
		//所有操作都在这里
		bc := tx.Bucket([]byte("bucketName"))
		if bc == nil{
			fmt.Println("创建数据桶")
			//为空创建 数据桶
			bc ,err = tx.CreateBucket([]byte("bucketName"))
			if err != nil{
				log.Panic("创建数据桶失败")
			}
		}

		//bucket已经创建完成，准备写入数据
		//写数据使用Put，读数据使用Get
		err = bc.Put([]byte("name1"),[]byte("tom"))
		if err != nil{
			fmt.Println("name1-tom写入失败")
		}
		err = bc.Put([]byte("name2"),[]byte("joe"))
		if err != nil{
			fmt.Println("name2-joe写入失败")
		}

		//读取数据
		name1 := bc.Get([]byte("name1"))
		name2 := bc.Get([]byte("name2"))
		name3 := bc.Get([]byte("name3"))

		fmt.Printf("name1: %s\n", name1)
		fmt.Printf("name2: %s\n", name2)
		fmt.Printf("name3: %s\n", name3)

		return nil
	})

}