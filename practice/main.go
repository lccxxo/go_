package main

import (
	"fmt"
	"time"
)

func main() {
	go printlnNumber()
	time.Sleep(time.Second * 5)
	fmt.Println("hello world222222222222222222222222222222")
	time.Sleep(time.Second * 2)
}

func printlnNumber() {
	for {
		fmt.Println("hello world")
		time.Sleep(1 * time.Second)
	}
}
