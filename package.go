package cachingstoragewithqueue

import (
	"context"
	"errors"
	"time"
)

// NewCacheStorage создает новое кэширующее хранилище, а также очередь из которой будут, в автоматическом
// режиме, браться объекты предназначенные для обработки. Для обработки объектов будет использоваться
// пользовательская функция-обёртка, которую, как и обрабатываемый объект, добавляют с использованием
// вспомогательного пользовательского типа.
func NewCacheStorage[T any](ctx context.Context, opts ...cacheOptions[T]) (*CacheStorageWithQueue[T], error) {
	cacheExObj := &CacheStorageWithQueue[T]{
		//значение по умолчанию для интервала автоматической обработки
		timeTick: time.Duration(5 * time.Second),
		//значение по умолчанию для времени жизни объекта
		maxTtl: time.Duration(30 * time.Second),
		//очередь
		queue: queueObjects[T]{
			storages: []T(nil),
		},
		cache: cacheStorages[T]{
			//значение по умолчанию максимального размера кэша
			maxSize: 8,
			//основное хранилище
			storages: map[string]storageParameters[T]{},
		},
	}

	for _, opt := range opts {
		if err := opt(cacheExObj); err != nil {
			return cacheExObj, err
		}
	}

	go cacheExObj.automaticExecution(ctx)

	return cacheExObj, nil
}

//	isExecution             bool           //статус выполнения
//	isCompletedSuccessfully bool           //результат выполнения

// automaticExecution работает с очередями и кэшем объектов
// С заданным интервалом времени выполняет следующие действия:
// 1. Проверка, есть ли в кеше объекты. Если кеш пустой, проверка
// наличия объектов в очереди, если она пуста - ожидание. Если очередь
// не пуста, взять объект из очереди и положить в кеш. Запуск вновь
// добавленного объекта. Если кеш не пуст, то выполняется пункт 2.
// 2. Проверка, есть ли в кеше объект со стаусом isExecution=true, если
// есть, то ожидание завершения выполнения объекта в результате которого
// меняется статус объекта на isExecution=true (может быть false?) и isCompletedSuccessfully=true.
// Если статус объекта меняется на isExecution=true (может быть false?), а isCompletedSuccessfully=false
// то выполняется повторный запуск объекта. Если статус объекта isExecution=true (может быть false?)
// и isCompletedSuccessfully=true, то выполняется пункт 3.
// 3. Проверка, есть ли в кеше свободное место, равное переменной maxSize, если
// свободное место есть, выполняется ПОИСК в кеше, по уникальному ключу, объекта
// соответствующего объекту принимаемому из очереди, при условии что, очередь
// не пуста. Если такой объект находится, то выполняется СРАВНЕНИЕ двух объектов.
// При их полном совпадении объект, полученный из очереди удаляется, а в обработку
// добавляется следующий в очереди объект.
// Вновь добавленный объект запускается на выполнение. Если в кеше нет свободного
// места переходим к пункту 4.
// 4. Проверка, есть ли в кеше объекты со статусом isExecution=false и
// isCompletedSuccessfully=true, если есть, то формируется список объектов подходящих
// под заданные условия из которых удаляется объект с самым старым timeMain. После
// удаления объекта с самым старым timeMain, обращение к очереди за новым объектом
// предназначенным для обработки.
func (cache *CacheStorageWithQueue[T]) automaticExecution(ctx context.Context) {
	tick := time.NewTicker(cache.timeTick * time.Second)
	go func(ctx context.Context, tick *time.Ticker) {
		<-ctx.Done()
		tick.Stop()
	}(ctx, tick)

	var (
		isEmpty       bool
		currentObject T
	)

	for range tick.C {
		//поиск и удаление из хранилища всех объектов у которых истекло время жизни
		cache.DeleteForTimeExpiryObjectFromCache()

		currentObject, isEmpty = cache.PullObjectFromQueue()
		// если очередь пуста, нет объектов для обработки
		if isEmpty {
			continue
		}

		cache.AddObjectToCache(currentObject.GetID, currentObject)
	}
}

// WithMaxTtl устанавливает максимальное время, по истечении которого запись в cacheStorages будет
// удалена, допустимый интервал времени хранения записи от 10 до 86400 секунд
func WithMaxTtl[T any](v int) cacheOptions[T] {
	return func(ceo *CacheStorageWithQueue[T]) error {
		if v < 10 || v > 86400 {
			return errors.New("the maximum time after which an entry in the cache will be deleted should not be less than 10 seconds or more than 24 hours (86400 seconds)")
		}

		ceo.maxTtl = time.Duration(v)

		return nil
	}
}

// WithTimeTick устанавливает интервал времени, заданное время такта, по истечении которого
// запускается новый виток автоматической обработки содержимого кэша, интервал значений должен
// быть в диапазоне от 3 до 120 секунд
func WithTimeTick[T any](v int) cacheOptions[T] {
	return func(ceo *CacheStorageWithQueue[T]) error {
		if v < 3 || v > 120 {
			return errors.New("the set clock cycle time should not be less than 3 seconds or more than 120 seconds")
		}

		ceo.timeTick = time.Duration(v)

		return nil
	}
}

// WithMaxSize устанавливает максимальный размер кэша, не может быть меньше 3 и больше 1000 хранимых объектов
func WithMaxSize[T any](v int) cacheOptions[T] {
	return func(ceo *CacheStorageWithQueue[T]) error {
		if v < 3 || v > 1000 {
			return errors.New("the maximum cache size cannot be less than 3 or more than 1000 objects")
		}

		ceo.cache.maxSize = v

		return nil
	}
}
