package main

import "fmt"

// 观察者模式：定义对象间一对多的依赖关系，当一个对象状态改变时，所有依赖它的对象都会得到通知

// 观察者接口
type Observer interface {
	Update(message string)
}

// 主题接口
type Subject interface {
	Attach(observer Observer)
	Detach(observer Observer)
	Notify(message string)
}

// 新闻发布者（具体主题）
type NewsPublisher struct {
	observers []Observer
}

func NewNewsPublisher() *NewsPublisher {
	return &NewsPublisher{
		observers: make([]Observer, 0),
	}
}

func (np *NewsPublisher) Attach(observer Observer) {
	np.observers = append(np.observers, observer)
}

func (np *NewsPublisher) Detach(observer Observer) {
	for i, obs := range np.observers {
		if obs == observer {
			np.observers = append(np.observers[:i], np.observers[i+1:]...)
			break
		}
	}
}

func (np *NewsPublisher) Notify(message string) {
	for _, observer := range np.observers {
		observer.Update(message)
	}
}

// 订阅者A（具体观察者）
type SubscriberA struct {
	name string
}

func NewSubscriberA(name string) *SubscriberA {
	return &SubscriberA{name: name}
}

func (s *SubscriberA) Update(message string) {
	fmt.Printf("[%s] 收到新闻: %s\n", s.name, message)
}

// 订阅者B（具体观察者）
type SubscriberB struct {
	name string
}

func NewSubscriberB(name string) *SubscriberB {
	return &SubscriberB{name: name}
}

func (s *SubscriberB) Update(message string) {
	fmt.Printf("[%s] 收到新闻: %s\n", s.name, message)
}

func main() {
	// 创建新闻发布者
	publisher := NewNewsPublisher()

	// 创建订阅者
	subscriber1 := NewSubscriberA("张三")
	subscriber2 := NewSubscriberB("李四")

	// 订阅新闻
	publisher.Attach(subscriber1)
	publisher.Attach(subscriber2)

	// 发布新闻
	publisher.Notify("今日头条：Go语言发布新版本！")

	// 取消订阅
	publisher.Detach(subscriber1)

	// 再次发布新闻
	publisher.Notify("科技新闻：AI技术取得重大突破！")
}
