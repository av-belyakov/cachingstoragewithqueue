package cachingstoragewithqueue

import (
	"context"
	"fmt"

	"github.com/av-belyakov/cachingstoragewithqueue/internal/supportingfunctions"
)

// syncExecution выполняет синхронную обработку функций из кэша
func (c *CacheStorageWithQueue[T]) syncExecution(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	//проверяем, вообще что либо в настоящий момент выполняется, если да, ожидание завершения
	if len(c.GetIndexesWithIsExecutionStatus()) > 0 {
		return
	}

	currentObject, isEmpty := c.PullObjectFromQueue()
	// если очередь с объектами для обработки не пуста
	if !isEmpty {
		if err := c.AddObjectToCache(currentObject.GetID(), currentObject); err != nil {
			c.logging.Write("warning", supportingfunctions.CustomError(fmt.Errorf("cachingstoragewithqueue package: '%s'", err.Error())).Error())

			return
		}
	}

	//проверяем, есть ли вообще что либо в кэше для обработки
	if c.GetCacheSize() == 0 {
		return
	}

	//получаем самую старую функцию, которая не выполняется или не была выполнена успешно
	index, f := c.GetFuncFromCacheMinTimeExpiry()
	if index == "" {
		return
	}

	c.cache.mutex.Lock()
	//функция для данного объекта выполняется
	c.setIsExecutionTrue(index)
	// увеличиваем количество попыток выполнения функции
	c.increaseNumberExecutionAttempts(index)
	c.cache.mutex.Unlock()

	//выполняем функцию и изменяем состояние задачи
	//меняется 'execution' на false, а успешность выполнения
	//задачи на значение полученное от функции
	c.ChangeValues(index, f(0))
}

// asyncExecution выполняет асинхронную обработку функций из кэша
func (c *CacheStorageWithQueue[T]) asyncExecution(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	listIndexes := c.GetIndexesWithIsExecutionStatus()

	//проверяем, количество выполняемых функций соответствует максимальному количеству
	// одновременно выполняемых задач (параметр задаётся в опциях)
	if len(listIndexes) >= c.isAsync {
		return
	}

	count := c.isAsync - len(listIndexes)
	pushObjectToCache := func(count int) []string {
		indexes := make([]string, 0, count)

		for i := 0; i < count; i++ {
			if c.GetCacheSize() >= c.cache.maxSize {
				return indexes
			}

			object, isEmpty := c.PullObjectFromQueue()
			if isEmpty {
				return indexes
			}

			if err := c.AddObjectToCache(object.GetID(), object); err != nil {
				c.logging.Write("warning", supportingfunctions.CustomError(fmt.Errorf("cachingstoragewithqueue package: '%s'", err.Error())).Error())
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

		go func(ind string) {
			c.ChangeValues(ind, f(0))
		}(index)
	}
}
