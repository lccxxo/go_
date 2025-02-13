package main

import "fmt"

// 外观模式

type API interface {
	Say()
}

func NewAPI() API {
	return &apiImpl{
		a: &aModuleImpl{},
		b: &bModuleImpl{},
	}
}

type apiImpl struct {
	a aModuleAPI
	b bModuleAPI
}

func (a *apiImpl) Say() {
	a.a.SayA()
	a.b.SayB()
}

type aModuleAPI interface {
	SayA()
}
type bModuleAPI interface {
	SayB()
}

type aModuleImpl struct{}
type bModuleImpl struct{}

func (a *aModuleImpl) SayA() {
	fmt.Println("say A")
}

func (b *bModuleImpl) SayB() {
	fmt.Println("say B")
}

func main() {
	api := NewAPI()
	api.Say()
}
