package cachingstoragewithqueue

import (
	"context"
	"errors"
	"fmt"
	"runtime"
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
		maxTtl:  time.Duration(30 * time.Second),
		logging: &writeLog{},
		//очередь
		queue: queueObjects[T]{
			storages: []CacheStorageFuncHandler[T](nil),
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
func (csq *CacheStorageWithQueue[T]) automaticExecution(ctx context.Context) {
	tick := time.NewTicker(csq.timeTick * time.Second)

	go func(ctx context.Context, tick *time.Ticker) {
		<-ctx.Done()
		tick.Stop()
	}(ctx, tick)

	for range tick.C {
		//поиск и удаление из хранилища всех объектов у которых истекло время жизни
		csq.DeleteForTimeExpiryObjectFromCache()

		//поиск и удаление самого старого объекта если размер кэша достиг максимального значения
		//выполняется удаление объекта который в настоящее время не выполняеться и ранее был успешно выполнен
		if csq.GetCacheSize() >= csq.cache.maxSize {
			if err := csq.DeleteOldestObjectFromCache(); err != nil {
				_, f, l, _ := runtime.Caller(0)
				csq.logging.Write("error", fmt.Sprintf("cachingstoragewithQueue package: '%s' %s:%d", err.Error(), f, l-1))
			}

			continue
		}

		if !csq.isAsync {
			//синхронная обработка задач
			csq.syncExecution()
		} else {
			//асинхронная обработка задач
			csq.asyncExecution()
		}
	}
}

// WithMaxTtl устанавливает максимальное время, по истечении которого запись в cacheStorages будет
// удалена, допустимый интервал времени хранения записи от 10 до 86400 секунд
func WithMaxTtl[T any](v int) cacheOptions[T] {
	return func(cswq *CacheStorageWithQueue[T]) error {
		if v < 10 || v > 86400 {
			return errors.New("the maximum time after which an entry in the cache will be deleted should not be less than 10 seconds or more than 24 hours (86400 seconds)")
		}

		cswq.maxTtl = time.Duration(v)

		return nil
	}
}

// WithTimeTick устанавливает интервал времени, заданное время такта, по истечении которого
// запускается новый виток автоматической обработки содержимого кэша, интервал значений должен
// быть в диапазоне от 3 до 120 секунд
func WithTimeTick[T any](v int) cacheOptions[T] {
	return func(cswq *CacheStorageWithQueue[T]) error {
		if v < 3 || v > 120 {
			return errors.New("the set clock cycle time should not be less than 3 seconds or more than 120 seconds")
		}

		cswq.timeTick = time.Duration(v)

		return nil
	}
}

// WithMaxSize устанавливает максимальный размер кэша, не может быть меньше 3 и больше 1000 хранимых объектов
func WithMaxSize[T any](v int) cacheOptions[T] {
	return func(cswq *CacheStorageWithQueue[T]) error {
		if v < 3 || v > 1000 {
			return errors.New("the maximum cache size cannot be less than 3 or more than 1000 objects")
		}

		cswq.cache.maxSize = v

		return nil
	}
}

// WithLogging устанавливает обработчик для записи информационных сообщений поступающих
// от модуля. Принимаемое значение должно соответствовать интерфейсу с едиственным
// методом Write(msgType, msg string) bool
func WithLogging[T any](customLogging WriterLoggingData) cacheOptions[T] {
	return func(cswq *CacheStorageWithQueue[T]) error {
		cswq.logging = customLogging

		return nil
	}
}

// WithEnableAsyncProcessing устанавливает асинхронное выполнение функций из кэша
func WithEnableAsyncProcessing[T any](bool) cacheOptions[T] {
	return func(cswq *CacheStorageWithQueue[T]) error {
		cswq.isAsync = true

		return nil
	}
}
