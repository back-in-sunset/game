package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	endpoint  = "http://localhost:8000" // Alternator 端口
	region    = "us-east-1"             // 任意值，仅占位
	tableName = "users"
)

// User 数据结构
type User struct {
	UserID   string `dynamodbav:"user_id"`
	Email    string `dynamodbav:"email"`
	Age      int    `dynamodbav:"age"`
	LastSeen string `dynamodbav:"last_seen"`
}

// 初始化Alternator客户端（带连接池）
func initAlternatorClient() (*dynamodb.Client, error) {
	endpoint := "http://localhost:8000"
	// customResolver := aws.EndpointResolverWithOptionsFunc(
	// 	func(service, region string, options ...interface{}) (aws.Endpoint, error) {
	// 		return aws.Endpoint{
	// 			URL:           endpoint,
	// 			SigningRegion: "us-east-1", // 任意值，仅占位
	// 		}, nil
	// 	},
	// )

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithBaseEndpoint(endpoint),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		config.WithDefaultRegion(region),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)

	}

	return dynamodb.NewFromConfig(cfg), nil
}

// 创建表
func createTable(client *dynamodb.Client) error {
	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("user_id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("email"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("user_id"),
				KeyType:       types.KeyTypeHash, // 分区键
			},
			{
				AttributeName: aws.String("email"),
				KeyType:       types.KeyTypeRange, // 排序键
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	}

	_, err := client.CreateTable(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("创建表失败: %w", err)
	}

	// 等待表激活
	waiter := dynamodb.NewTableExistsWaiter(client)
	return waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}, 2*time.Minute)
}

// 插入用户
func putUser(client *dynamodb.Client, user User) error {
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("数据序列化失败: %w", err)
	}

	_, err = client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	return err
}

// 查询用户
func getUser(client *dynamodb.Client, userID, email string) (*User, error) {
	key, err := attributevalue.MarshalMap(map[string]string{
		"user_id": userID,
		"email":   email,
	})
	if err != nil {
		return nil, fmt.Errorf("键序列化失败: %w", err)
	}

	result, err := client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	var user User
	if err := attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return nil, fmt.Errorf("数据解析失败: %w", err)

	}
	return &user, nil
}
