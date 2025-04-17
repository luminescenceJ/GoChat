package db

import (
	"Go-Chat/config"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

// Mongo全局客户端
var MongoClient *mongo.Client

type User struct {
	ID    string `bson:"_id,omitempty"`
	Name  string `bson:"name"`
	Email string `bson:"email"`
	Age   int    `bson:"age"`
}

func InitMongoDB() error {
	conf := config.Conf.Mongo // 假设你配置好了 config.Conf.Mongo

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s", conf.Username, conf.Password, conf.Host, conf.Port)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri).SetMaxPoolSize(100).SetMinPoolSize(20)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("连接失败:", err)
		return err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Ping失败:", err)
		return err
	}

	fmt.Println("✅ MongoDB 连接成功！")
	MongoClient = client
	return nil
}

func InsertUser(user User) {
	collection := MongoClient.Database("gochat").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Fatal("插入失败:", err)
	}
	fmt.Println("插入ID:", res.InsertedID)
}

func FindUsers() {
	collection := MongoClient.Database("gochat").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			log.Println("解码失败:", err)
		}
		fmt.Println(user)
	}
}

func UpdateUser(name string, newAge int) {
	collection := MongoClient.Database("gochat").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"name": name}
	update := bson.M{"$set": bson.M{"age": newAge}}

	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Fatal("更新失败:", err)
	}
	fmt.Printf("更新 %d 个文档\n", res.ModifiedCount)
}

func DeleteUser(name string) {
	collection := MongoClient.Database("gochat").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := collection.DeleteOne(ctx, bson.M{"name": name})
	if err != nil {
		log.Fatal("删除失败:", err)
	}
	fmt.Printf("删除 %d 个文档\n", res.DeletedCount)
}

//func main() {
//	InitMongoDB()
//
//	// 插入
//	user := User{Name: "Alice", Email: "alice@example.com", Age: 25}
//	InsertUser(user)
//
//	// 查询
//	FindUsers()
//
//	// 更新
//	UpdateUser("Alice", 30)
//
//	// 删除
//	DeleteUser("Alice")
//}
