# Пакет cachingstoragewithqueue

Пакет cachingstoragewithqueue реализует кэширующее хранилище произвольных объектов с очередью и обработкой этих объектов в автоматическом режиме. Для выполнения заданных действий над хранящимися объектами пользователем должны быть реализованы вспомогательные функции.

#### Подробно об использовании в примере ниже:

1. Инициализируем новое хранилище где 'any' любой тип данных который будет добавлятся в очередь и хранится в кэше как исходный объект. В тестах используется специально созданный интерфейс SpecialObjectComparator обладающий определённым набором методов:

```
	cache, err = cachingstoragewithq//кладем в очередь объекты которые необходимо обработать
		addObjectToQueue(listId)ueue.NewCacheStorage[<тип_объекта>](
		context.Background(),
		cachingstoragewithqueue.WithMaxTtl[<тип_объекта>](int),
		cachingstoragewithqueue.WithTimeTick[<тип_объекта>](int),
		cachingstoragewithqueue.WithMaxSize[<тип_объекта>](int))
```

При инициализации хранилища запускается модуль CacheStorageWithQueue.automaticExecution обеспечивающий автоматическую обработку объектов добавленных в очередь. Этот модуль выполняет обработку объектов, хранение заданного максимального количества объектов, сравнение объектов на наличие дубликатов, удаление старых объектов. То есть объектов, срок жизни которых истёк или объектов, которые были успешно выполнены, если максимальное количество которых превышает допустимое значение выставляемое методом cachingstoragewithqueue.WithMaxSize.

2.  Добавление нового объекта в очередь:
    Для добавления объекта в очередь нужно использовать любой вспомогательный объект который реализует интерфейс CacheStorageFuncHandler[T any].
    Например, используем пользовательский тип SpecialObjectForCache со вспомогательным интерфейсом SpecialObjectComparator,
    который, в том числе реализует методы сравнения схожих объектов для каждого из свойств объекта

```
type SpecialObjectForCache[T SpecialObjectComparator] struct {
	object      T
	handlerFunc func(int) bool
	id          string
}

func NewSpecialObjectForCache[T SpecialObjectComparator]() *SpecialObjectForCache[T] {
	return &SpecialObjectForCache[T]{}
}
```

Реализуем необходимые методы

```
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
    //некие сравнения...

    return true
}
```

Добавляем вспомогательный объект в очередь хранилища

```
cache.PushObjectToQueue(CacheStorageFuncHandler[T any])
```

После добавления вспомогательного объекта в очередь основная работа выполняется автоматически внутри хранилища.

ЗДЕСЬ ПОЗЖЕ НЕОБХОДИМО ОПИСАТЬ ПОДРОБНЫЙ ХОД ВЫПОЛНЕНИЯ ЗАДАЧИ

Подробный пример использования пакета, можно посмотреть в файле package_test.go где для тестирования используются наборы типов и методов из пакета github.com/av-belyakov/objectsmispformat применяемых для упрощения взаимодействия с API MISP.
