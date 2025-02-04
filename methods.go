package cachingstoragewithqueue

import (
	"fmt"
	"time"
)

// SizeObjectToQueue размер очереди
func (c *CacheStorageWithQueue[T]) SizeObjectToQueue() int {
	return len(c.queue.storages)
}

// CleanQueue очистка очереди
func (c *CacheStorageWithQueue[T]) CleanQueue() {
	c.queue.mutex.Lock()
	defer c.queue.mutex.Unlock()

	c.queue.storages = []CacheStorageFuncHandler[T](nil)
}

// PushObjectToQueue добавляет в очередь объектов новый объект
func (c *CacheStorageWithQueue[T]) PushObjectToQueue(v CacheStorageFuncHandler[T]) {
	c.queue.mutex.Lock()
	defer c.queue.mutex.Unlock()

	c.queue.storages = append(c.queue.storages, v)
}

// PullObjectFromQueue забирает из очереди новый объект или возвращает TRUE если очередь пуста
func (c *CacheStorageWithQueue[T]) PullObjectFromQueue() (CacheStorageFuncHandler[T], bool) {
	c.queue.mutex.Lock()
	defer c.queue.mutex.Unlock()

	var obj CacheStorageFuncHandler[T]
	size := len(c.queue.storages)
	if size == 0 {
		return obj, true
	}

	obj = c.queue.storages[0]

	if size == 1 {
		c.queue.storages = make([]CacheStorageFuncHandler[T], 0)

		return obj, false
	}

	c.queue.storages = c.queue.storages[1:]

	return obj, false
}

// AddObjectToCache добавляет новый объект в хранилище, при этом выполняются следующие действия:
// - сравнение размера кэша с параметром сache.maxSize
// - при достижении максимального размера кэша, удаление самого старого объекта (поиск по timeMain)
// - проверка, есть ли уже в кэше объект с таким же значением ключа и выполняется ли он
// - поиск объектов в кэше которые имеют тот же id ключа что и объект из сигнатуры функции
// - сравнение, по ключу, объектов из кэша и полученного из очереди
func (c *CacheStorageWithQueue[T]) AddObjectToCache(key string, value CacheStorageFuncHandler[T]) error {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	//если размер кэша достиг максимального значения, то выполнить поиск и удаление самого старого объекта
	//который в настоящее время не выполняеться и ранее был успешно выполнен
	if len(c.cache.storages) >= c.cache.maxSize {
		//получаем самый старый объект в кэше
		index := c.getOldestObjectFromCache()
		if storage, ok := c.cache.storages[index]; ok {
			if storage.isExecution == false && storage.isCompletedSuccessfully == true {
				delete(c.cache.storages, index)
			} else {
				return fmt.Errorf("the object with id '%s' cannot be deleted, it may be in progress or has not been completed successfully", index)
			}
		}
	}

	//если поиск подобного объекта по ключу не дал результатов то просто добавляем объект
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

	//сравнение объектов из кэша и полученного из очереди, если они одинаковые то ошибка
	if value.Comparison(storage.originalObject) {
		return fmt.Errorf("objects with key ID '%s' are completely identical, adding an object to the cache is not performed", key)
	}

	//если объекты с одним и тем же ключём разные, заменяем оббъект в кэше более новым
	storage.timeMain = time.Now()
	storage.timeExpiry = time.Now().Add(c.maxTtl)
	storage.isExecution = false
	storage.isCompletedSuccessfully = false
	storage.originalObject = value.GetObject()
	storage.cacheFunc = value.GetFunc()

	//добавление нового объекта в кэш
	c.cache.storages[key] = storage

	return nil
}

// GetOldestObjectFromCache возвращает индекс самого старого объекта
func (c *CacheStorageWithQueue[T]) GetOldestObjectFromCache() string {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	return c.getOldestObjectFromCache()
}

// GetObjectFromCacheByKey возвращает объект из кэша по ключу
func (c *CacheStorageWithQueue[T]) GetObjectFromCacheByKey(key string) (T, bool) {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	storage, ok := c.cache.storages[key]

	return storage.originalObject, ok
}

// GetFuncFromCacheByKey возвращает исполняемую функцию из кэша по ключу
func (c *CacheStorageWithQueue[T]) GetFuncFromCacheByKey(key string) (func(int) bool, bool) {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	storage, ok := c.cache.storages[key]

	return storage.cacheFunc, ok
}

// GetObjectFromCacheMinTimeExpiry возвращает из кэша объект, функция которого в настоящее время
// не выполняется, не была успешно выполнена и время истечения жизни объекта которой самое меньшее
// то есть фактически ищется наиболее старые объекты
func (c *CacheStorageWithQueue[T]) GetObjectFromCacheMinTimeExpiry() (key string, obj T) {
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
// то есть фактически ищется наиболее старые объекты
func (c *CacheStorageWithQueue[T]) GetFuncFromCacheMinTimeExpiry() (key string, f func(int) bool) {
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
func (c *CacheStorageWithQueue[T]) GetCacheSize() int {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	return len(c.cache.storages)
}

// SetTimeExpiry устанавливает или обновляет значение параметра timeExpiry
func (c *CacheStorageWithQueue[T]) SetTimeExpiry(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	if storage, ok := c.cache.storages[key]; ok {
		storage.timeExpiry = time.Now().Add(c.maxTtl)
		c.cache.storages[key] = storage
	}
}

// SetIsExecutionTrue устанавливает значение параметра isExecution
func (c *CacheStorageWithQueue[T]) SetIsExecutionTrue(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	if storage, ok := c.cache.storages[key]; ok {
		storage.isExecution = true
		c.cache.storages[key] = storage
	}
}

// SetIsExecutionFalse устанавливает значение параметра isExecution
func (c *CacheStorageWithQueue[T]) SetIsExecutionFalse(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	if storage, ok := c.cache.storages[key]; ok {
		storage.isExecution = false
		c.cache.storages[key] = storage
	}
}

// SetIsCompletedSuccessfullyTrue устанавливает значение параметра isCompletedSuccessfully
func (c *CacheStorageWithQueue[T]) SetIsCompletedSuccessfullyTrue(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	if storage, ok := c.cache.storages[key]; ok {
		storage.isCompletedSuccessfully = true
		c.cache.storages[key] = storage
	}
}

// SetIsCompletedSuccessfullyFalse устанавливает значение параметра isCompletedSuccessfully
func (c *CacheStorageWithQueue[T]) SetIsCompletedSuccessfullyFalse(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	if storage, ok := c.cache.storages[key]; ok {
		storage.isCompletedSuccessfully = false
		c.cache.storages[key] = storage
	}
}

// DeleteForTimeExpiryObjectFromCache удаляет все объекты у которых истекло время жизни, без учета других параметров
func (c *CacheStorageWithQueue[T]) DeleteForTimeExpiryObjectFromCache() {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	for key, storage := range c.cache.storages {
		if storage.timeExpiry.Before(time.Now()) {
			delete(c.cache.storages, key)
		}
	}
}

// getOldestObjectFromCache возвращает индекс самого старого объекта
func (c *CacheStorageWithQueue[T]) getOldestObjectFromCache() string {
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
// без учета других параметров, таких как isCompletedSuccessfully и isExecution
func (c *CacheStorageWithQueue[T]) deleteOldestObjectFromCache() {
	delete(c.cache.storages, c.getOldestObjectFromCache())
}

//*********** Методы необходимые для выполнения дополнительного тестирования ************

// AddObjectToCache_TestTimeExpiry добавляет новый объект в хранилище (только для теста)
func (c *CacheStorageWithQueue[T]) AddObjectToCache_TestTimeExpiry(key string, timeExpiry time.Time, value CacheStorageFuncHandler[T]) error {
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
func (c *CacheStorageWithQueue[T]) CleanCache_Test() {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	c.cache.storages = map[string]storageParameters[T]{}
}
