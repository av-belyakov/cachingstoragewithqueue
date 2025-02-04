package cachingstoragewithqueue

type CacheStorageFuncHandler[T any] interface {
	CacheStorageGetter[T]
	CacheStorageSetter[T]
	Comparison(T) bool
}

type CacheStorageGetter[T any] interface {
	GetFunc() func(int) bool
	GetObject() T
}

type CacheStorageSetter[T any] interface {
	SetFunc(func(int) bool)
	SetObject(T)
}
