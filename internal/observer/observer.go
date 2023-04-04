package observer

// Subject is an interface for objects that can be observed.
type Subject interface {
	Register(observer Observer)
	Remove(observer Observer)
	Notify()
}

// Observer is an interface for objects that can observe a Subject.
type Observer interface {
	Update()
}
