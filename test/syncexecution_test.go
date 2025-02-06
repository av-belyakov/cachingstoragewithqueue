package cachingstoragewithqueue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/cachingstoragewithqueue"
	"github.com/av-belyakov/cachingstoragewithqueue/examples"
	"github.com/av-belyakov/objectsmispformat"
)

func TestSyncExecution(t *testing.T) {
	cache, err := cachingstoragewithqueue.NewCacheStorage[*objectsmispformat.ListFormatsMISP](
		context.Background(),
		cachingstoragewithqueue.WithMaxTtl[*objectsmispformat.ListFormatsMISP](60),
		cachingstoragewithqueue.WithTimeTick[*objectsmispformat.ListFormatsMISP](3),
		cachingstoragewithqueue.WithMaxSize[*objectsmispformat.ListFormatsMISP](10))
	assert.NoError(t, err)

	listIdOne := []string{
		"0000-0000",
		"1111-1111",
		"2222-2222",
		"3333-3333",
		"4444-4444",
		"5555-5555",
		"6666-6666",
		"7777-7777",
		"8888-8888",
		"9999-9999",
	}

	/*listIdTwo := []string{
		"aaaa-1111",
		"bbbb-2222",
		"cccc-3333",
		"dddd-4444",
		"eeee-5555",
		"ffff-6666",
		"gggg-7777",
		"hhhh-8888",
		"iiii-9999",
		"ssss-0000",
	}*/

	//добавление в очередь новых объектов
	addObjectToQueue := func(lid []string) {
		for _, id := range lid {
			soc := examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
			//для примера используем конструктор списка форматов MISP
			objectTemplate := objectsmispformat.NewListFormatsMISP()
			objectTemplate.ID = id

			//t.Log("Object ID:", objectTemplate.GetID())

			//заполняем вспомогательный тип
			soc.SetID(objectTemplate.GetID())
			soc.SetObject(objectTemplate)
			soc.SetFunc(func(int) bool {
				//здесь некий обработчик...
				//в контексе работы с MISP здесь должен быть код отвечающий
				//за REST запросы к серверу MISP
				fmt.Println("function with ID:", soc.GetID())

				return true
			})
			cache.PushObjectToQueue(soc)
		}
	}

	t.Run("Тест 1, в результате которого все объекты должны быть помечены как выполненные", func(t *testing.T) {
		//кладем в очередь объекты которые необходимо обработать
		addObjectToQueue(listIdOne)
		for {
			if cache.GetSizeObjectToQueue() == 0 {
				break
			}

			//кладем в кэш все объекты из очереди
			cache.SyncExecution_Test()
		}

		assert.Equal(t, len(cache.GetIndexesWithIsCompletedSuccessfully()), len(listIdOne))

		//очищаем кэш
		cache.CleanCache()
	})

	t.Run("Тест 2, добавление новых объектов из очереди не производится если в кэше есть хотя бы одно активное задание", func(t *testing.T) {
		//очередь пуста
		assert.Equal(t, cache.GetSizeObjectToQueue(), 0)

		soc := examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		//для примера используем конструктор списка форматов MISP
		objectTemplate := objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "1234-5678"
		//заполняем вспомогательный тип
		soc.SetID(objectTemplate.GetID())
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			//здесь некий обработчик...
			//в контексе работы с MISP здесь должен быть код отвечающий
			//за REST запросы к серверу MISP
			//fmt.Println("function with ID:", soc.GetID())

			return false
		})
		cache.PushObjectToQueue(soc)
		obj, isEmpty := cache.PullObjectFromQueue()
		assert.False(t, isEmpty)

		//добавляем один объект в кэш
		err := cache.AddObjectToCache(obj.GetID(), obj)
		assert.NoError(t, err)

		//один объект в кэше
		assert.Equal(t, cache.GetCacheSize(), 1)

		// кладем в очередь объекты которые необходимо обработать
		addObjectToQueue(listIdOne)
		//размер очереди должен быть 10 объектов
		assert.Equal(t, cache.GetSizeObjectToQueue(), 10)

		//меняем статус состояния объекта isExecute = true
		cache.SetIsExecutionTrue(obj.GetID())

		var num int
		for {
			num++

			if num == 4 {
				//меняем статус состояния объекта isExecute = true
				cache.SetIsExecutionFalse(obj.GetID())
			}

			if cache.GetSizeObjectToQueue() == 0 {
				break
			}

			//поиск и удаление самого старого объекта если размер кэша достиг максимального значения
			//выполняется удаление объекта который в настоящее время не выполняеться и ранее был успешно выполнен
			if cache.GetCacheSize() == cache.GetCacheMaxSize_Test() {
				if err := cache.DeleteOldestObjectFromCache(); err != nil {
					t.Log("Delete:", err)
				}

				continue
			}

			cache.SyncExecution_Test()
		}

		//проверяем количество добавленых объектов, ни один из объектов не должен быть
		// добавлен в обработку, так как в кэше находится объект в статусе обработки
		assert.Equal(t, cache.GetCacheSize(), 10)

		//добавлен список только с 9 элементами не включая последний элемент списка listId
		o, isExist := cache.GetObjectFromCacheByKey(listIdOne[len(listIdOne)-1])
		t.Log("___ object ___", o)
		assert.True(t, isExist)
	})

	/*t.Run("Тест 3, добавление объектов производится ни один объект в кэше не участвует в обработке", func(t *testing.T) {
		// кладем в очередь объекты которые необходимо обработать
		addObjectToQueue(listIdTwo)

		for {
			//поиск и удаление самого старого объекта если размер кэша достиг максимального значения
			//выполняется удаление объекта который в настоящее время не выполняеться и ранее был успешно выполнен
			if cache.GetCacheSize() >= cache.GetCacheMaxSize_Test() {
				if err := cache.DeleteOldestObjectFromCache(); err != nil {
					t.Log("error:", err)
				}

				continue
			}

			if cache.GetSizeObjectToQueue() == 0 {
				break
			}

			cache.SyncExecution_Test()
		}

		t.Log("--------", cache.GetCacheSize())
		//так как размер кеша 10 объектов, в кэше уже есть 1 элемент, а список добавляемых
		//объектов состоит из 10, то при добавлении самый старый объект (первый из добавленых)
		//должен быть удалён
		assert.Equal(t, cache.GetCacheSize(), 10)
	})*/
}
