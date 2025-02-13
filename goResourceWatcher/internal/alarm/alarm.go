package alarm

// todo 具体的告警

// Observer
type Observer interface {
	Notify(message string)
}

// Subject
type Subject interface {
	RegisterObserver(observer Observer)
	RemoveObserver(observer Observer)
	NotifyObservers(message string)
}

// Alarm
type Alarm struct {
	observers map[Observer]struct{}
}

// NewAlarm
func NewAlarm() *Alarm {
	return &Alarm{
		observers: make(map[Observer]struct{}),
	}
}

func (a *Alarm) RegisterObservers(observers ...Observer) {
	for _, observer := range observers {
		a.observers[observer] = struct{}{}
	}
}

func (a *Alarm) RemoveObserver(observer Observer) {
	delete(a.observers, observer)
}

func (a *Alarm) NotifyObservers(message string) {
	for observer := range a.observers {
		observer.Notify(message)
	}
}
