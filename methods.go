package cachingstoragewithqueue

import (
	"fmt"
	"time"
)

// SizeObjectToQueue размер очереди
func (c *CacheExecutedObjects[T]) SizeObjectToQueue() int {
	return len(c.queue.storages)
}

// CleanQueue очистка очереди
func (c *CacheExecutedObjects[T]) CleanQueue() {
	c.queue.mutex.Lock()
	defer c.queue.mutex.Unlock()

	c.queue.storages = []T(nil)
}

// PushObjectToQueue добавляет в очередь объектов новый объект
func (c *CacheExecutedObjects[T]) PushObjectToQueue(v T) {
	c.queue.mutex.Lock()
	defer c.queue.mutex.Unlock()

	c.queue.storages = append(c.queue.storages, v)
}

// PullObjectFromQueue забирает с начала очереди новый объект или возвращает
// TRUE если очередь пуста
func (c *CacheExecutedObjects[T]) PullObjectFromQueue() (T, bool) {
	c.queue.mutex.Lock()
	defer c.queue.mutex.Unlock()

	var obj T
	size := len(c.queue.storages)
	if size == 0 {
		return obj, true
	}

	obj = c.queue.storages[0]

	if size == 1 {
		c.queue.storages = make([]T, 0)

		return obj, false
	}

	c.queue.storages = c.queue.storages[1:]

	return obj, false
}

// AddObjectToCache добавляет новый объект в хранилище
func (c *CacheExecutedObjects[T]) AddObjectToCache(key string, value CacheStorageFuncHandler[T]) error {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	if len(c.cache.storages) >= c.cache.maxSize {
		//удаление самого старого объекта, осуществляется по параметру timeMain
		c.deleteOldestObjectFromCache()
	}

	storage, ok := c.cache.storages[key]
	if !ok {
		c.cache.storages[key] = storageParameters[T]{
			timeMain:       time.Now(),
			timeExpiry:     time.Now().Add(c.maxTtl),
			originalObject: value.GetObject(),
			cacheFunc:      value.GetFunc(),
		}

		return nil
	}

	//найден объект у которого ключ совпадает с объектом принятом в обработку

	//объект в настоящее время выполняется
	if storage.isExecution {
		return fmt.Errorf("an object has been received whose key ID '%s' matches the already running object, ignore it", key)
	}

	//сравнение объектов из кэша и полученного из очереди
	if value.Comparison(storage.originalObject) {
		return fmt.Errorf("objects with key ID '%s' are completely identical, adding an object to the cache is not performed", key)
	}

	storage.timeMain = time.Now()
	storage.timeExpiry = time.Now().Add(c.maxTtl)
	storage.isExecution = false
	storage.isCompletedSuccessfully = false
	storage.originalObject = value.GetObject()
	storage.cacheFunc = value.GetFunc()

	c.cache.storages[key] = storage

	return nil
}

// GetOldestObjectFromCache возвращает индекс самого старого объекта
func (c *CacheExecutedObjects[T]) GetOldestObjectFromCache() string {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	return c.getOldestObjectFromCache()
}

// GetObjectFromCacheByKey возвращает орбъект из кэша по ключу
func (c *CacheExecutedObjects[T]) GetObjectFromCacheByKey(key string) (T, bool) {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	storage, ok := c.cache.storages[key]

	return storage.originalObject, ok
}

// GetFuncFromCacheByKey возвращает исполняемую функцию из кэша по ключу
func (c *CacheExecutedObjects[T]) GetFuncFromCacheByKey(key string) (func(int) bool, bool) {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	storage, ok := c.cache.storages[key]

	return storage.cacheFunc, ok
}

// GetObjectFromCacheMinTimeExpiry возвращает из кэша объект, функция которого в настоящее время
// не выполняется, не была успешно выполнена и время истечения жизни объекта которой самое меньшее
func (c *CacheExecutedObjects[T]) GetObjectFromCacheMinTimeExpiry() (key string, obj T) {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	var early time.Time
	for k, v := range c.cache.storages {
		if key == "" {
			key = k
			early = v.timeExpiry
			obj = v.originalObject

			continue
		}

		if v.timeExpiry.Before(early) {
			key = k
			early = v.timeExpiry
			obj = v.originalObject
		}
	}

	return
}

// GetFuncFromCacheMinTimeExpiry возвращает из кэша исполняемую функцию которая в настоящее время
// не выполняется, не была успешно выполнена и время истечения жизни объекта которой самое меньшее
func (c *CacheExecutedObjects[T]) GetFuncFromCacheMinTimeExpiry() (key string, f func(int) bool) {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	var early time.Time
	for k, v := range c.cache.storages {
		if key == "" {
			key = k
			early = v.timeExpiry
			f = v.cacheFunc

			continue
		}

		if v.timeExpiry.Before(early) {
			key = k
			early = v.timeExpiry
			f = v.cacheFunc
		}
	}

	return key, f
}

// GetCacheSize возвращает общее количество объектов в кэше
func (c *CacheExecutedObjects[T]) GetCacheSize() int {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	return len(c.cache.storages)
}

// SetTimeExpiry устанавливает или обновляет значение параметра timeExpiry
func (c *CacheExecutedObjects[T]) SetTimeExpiry(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	if storage, ok := c.cache.storages[key]; ok {
		storage.timeExpiry = time.Now().Add(c.maxTtl)
		c.cache.storages[key] = storage
	}
}

// SetIsExecutionTrue устанавливает значение параметра isExecution
func (c *CacheExecutedObjects[T]) SetIsExecutionTrue(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	if storage, ok := c.cache.storages[key]; ok {
		storage.isExecution = true
		c.cache.storages[key] = storage
	}
}

// SetIsExecutionFalse устанавливает значение параметра isExecution
func (c *CacheExecutedObjects[T]) SetIsExecutionFalse(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	if storage, ok := c.cache.storages[key]; ok {
		storage.isExecution = false
		c.cache.storages[key] = storage
	}
}

// SetIsCompletedSuccessfullyTrue устанавливает значение параметра isCompletedSuccessfully
func (c *CacheExecutedObjects[T]) SetIsCompletedSuccessfullyTrue(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	if storage, ok := c.cache.storages[key]; ok {
		storage.isCompletedSuccessfully = true
		c.cache.storages[key] = storage
	}
}

// SetIsCompletedSuccessfullyFalse устанавливает значение параметра isCompletedSuccessfully
func (c *CacheExecutedObjects[T]) SetIsCompletedSuccessfullyFalse(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	if storage, ok := c.cache.storages[key]; ok {
		storage.isCompletedSuccessfully = false
		c.cache.storages[key] = storage
	}
}

// DeleteForTimeExpiryObjectFromCache удаляет все объекты у которых истекло время жизни
func (c *CacheExecutedObjects[T]) DeleteForTimeExpiryObjectFromCache() {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	for key, storage := range c.cache.storages {
		if storage.timeExpiry.Before(time.Now()) {
			delete(c.cache.storages, key)
		}
	}
}

// getOldestObjectFromCache возвращает индекс самого старого объекта
func (c *CacheExecutedObjects[T]) getOldestObjectFromCache() string {
	var (
		index      string
		timeExpiry time.Time
	)

	for k, v := range c.cache.storages {
		if index == "" {
			index = k
			timeExpiry = v.timeExpiry

			continue
		}

		if v.timeExpiry.Before(timeExpiry) {
			index = k
			timeExpiry = v.timeExpiry
		}
	}

	return index
}

// deleteOldestObjectFromCache удаляет самый старый объект по timeMain
func (c *CacheExecutedObjects[T]) deleteOldestObjectFromCache() {
	delete(c.cache.storages, c.getOldestObjectFromCache())
}

//*********** Медоды необходимые для выполнения дополнительного тестирования ************

// AddObjectToCache_TestTimeExpiry добавляет новый объект в хранилище (только для теста)
func (c *CacheExecutedObjects[T]) AddObjectToCache_TestTimeExpiry(key string, timeExpiry time.Time, value CacheStorageFuncHandler[T]) error {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	if len(c.cache.storages) >= c.cache.maxSize {
		//удаление самого старого объекта, осуществляется по параметру timeMain
		c.deleteOldestObjectFromCache()
	}

	storage, ok := c.cache.storages[key]
	if !ok {
		c.cache.storages[key] = storageParameters[T]{
			timeMain:       time.Now(),
			timeExpiry:     timeExpiry,
			originalObject: value.GetObject(),
			cacheFunc:      value.GetFunc(),
		}

		return nil
	}

	//найден объект у которого ключ совпадает с объектом принятом в обработку

	//объект в настоящее время выполняется
	if storage.isExecution {
		return fmt.Errorf("an object has been received whose key ID '%s' matches the already running object, ignore it", key)
	}

	//сравнение объектов из кэша и полученного из очереди
	if value.Comparison(storage.originalObject) {
		return fmt.Errorf("objects with key ID '%s' are completely identical, adding an object to the cache is not performed", key)
	}

	storage.timeMain = time.Now()
	storage.timeExpiry = timeExpiry
	storage.isExecution = false
	storage.isCompletedSuccessfully = false
	storage.originalObject = value.GetObject()
	storage.cacheFunc = value.GetFunc()

	c.cache.storages[key] = storage

	return nil
}

// CleanCache_Test очищает кэш
func (c *CacheExecutedObjects[T]) CleanCache_Test() {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	c.cache.storages = map[string]storageParameters[T]{}
}
