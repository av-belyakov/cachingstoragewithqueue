package cachingstoragewithqueue

import (
	"fmt"
	"time"
)

// GetSizeObjectToQueue размер очереди
func (c *CacheStorageWithQueue[T]) GetSizeObjectToQueue() int {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	return len(c.queue.storages)
}

// CleanQueue очистка очереди
func (c *CacheStorageWithQueue[T]) CleanQueue() {
	c.queue.mutex.Lock()
	defer c.queue.mutex.Unlock()

	c.queue.storages = []CacheStorageFuncHandler[T](nil)
}

// CleanCache очистка кэша
func (c *CacheStorageWithQueue[T]) CleanCache() {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	c.cache.storages = map[string]storageParameters[T]{}
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

// AddObjectToCache добавляет новый объект в кэш
func (c *CacheStorageWithQueue[T]) AddObjectToCache(key string, value CacheStorageFuncHandler[T]) error {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

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

	//если объекты с одним и тем же ключём разные, заменяем объект в кэше более новым
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

	fmt.Println("||| func 'GetObjectFromCacheByKey', cache list:")
	for k := range c.cache.storages {
		fmt.Println("ID:", k)
	}

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
// не выполняется, не была успешно выполнена и время истечения жизни объекта которой самое меньшее,
// то есть фактически, ищется наиболее старые объекты
func (c *CacheStorageWithQueue[T]) GetObjectFromCacheMinTimeExpiry() (key string, obj T) {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	var early time.Time
	for k, v := range c.cache.storages {
		if v.isExecution || v.isCompletedSuccessfully {
			continue
		}

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
// не выполняется, не была успешно выполнена и время истечения жизни объекта которой самое меньшее,
// то есть фактически, ищется наиболее старые объекты
func (c *CacheStorageWithQueue[T]) GetFuncFromCacheMinTimeExpiry() (key string, f func(int) bool) {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	//тут не выполняется проверка для найденного самого старого объекта
	// выполняется ли функция в настоящее время
	//и была ли функция успешно выполнена

	var early time.Time
	for k, v := range c.cache.storages {
		if v.isExecution || v.isCompletedSuccessfully {
			continue
		}

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

	fmt.Println("||| func 'GetCacheSize', cache list:")
	for k := range c.cache.storages {
		fmt.Println("ID:", k)
	}

	return len(c.cache.storages)
}

// GetIndexesWithIsExecutionStatus возвращает список индексов объектов, по которым выполяется обработка
func (c *CacheStorageWithQueue[T]) GetIndexesWithIsExecutionStatus() []string {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	var indexes []string

	for index, storage := range c.cache.storages {
		if !storage.isExecution {
			continue
		}

		indexes = append(indexes, index)
	}

	return indexes
}

// GetIndexesWithIsCompletedSuccessfully возвращает список индексов объектов, которые были успешно выполнены
func (c *CacheStorageWithQueue[T]) GetIndexesWithIsCompletedSuccessfully() []string {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	var indexes []string

	for index, storage := range c.cache.storages {
		if !storage.isCompletedSuccessfully {
			continue
		}

		indexes = append(indexes, index)
	}

	return indexes
}

// SetTimeExpiry устанавливает или обновляет значение параметра timeExpiry
func (c *CacheStorageWithQueue[T]) SetTimeExpiry(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	c.setTimeExpiry(key)
}

// GetIsExecution возвращает статус параметра isExecution объекта в кэше и найден ли такой объект по ключу
func (c *CacheStorageWithQueue[T]) GetIsExecution(key string) (status bool, isExist bool) {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	sp, ok := c.getStorageParameters(key)
	status = sp.isExecution
	isExist = ok

	return
}

// SetIsExecutionTrue устанавливает значение параметра isExecution
func (c *CacheStorageWithQueue[T]) SetIsExecutionTrue(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	c.setIsExecutionTrue(key)
}

// SetIsExecutionFalse устанавливает значение параметра isExecution
func (c *CacheStorageWithQueue[T]) SetIsExecutionFalse(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	c.setIsExecutionFalse(key)
}

// GetIsCompletedSuccessfully возвращает статус параметра isCompletedSuccessfully объекта в кэше и найден ли такой объект по ключу
func (c *CacheStorageWithQueue[T]) GetIsCompletedSuccessfully(key string) (status bool, isExist bool) {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	sp, ok := c.getStorageParameters(key)
	status = sp.isCompletedSuccessfully
	isExist = ok

	return
}

// SetIsCompletedSuccessfullyTrue устанавливает значение параметра isCompletedSuccessfully
func (c *CacheStorageWithQueue[T]) SetIsCompletedSuccessfullyTrue(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	c.setIsCompletedSuccessfullyTrue(key)
}

// SetIsCompletedSuccessfullyFalse устанавливает значение параметра isCompletedSuccessfully
func (c *CacheStorageWithQueue[T]) SetIsCompletedSuccessfullyFalse(key string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	c.setIsCompletedSuccessfullyFalse(key)
}

// GetNumberExecutionAttempts возвращает количество попыток выполнения функции
func (c *CacheStorageWithQueue[T]) GetNumberExecutionAttempts(key string) (num int, status bool) {
	c.cache.mutex.RLock()
	defer c.cache.mutex.RUnlock()

	if storage, ok := c.cache.storages[key]; ok {
		num = storage.numberExecutionAttempts
		status = ok
	}

	return num, status
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

// DeleteOldestObjectFromCache поиск и удаление самого старого объекта в кэше
func (c *CacheStorageWithQueue[T]) DeleteOldestObjectFromCache() error {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	//получаем самый старый объект в кэше
	index := c.getOldestObjectFromCache()
	if storage, ok := c.cache.storages[index]; ok {
		fmt.Printf("func 'DeleteOldestObjectFromCache', isExecution:'%t', isCompletedSuccessfully:'%t', numberExecutionAttempts:'%d'\n", storage.isExecution, storage.isCompletedSuccessfully, storage.numberExecutionAttempts)

		if storage.isExecution == false && (storage.isCompletedSuccessfully == true || storage.numberExecutionAttempts > 3) {
			fmt.Println("func 'DeleteOldestObjectFromCache' - DELETE")

			delete(c.cache.storages, index)
		} else {
			return fmt.Errorf("the object with id '%s' cannot be deleted, it may be in progress", index)
		}
	}

	return nil
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

// getStorageParameters получить общие параметры объекта из кэша
func (c *CacheStorageWithQueue[T]) getStorageParameters(key string) (storageParameters[T], bool) {
	if storage, ok := c.cache.storages[key]; ok {
		return storage, true
	}

	return storageParameters[T]{}, false
}

// deleteOldestObjectFromCache удаляет самый старый объект по timeMain
// без учета других параметров, таких как isCompletedSuccessfully и isExecution
func (c *CacheStorageWithQueue[T]) deleteOldestObjectFromCache() {
	delete(c.cache.storages, c.getOldestObjectFromCache())
}

// setTimeExpiry устанавливает или обновляет значение параметра timeExpiry
func (c *CacheStorageWithQueue[T]) setTimeExpiry(key string) {
	if storage, ok := c.cache.storages[key]; ok {
		storage.timeExpiry = time.Now().Add(c.maxTtl)
		c.cache.storages[key] = storage
	}
}

// setIsExecutionTrue устанавливает значение параметра isExecution
func (c *CacheStorageWithQueue[T]) setIsExecutionTrue(key string) {
	if storage, ok := c.cache.storages[key]; ok {
		storage.isExecution = true
		c.cache.storages[key] = storage
	}
}

// setIsExecutionFalse устанавливает значение параметра isExecution
func (c *CacheStorageWithQueue[T]) setIsExecutionFalse(key string) {
	if storage, ok := c.cache.storages[key]; ok {
		storage.isExecution = false
		c.cache.storages[key] = storage
	}
}

// setIsCompletedSuccessfullyTrue устанавливает значение параметра isCompletedSuccessfully
func (c *CacheStorageWithQueue[T]) setIsCompletedSuccessfullyTrue(key string) {
	if storage, ok := c.cache.storages[key]; ok {
		storage.isCompletedSuccessfully = true
		c.cache.storages[key] = storage
	}
}

// setIsCompletedSuccessfullyFalse устанавливает значение параметра isCompletedSuccessfully
func (c *CacheStorageWithQueue[T]) setIsCompletedSuccessfullyFalse(key string) {
	if storage, ok := c.cache.storages[key]; ok {
		storage.isCompletedSuccessfully = false
		c.cache.storages[key] = storage
	}
}

// increaseNumberExecutionAttempts увеличивает количество попыток выполнения функции
func (c *CacheStorageWithQueue[T]) increaseNumberExecutionAttempts(key string) {
	if storage, ok := c.cache.storages[key]; ok {
		storage.numberExecutionAttempts = storage.numberExecutionAttempts + 1
		c.cache.storages[key] = storage
	}
}

// syncExecution выполняет синхронную обработку функций из кэша
func (csq *CacheStorageWithQueue[T]) syncExecution() error {
	//******* test *****
	//fmt.Println("func 'syncExecution', START...")
	//******************

	//проверяем, вообще что либо в настоящий момент выполняется, если да, ожидание завершения
	if len(csq.GetIndexesWithIsExecutionStatus()) > 0 {
		//******* test *****
		fmt.Println("func 'syncExecution', GetIndexesWithIsExecutionStatus > 0, EXIT, csq.GetIndexesWithIsExecutionStatus() =", csq.GetIndexesWithIsExecutionStatus())
		//******************

		return nil
	}

	currentObject, isEmpty := csq.PullObjectFromQueue()
	// если очередь с объектами для обработки не пуста
	if !isEmpty {
		//******* test *****
		fmt.Println("func 'syncExecution', Add Object To Cache, ID:", currentObject.GetID())
		//******************

		if err := csq.AddObjectToCache(currentObject.GetID(), currentObject); err != nil {
			return err
		}
	}

	//проверяем, есть ли вообще что либо в кэше для обработки
	if csq.GetCacheSize() == 0 {
		//******* test *****
		//fmt.Println("func 'syncExecution', CacheSize == 0, EXIT")
		//******************

		return nil
	}

	fmt.Println("func 'syncExecution', csq.GetCacheSize() =", csq.GetCacheSize())

	//получаем самую старую функцию, которая не выполняется или не была выполнена успешно
	index, f := csq.GetFuncFromCacheMinTimeExpiry()

	//******* test *****
	fmt.Println("func 'syncExecution' get the oldest function with index:", index)
	numAttempt, status := csq.GetNumberExecutionAttempts(index)
	fmt.Println("func 'syncExecution', count attempt:", numAttempt, " status:", status)
	//******************

	csq.cache.mutex.Lock()
	defer csq.cache.mutex.Unlock()

	//функция для данного объекта выполняется
	csq.setIsExecutionTrue(index)
	// увеличиваем количество попыток выполнения функции
	csq.increaseNumberExecutionAttempts(index)

	if f(0) {
		csq.setIsCompletedSuccessfullyTrue(index)
	} else {
		csq.setIsCompletedSuccessfullyFalse(index)
	}

	//функция не обрабатывается
	csq.setIsExecutionFalse(index)

	//******* test *****
	//if storage, ok := csq.cache.storages[index]; ok {
	//	fmt.Println("func 'syncExecution', IsExecution:", storage.isExecution, " isCompletedSuccessfully:", storage.isCompletedSuccessfully)
	//}
	//******************

	return nil
}

// asyncExecution выполняет асинхронную обработку функций из кэша
func (csq *CacheStorageWithQueue[T]) asyncExecution() {

}

// Write метод-заглушка реализуемая в конструкторе CacheStorageWithQueue 'по умолчанию'
// если при инициализации конструктора не была добавлена опция WithLogging
func (wl *writeLog) Write(msgType, msg string) bool {
	fmt.Printf("package 'CacheStorageWithQueue', default logging: %s - %s\n", msgType, msg)

	return true
}

//*********** Методы необходимые для выполнения дополнительного тестирования ************

func (csq *CacheStorageWithQueue[T]) GetCacheMaxSize_Test() int {
	return csq.cache.maxSize
}

// SyncExecution_Test выполняет синхронную обработку функций из кэша (только для теста)
func (csq *CacheStorageWithQueue[T]) SyncExecution_Test() {
	csq.syncExecution()
}

// AddObjectToCache_Test добавляет новый объект в хранилище (только для теста)
func (c *CacheStorageWithQueue[T]) AddObjectToCache_Test(key string, timeExpiry time.Time, value CacheStorageFuncHandler[T]) error {
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
