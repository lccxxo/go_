package main

import (
	"fmt"
)

// 简单工厂设计模式
// 一般由一个NewXXX方法返回一个interface类型来调用方法

type API interface {
	Say()
}

type Animal struct {
	Name string
}

type People struct {
	Name string
}

func (a *Animal) Say() {
	fmt.Println("i am a animal, name is :", a.Name)
}

func (p *People) Say() {
	fmt.Println("i am a people, name is :", p.Name)
}

func NewAPI(iType, name string) API {
	if iType == "animal" {
		return &Animal{Name: name}
	} else if iType == "people" {
		return &People{Name: name}
	}
	return nil
}

func main() {
	peopleApi := NewAPI("people", "BOB")
	peopleApi.Say()
}
