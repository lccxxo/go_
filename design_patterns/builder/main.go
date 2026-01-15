package main

import "fmt"

// 建造者模式：将一个复杂对象的构建与它的表示分离，使得同样的构建过程可以创建不同的表示

// 产品：电脑
type Computer struct {
	CPU    string
	Memory string
	Disk   string
	GPU    string
}

func (c *Computer) Show() {
	fmt.Printf("电脑配置:\n")
	fmt.Printf("  CPU: %s\n", c.CPU)
	fmt.Printf("  内存: %s\n", c.Memory)
	fmt.Printf("  硬盘: %s\n", c.Disk)
	fmt.Printf("  显卡: %s\n", c.GPU)
}

// 建造者接口
type ComputerBuilder interface {
	SetCPU(cpu string) ComputerBuilder
	SetMemory(memory string) ComputerBuilder
	SetDisk(disk string) ComputerBuilder
	SetGPU(gpu string) ComputerBuilder
	Build() *Computer
}

// 具体建造者
type ConcreteComputerBuilder struct {
	computer *Computer
}

func NewComputerBuilder() *ConcreteComputerBuilder {
	return &ConcreteComputerBuilder{
		computer: &Computer{},
	}
}

func (b *ConcreteComputerBuilder) SetCPU(cpu string) ComputerBuilder {
	b.computer.CPU = cpu
	return b
}

func (b *ConcreteComputerBuilder) SetMemory(memory string) ComputerBuilder {
	b.computer.Memory = memory
	return b
}

func (b *ConcreteComputerBuilder) SetDisk(disk string) ComputerBuilder {
	b.computer.Disk = disk
	return b
}

func (b *ConcreteComputerBuilder) SetGPU(gpu string) ComputerBuilder {
	b.computer.GPU = gpu
	return b
}

func (b *ConcreteComputerBuilder) Build() *Computer {
	return b.computer
}

// 指挥者（可选）
type Director struct {
	builder ComputerBuilder
}

func NewDirector(builder ComputerBuilder) *Director {
	return &Director{builder: builder}
}

func (d *Director) BuildGamingComputer() *Computer {
	return d.builder.
		SetCPU("Intel i9-13900K").
		SetMemory("32GB DDR5").
		SetDisk("1TB NVMe SSD").
		SetGPU("RTX 4090").
		Build()
}

func (d *Director) BuildOfficeComputer() *Computer {
	return d.builder.
		SetCPU("Intel i5-12400").
		SetMemory("16GB DDR4").
		SetDisk("512GB SSD").
		SetGPU("集成显卡").
		Build()
}

func main() {
	// 方式1：直接使用建造者
	builder := NewComputerBuilder()
	computer1 := builder.
		SetCPU("AMD Ryzen 7 5800X").
		SetMemory("16GB DDR4").
		SetDisk("512GB SSD").
		SetGPU("RTX 3070").
		Build()
	computer1.Show()

	fmt.Println()

	// 方式2：使用指挥者
	director := NewDirector(NewComputerBuilder())
	gamingPC := director.BuildGamingComputer()
	fmt.Println("游戏电脑:")
	gamingPC.Show()

	fmt.Println()

	officePC := director.BuildOfficeComputer()
	fmt.Println("办公电脑:")
	officePC.Show()
}
