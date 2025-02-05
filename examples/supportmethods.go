package examples

// NewSpecialObjectForCache конструктор вспомогательного типа реализующий интерфейс CacheStorageFuncHandler[T any]
func NewSpecialObjectForCache[T SpecialObjectComparator]() *SpecialObjectForCache[T] {
	return &SpecialObjectForCache[T]{}
}

func (o *SpecialObjectForCache[T]) SetID(v string) {
	o.id = v
}

func (o *SpecialObjectForCache[T]) GetID() string {
	return o.id
}

func (o *SpecialObjectForCache[T]) SetObject(v T) {
	o.object = v
}

func (o *SpecialObjectForCache[T]) GetObject() T {
	return o.object
}

func (o *SpecialObjectForCache[T]) SetFunc(f func(int) bool) {
	o.handlerFunc = f
}

func (o *SpecialObjectForCache[T]) GetFunc() func(int) bool {
	return o.handlerFunc
}

func (o *SpecialObjectForCache[T]) Comparison(objFromCache T) bool {
	if !o.object.ComparisonID(objFromCache.GetID()) {
		return false
	}

	if !o.object.ComparisonEvent(objFromCache.GetEvent()) {
		return false
	}

	if !o.object.ComparisonReports(objFromCache.GetReports()) {
		return false
	}

	if !o.object.ComparisonAttributes(objFromCache.GetAttributes()) {
		return false

	}

	if !o.object.ComparisonObjects(objFromCache.GetObjects()) {
		return false
	}

	if !o.object.ComparisonObjectTags(o.object.GetObjectTags()) {
		return false
	}

	return true
}
