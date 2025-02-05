package cachingstoragewithqueue

type CacheStorageFuncHandler[T any] interface {
	CacheStorageGetter[T]
	CacheStorageSetter[T]
	Comparison(T) bool
}

type CacheStorageGetter[T any] interface {
	GetID() string
	GetFunc() func(int) bool
	GetObject() T
}

type CacheStorageSetter[T any] interface {
	SetID(string)
	SetFunc(func(int) bool)
	SetObject(T)
}

type WriterLoggingData interface {
	Write(msgType, msg string) bool
}
