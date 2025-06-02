package cachingstoragewithqueue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/av-belyakov/cachingstoragewithqueue/internal/supportingfunctions"
)

// NewCacheStorage создает новое кэширующее хранилище, а также очередь из которой будут, в автоматическом
// режиме, браться объекты предназначенные для обработки. Для обработки объектов будет использоваться
// пользовательская функция-обёртка, которую, как и обрабатываемый объект, добавляют с использованием
// вспомогательного пользовательского типа.
func NewCacheStorage[T any](opts ...cacheOptions[T]) (*CacheStorageWithQueue[T], error) {
	cacheExObj := &CacheStorageWithQueue[T]{
		//значение по умолчанию для интервала автоматической обработки
		timeTick: time.Duration(5 * time.Second),
		//значение по умолчанию для времени жизни объекта
		maxTtl:  time.Duration(3600 * time.Second),
		logging: &writeLog{},
		//очередь
		queue: queueObjects[T]{
			storages: []CacheStorageHandler[T](nil),
		},
		cache: cacheStorages[T]{
			//значение по умолчанию максимального размера кэша
			maxSize: 15,
			//основное хранилище
			storages: map[string]storageParameters[T]{},
		},
	}

	for _, opt := range opts {
		if err := opt(cacheExObj); err != nil {
			return cacheExObj, err
		}
	}

	//проверяем количество потоков и размер кэша, если многопоточный режим выполнения
	//активирован
	if cacheExObj.isAsync >= 2 {
		//размер кэша должен как минимум в два раза превышать количество потоков асинхронного выполнения
		if cacheExObj.cache.maxSize < cacheExObj.isAsync || (cacheExObj.cache.maxSize/cacheExObj.isAsync) < 2 {
			return cacheExObj, errors.New("the cache size must be at least twice the number of asynchronous execution threads")
		}
	}

	return cacheExObj, nil
}

// StartAutomaticExecution автоматическая обработка очередей и объектов в кэше
func (c *CacheStorageWithQueue[T]) StartAutomaticExecution(ctx context.Context) {
	go func() {
		tick := time.NewTicker(c.timeTick)
		defer tick.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-tick.C:
				//поиск и удаление из хранилища всех объектов у которых истекло время жизни
				c.DeleteForTimeExpiryObjectFromCache()

				//поиск и удаление самого старого объекта если размер кэша достиг максимального значения
				//выполняется удаление объекта который в настоящее время не выполняеться и ранее был успешно выполнен
				if c.GetCacheSize() == c.cache.maxSize {
					if err := c.DeleteOldestObjectFromCache(); err != nil {
						c.logging.Write("error", supportingfunctions.CustomError(fmt.Errorf("cachingstoragewithqueue package: '%s'", err.Error())).Error())
					}
				}

				if c.isAsync >= 2 {
					//асинхронная обработка задач
					c.asyncExecution(ctx)
				} else {
					//синхронная обработка задач
					c.syncExecution(ctx)
				}
			}
		}
	}()
}

// WithMaxTtl устанавливает максимальное время, по истечении которого запись в cacheStorages будет
// удалена, допустимый интервал времени хранения записи от 60 до 86400 секунд
func WithMaxTtl[T any](v int) cacheOptions[T] {
	return func(cswq *CacheStorageWithQueue[T]) error {
		if v < 60 || v > 86400 {
			return errors.New("the maximum time after which an entry in the cache will be deleted should not be less than 300 seconds or more than 24 hours (86400 seconds)")
		}

		cswq.maxTtl = time.Duration(v) * time.Second

		return nil
	}
}

// WithTimeTick устанавливает интервал времени, заданное время такта, по истечении которого
// запускается новый виток автоматической обработки содержимого кэша, интервал значений должен
// быть в диапазоне от 1 до 120 секунд
func WithTimeTick[T any](v int) cacheOptions[T] {
	return func(cswq *CacheStorageWithQueue[T]) error {
		if v < 1 || v > 120 {
			return errors.New("the set clock cycle time should not be less than 3 seconds or more than 120 seconds")
		}

		cswq.timeTick = time.Duration(v) * time.Second

		return nil
	}
}

// WithMaxSize устанавливает максимальный размер кэша, не может быть меньше 3 и больше 1000 хранимых объектов
// кроме того, размер кэша должен минимум в ДВА раза первышать количество асинхронных потоков
// выполнения, если асинхронный режим активирован
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

// WithEnableAsyncProcessing устанавливает асинхронное выполнение функций в кэша, при этом
// асинхронное выполнение будет активировано только если количество потоков, заданных
// через эту функцию, будут два и более. Максимальное количество потоков должно быть меньше
// размер кэша как минимум в ДВА раза. Например, если количество потоков 4, размер кэша не может
// быть меньше 8
func WithEnableAsyncProcessing[T any](numberStreams int) cacheOptions[T] {
	return func(cswq *CacheStorageWithQueue[T]) error {
		cswq.isAsync = numberStreams

		return nil
	}
}
