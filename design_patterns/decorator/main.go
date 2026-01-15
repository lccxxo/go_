package main

import "fmt"

// 装饰器模式：动态地给对象添加一些额外的职责

// 咖啡接口
type Coffee interface {
	GetDescription() string
	GetCost() float64
}

// 基础咖啡
type SimpleCoffee struct{}

func (c *SimpleCoffee) GetDescription() string {
	return "普通咖啡"
}

func (c *SimpleCoffee) GetCost() float64 {
	return 10.0
}

// 咖啡装饰器
type CoffeeDecorator struct {
	coffee Coffee
}

func (cd *CoffeeDecorator) GetDescription() string {
	return cd.coffee.GetDescription()
}

func (cd *CoffeeDecorator) GetCost() float64 {
	return cd.coffee.GetCost()
}

// 牛奶装饰器
type MilkDecorator struct {
	CoffeeDecorator
}

func NewMilkDecorator(coffee Coffee) *MilkDecorator {
	return &MilkDecorator{
		CoffeeDecorator: CoffeeDecorator{coffee: coffee},
	}
}

func (m *MilkDecorator) GetDescription() string {
	return m.coffee.GetDescription() + " + 牛奶"
}

func (m *MilkDecorator) GetCost() float64 {
	return m.coffee.GetCost() + 2.0
}

// 糖装饰器
type SugarDecorator struct {
	CoffeeDecorator
}

func NewSugarDecorator(coffee Coffee) *SugarDecorator {
	return &SugarDecorator{
		CoffeeDecorator: CoffeeDecorator{coffee: coffee},
	}
}

func (s *SugarDecorator) GetDescription() string {
	return s.coffee.GetDescription() + " + 糖"
}

func (s *SugarDecorator) GetCost() float64 {
	return s.coffee.GetCost() + 1.0
}

func main() {
	// 基础咖啡
	coffee := &SimpleCoffee{}
	fmt.Printf("%s - 价格: %.2f 元\n", coffee.GetDescription(), coffee.GetCost())

	// 加牛奶
	milkCoffee := NewMilkDecorator(coffee)
	fmt.Printf("%s - 价格: %.2f 元\n", milkCoffee.GetDescription(), milkCoffee.GetCost())

	// 加牛奶和糖
	sugarMilkCoffee := NewSugarDecorator(milkCoffee)
	fmt.Printf("%s - 价格: %.2f 元\n", sugarMilkCoffee.GetDescription(), sugarMilkCoffee.GetCost())
}
