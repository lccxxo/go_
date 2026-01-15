package main

import "fmt"

// 工厂方法模式：定义一个创建对象的接口，让子类决定实例化哪一个类

// 支付接口
type Payment interface {
	Pay(amount float64) error
}

// 支付宝支付
type Alipay struct{}

func (a *Alipay) Pay(amount float64) error {
	fmt.Printf("使用支付宝支付 %.2f 元\n", amount)
	return nil
}

// 微信支付
type WeChatPay struct{}

func (w *WeChatPay) Pay(amount float64) error {
	fmt.Printf("使用微信支付 %.2f 元\n", amount)
	return nil
}

// 支付工厂接口
type PaymentFactory interface {
	CreatePayment() Payment
}

// 支付宝工厂
type AlipayFactory struct{}

func (f *AlipayFactory) CreatePayment() Payment {
	return &Alipay{}
}

// 微信支付工厂
type WeChatPayFactory struct{}

func (f *WeChatPayFactory) CreatePayment() Payment {
	return &WeChatPay{}
}

func main() {
	// 使用支付宝工厂
	alipayFactory := &AlipayFactory{}
	alipay := alipayFactory.CreatePayment()
	alipay.Pay(100.00)

	// 使用微信支付工厂
	wechatFactory := &WeChatPayFactory{}
	wechat := wechatFactory.CreatePayment()
	wechat.Pay(200.00)
}
