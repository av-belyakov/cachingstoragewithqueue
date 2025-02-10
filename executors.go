package cachingstoragewithqueue

import (
	"fmt"
	"runtime"
)

// syncExecution выполняет синхронную обработку функций из кэша
func (c *CacheStorageWithQueue[T]) syncExecution(chStop chan<- HandlerOptionsStoper) {
	//проверяем, вообще что либо в настоящий момент выполняется, если да, ожидание завершения
	if len(c.GetIndexesWithIsExecutionStatus()) > 0 {
		return
	}

	currentObject, isEmpty := c.PullObjectFromQueue()
	// если очередь с объектами для обработки не пуста
	if !isEmpty {
		if err := c.AddObjectToCache(currentObject.GetID(), currentObject); err != nil {
			_, f, l, _ := runtime.Caller(0)
			c.logging.Write("error", fmt.Sprintf("cachingstoragewithQueue package: '%s' %s:%d", err.Error(), f, l-1))

			return
		}
	}

	//проверяем, есть ли вообще что либо в кэше для обработки
	if c.GetCacheSize() == 0 {
		return
	}

	//получаем самую старую функцию, которая не выполняется или не была выполнена успешно
	index, f := c.GetFuncFromCacheMinTimeExpiry()

	c.cache.mutex.Lock()
	//функция для данного объекта выполняется
	c.setIsExecutionTrue(index)
	// увеличиваем количество попыток выполнения функции
	c.increaseNumberExecutionAttempts(index)
	c.cache.mutex.Unlock()

	sho := NewStopHandlerOptions()
	sho.SetIndex(index)
	sho.SetIsSuccess(f(0))

	chStop <- sho
}

// asyncExecution выполняет асинхронную обработку функций из кэша
func (c *CacheStorageWithQueue[T]) asyncExecution(chStop chan<- HandlerOptionsStoper) {
	fmt.Println("func 'asyncExecution', START...")

	listIndexes := c.GetIndexesWithIsExecutionStatus()

	//проверяем, количество выполняемых функций соответствует максимальному количеству
	// одновременно выполняемых задач (параметр задаётся в опциях)
	if len(listIndexes) >= c.isAsync {
		return
	}

	count := c.isAsync - len(listIndexes)

	fmt.Println("func 'asyncExecution', count:", count)

	pushObjectToCache := func(count int) []string {
		indexes := make([]string, 0, count)

		for i := 0; i < count; i++ {
			object, isEmpty := c.PullObjectFromQueue()
			if isEmpty {
				return indexes
			}

			if err := c.AddObjectToCache(object.GetID(), object); err != nil {
				_, f, l, _ := runtime.Caller(0)
				c.logging.Write("error", fmt.Sprintf("cachingstoragewithQueue package: '%s' %s:%d", err.Error(), f, l-1))
			} else {
				indexes = append(indexes, object.GetID())
			}
		}

		return indexes
	}

	//список индексов объекты которых были добавлены в кэш
	indexes := pushObjectToCache(count)
	if len(indexes) == 0 {
		return
	}

	fmt.Println("func 'asyncExecution', indexes:", indexes)

	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	for _, index := range indexes {
		f, isExist := c.getFuncFromCacheByKey(index)
		if !isExist {
			continue
		}

		//функция для данного объекта выполняется
		c.setIsExecutionTrue(index)
		// увеличиваем количество попыток выполнения функции
		c.increaseNumberExecutionAttempts(index)

		go func() {
			chStop <- &stopHandlerOptions{
				index:     index,
				isSuccess: f(0),
			}
		}()
	}
}
