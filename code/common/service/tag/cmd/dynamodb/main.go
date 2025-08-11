package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type AccountID string

type Identifier interface {
	Identify() string
}

func (i *AccountID) Identify() string {
	return string(*i)
}

type Account struct {
	ID          AccountID   `json:"id"`
	PhoneNumber PhoneNumber `json:"phone_number"`
}

type PhoneNumber struct {
	Protoc string `json:"protoc"`
	Number string `json:"number"`
}

func (p PhoneNumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%s-%s", p.Protoc, p.Number))
}

func main() {
	// account := Account{
	// 	ID: AccountID("1"),
	// 	PhoneNumber: PhoneNumber{
	// 		Protoc: "+86",
	// 		Number: "17854255492",
	// 	},
	// }

	// b, err := json.Marshal(account)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(string(b))

	// 初始化客户端
	client, err := initAlternatorClient()
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}

	// 创建表
	// if err := createTable(client); err != nil {
	// 	log.Printf("警告: %v", err) // 表可能已存在
	// }

	// 写入数据
	// user := User{
	// 	UserID:   "u123",
	// 	Email:    "alice@example.com",
	// 	Age:      28,
	// 	LastSeen: time.Now().Format(time.RFC3339),
	// }
	// if err := putUser(client, user); err != nil {
	// 	log.Fatalf("写入失败: %v", err)
	// }

	// 查询数据
	retrieved, err := getUser(client, "u123", "alice@example.com")
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}
	fmt.Printf("查询结果: %+v\n", retrieved)

	time.Sleep(1 * time.Minute)
}
