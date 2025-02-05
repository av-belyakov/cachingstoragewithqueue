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

	listId := []string{
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

	/*
			не нармальное поведение, надо разобратся

		=== RUN   TestSyncExecution
		=== RUN   TestSyncExecution/Тест,_в_результате_которого_все_объекты_должны_быть_помечены_как_выполненные
		func 'syncExecution', START...
		func 'syncExecution', Add Object To Cache, ID: 1111-1111
		func 'syncExecution' get the oldest function with index: 1111-1111
		function with ID: 1111-1111
		func 'syncExecution', END...
		func 'syncExecution', START...
		func 'syncExecution', Add Object To Cache, ID: 2222-2222
		func 'syncExecution' get the oldest function with index: 1111-1111
		function with ID: 1111-1111
		func 'syncExecution', END...
		func 'syncExecution', START...
		func 'syncExecution', Add Object To Cache, ID: 3333-3333
		func 'syncExecution' get the oldest function with index: 1111-1111
		function with ID: 1111-1111
		func 'syncExecution', END...
		func 'syncExecution', START...
		func 'syncExecution', Add Object To Cache, ID: 4444-4444
		func 'syncExecution' get the oldest function with index: 1111-1111
		function with ID: 1111-1111
		func 'syncExecution', END...
		func 'syncExecution', START...
		func 'syncExecution', Add Object To Cache, ID: 5555-5555
		func 'syncExecution' get the oldest function with index: 1111-1111
		function with ID: 1111-1111
		func 'syncExecution', END...
		func 'syncExecution', START...
		func 'syncExecution', Add Object To Cache, ID: 6666-6666
		func 'syncExecution' get the oldest function with index: 1111-1111
	*/

	t.Run("Тест, в результате которого все объекты должны быть помечены как выполненные", func(t *testing.T) {
		//кладем в очередь объекты которые необходимо обработать
		addObjectToQueue(listId)

		for {
			if cache.GetSizeObjectToQueue() == 0 {
				break
			}

			cache.SyncExecution_Test()
		}

		assert.Equal(t, len(cache.GetIndexesWithIsCompletedSuccessfully()), len(listId))
		cache.CleanCache()
	})

	t.Run("Тест, добавление новых объектов из очереди не производится если в кэше есть хотя бы одно активное задание", func(t *testing.T) {
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
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)
		obj, isEmpty := cache.PullObjectFromQueue()
		assert.False(t, isEmpty)

		//добавляем один объект в кэш
		err := cache.AddObjectToCache(obj.GetID(), obj)
		assert.NoError(t, err)

		//меняем статус состояния объекта isExecute = true
		cache.SetIsExecutionTrue(obj.GetID())

		//пытаемся добавить новые объекты для обработки
		addObjectToQueue(listId)

		//проверяем количество добавленых объектов, ни один из объектов не должен быть
		// добавлен в обработку, так как в кэше находится объект в статусе обработки
		assert.Equal(t, cache.GetCacheSize(), 1)
	})
}
