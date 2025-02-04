package cachingstoragewithqueue

import (
	"sync"
	"time"
)

// CacheExecutedObjects кэш объектов с очередью
type CacheExecutedObjects[T any] struct {
	queue    listQueueObjects[T] //очередь объектов предназначенных для выполнения
	cache    cacheStorages[T]    //кеш хранилища обработанных объектов
	maxTtl   time.Duration       //максимальное время, по истечении которого запись в cacheStorages будет удалена
	timeTick time.Duration       //интервал с которым будут выполнятся автоматические действия
}

type listQueueObjects[T any] struct {
	mutex    sync.Mutex
	storages []T
}

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
	//основное время, по данному времени можно найти самый старый объект в кэше
	timeMain time.Time
	//общее время истечения жизни, время по истечению которого объект удаляется в любом
	//случае в независимости от того, был ли он выполнен или нет, формируется time.Now().Add(c.maxTTL)
	timeExpiry time.Time
	//фунция-обертка выполнения
	cacheFunc func(int) bool
	//статус выполнения
	isExecution bool
	//результат выполнения
	isCompletedSuccessfully bool
}

type cacheOptions[T any] func(*CacheExecutedObjects[T]) error
