package main

import (
	"fmt"
	"sync"
)

// 单例模式：确保一个类只有一个实例，并提供全局访问点

// 数据库连接示例
type Database struct {
	connection string
}

var (
	instance *Database
	once     sync.Once
)

// GetInstance 使用 sync.Once 确保线程安全的单例
func GetInstance() *Database {
	once.Do(func() {
		instance = &Database{
			connection: "mysql://localhost:3306/mydb",
		}
		fmt.Println("数据库连接已创建")
	})
	return instance
}

func (d *Database) Query(sql string) {
	fmt.Printf("执行查询: %s (连接: %s)\n", sql, d.connection)
}

func main() {
	// 多次获取实例，但只会创建一个
	db1 := GetInstance()
	db2 := GetInstance()

	fmt.Printf("db1 == db2: %v\n", db1 == db2) // true

	db1.Query("SELECT * FROM users")
	db2.Query("SELECT * FROM orders")
}
