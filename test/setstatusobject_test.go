package cachingstoragewithqueue_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/cachingstoragewithqueue"
	"github.com/av-belyakov/cachingstoragewithqueue/examples"
	"github.com/av-belyakov/objectsmispformat"
)

func TestSetStatusObject(t *testing.T) {
	cache, err := cachingstoragewithqueue.NewCacheStorage[*objectsmispformat.ListFormatsMISP](
		cachingstoragewithqueue.WithMaxTtl[*objectsmispformat.ListFormatsMISP](60),
		cachingstoragewithqueue.WithTimeTick[*objectsmispformat.ListFormatsMISP](3),
		cachingstoragewithqueue.WithMaxSize[*objectsmispformat.ListFormatsMISP](10))
	assert.NoError(t, err)

	listId := [...]string{"61317-81923", "12332-32343", "13344-14535", "14451-73422"}

	//кладем в очередь
	for _, id := range listId {
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

	//перекладываем из очереди в кэш
	for {
		obj, isEmpty := cache.PullObjectFromQueue()
		if isEmpty {
			break
		}

		err := cache.AddObjectToCache(obj.GetID(), obj)
		assert.NoError(t, err)

		f, ok := cache.GetFuncFromCacheByKey(obj.GetID())
		assert.True(t, ok)
		f(0)
	}

	t.Run("Тест метода SetIsExecutionTrue который устанавливает статус выполнения задачи в TRUE", func(t *testing.T) {
		cache.SetIsExecutionTrue(listId[1])

		isExecution, ok := cache.GetIsExecution(listId[1])
		assert.True(t, ok)
		assert.True(t, isExecution)

	})

	t.Run("Тест метода который возвращает список индексов объектов, по которым выполяется обработка", func(t *testing.T) {
		assert.Equal(t, len(cache.GetIndexesWithIsExecutionStatus()), 1)
	})

	t.Run("Тест метода SetIsExecutionFalse который устанавливает статус выполнения задачи в FALSE", func(t *testing.T) {
		cache.SetIsExecutionFalse(listId[1])

		isExecution, ok := cache.GetIsExecution(listId[1])
		assert.True(t, ok)
		assert.False(t, isExecution)
	})

	t.Run("Тест метода SetIsCompletedSuccessfullyTrue который устанавливает статус успешности выполнения задачи в TRUE", func(t *testing.T) {
		cache.SetIsCompletedSuccessfullyTrue(listId[2])

		isCompletedSuccessfully, ok := cache.GetIsCompletedSuccessfully(listId[2])
		assert.True(t, ok)
		assert.True(t, isCompletedSuccessfully)
	})

	t.Run("Тест метода SetIsCompletedSuccessfullyFalse который устанавливает статус успешности выполнения задачи в FALSE", func(t *testing.T) {
		cache.SetIsCompletedSuccessfullyFalse(listId[2])

		isCompletedSuccessfully, ok := cache.GetIsCompletedSuccessfully(listId[2])
		assert.True(t, ok)
		assert.False(t, isCompletedSuccessfully)
	})
}
