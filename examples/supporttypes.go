package examples

// SpecialObjectForCache является вспомогательным типом который реализует интерфейс
// CacheStorageFuncHandler[T any] где в методе Comparison(objFromCache T) bool необходимо
// реализовать подробное сравнение объекта типа T
type SpecialObjectForCache[T SpecialObjectComparator] struct {
	object      T
	handlerFunc func(int) bool
	id          string
}
