package observer

// Subject is an interface for objects that can be observed.
type Subject[T VersionedConfig] interface {
	Register(observer Observer[T])
	Remove(observer Observer[T])
}

// Observer is an interface for objects that can observe a Subject.
type Observer[T VersionedConfig] interface {
	ID() string
	Update(VersionedConfig)
}
