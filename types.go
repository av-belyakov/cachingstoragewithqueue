package cachingstoragewithqueue

import (
	"sync"
	"time"
)

// CacheStorageWithQueue кэш объектов с очередью
type CacheStorageWithQueue[T any] struct {
	queue    queueObjects[T]   //очередь объектов предназначенных для выполнения
	cache    cacheStorages[T]  //кеш хранилища обработанных объектов
	logging  WriterLoggingData //логирование данных
	maxTtl   time.Duration     //максимальное время, по истечении которого запись в cacheStorages будет удалена
	timeTick time.Duration     //интервал с которым будут выполнятся автоматические действия
	isAsync  int               //включить асинхронное выполнение заданий в кэше
}

// queueObjects очередь объектов
type queueObjects[T any] struct {
	mutex    sync.RWMutex
	storages []CacheStorageFuncHandler[T]
}

// cacheStorages кэш данных
type cacheStorages[T any] struct {
	mutex sync.RWMutex
	//основное хранилище
	storages map[string]storageParameters[T]
	//максимальный размер кэша при привышении которого выполняется удаление самой старой записи
	maxSize int
}

type storageParameters[T any] struct {
	//исходный объект над которым выполняются действия
	originalObject T
	//фунция-обертка выполнения
	cacheFunc func(int) bool
	//количество попыток выполнения функции
	numberExecutionAttempts int
	//общее время истечения жизни, время по истечению которого объект удаляется в любом
	//случае в независимости от того, был ли он выполнен или нет, формируется time.Now().Add(c.maxTTL)
	timeExpiry time.Time
	//основное время, по нему можно найти самый старый объект в кэше
	timeMain time.Time
	//результат выполнения
	isCompletedSuccessfully bool
	//статус выполнения
	isExecution bool
}

type cacheOptions[T any] func(*CacheStorageWithQueue[T]) error

type writeLog struct{}
