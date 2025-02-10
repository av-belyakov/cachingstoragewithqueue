package cachingstoragewithqueue_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/cachingstoragewithqueue"
	"github.com/av-belyakov/cachingstoragewithqueue/examples"
	"github.com/av-belyakov/objectsmispformat"
)

type stopOptions struct {
	index     string
	isSuccess bool
}

// GetIndex получить индекс
func (sho *stopOptions) GetIndex() string {
	return sho.index
}

// SetIndex установить индекс
func (sho *stopOptions) SetIndex(v string) {
	sho.index = v
}

// GetIsSuccess получить статус
func (sho *stopOptions) GetIsSuccess() bool {
	return sho.isSuccess
}

// SetIsSuccess установить статус
func (sho *stopOptions) SetIsSuccess(v bool) {
	sho.isSuccess = v
}

func TestSyncExecution(t *testing.T) {
	cache, err := cachingstoragewithqueue.NewCacheStorage[*objectsmispformat.ListFormatsMISP](
		context.Background(),
		cachingstoragewithqueue.WithMaxTtl[*objectsmispformat.ListFormatsMISP](300),
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

	listIdTwo := []string{
		"aaaa-1111",
		"bbbb-2222",
		"cccc-3333",
		"dddd-4444",
		"eeee-5555",
		//"ffff-6666",
		//"gggg-7777",
		//"hhhh-8888",
		//"iiii-9999",
		//"ssss-0000",
	}

	//добавление в очередь новых объектов
	addObjectToQueue := func(lid []string) {
		for _, id := range lid {
			soc := examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
			//для примера используем конструктор списка форматов MISP
			objectTemplate := objectsmispformat.NewListFormatsMISP()
			objectTemplate.ID = id

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

	t.Run("Тест 1. Все объекты должны быть помечены как выполненные", func(t *testing.T) {
		chStop := make(chan cachingstoragewithqueue.HandlerOptionsStoper)
		ctx, ctxClose := context.WithCancel(context.Background())

		go func() {
			for {
				select {
				case <-ctx.Done():
					return

				case data := <-chStop:
					fmt.Printf("received from chan, index:'%s', isSuccess:'%t'\n", data.GetIndex(), data.GetIsSuccess())

					cache.ChangeValues(data.GetIndex(), data.GetIsSuccess())
				}
			}
		}()

		//кладем в очередь объекты которые необходимо обработать
		addObjectToQueue(listIdOne)
		for {
			if cache.GetSizeObjectToQueue() == 0 {
				break
			}

			//кладем в кэш все объекты из очереди
			cache.SyncExecution_Test(chStop)
		}

		time.Sleep(1 * time.Second)

		ctxClose()

		t.Log("cache.GetCacheSize():", cache.GetCacheSize())
		assert.Equal(t, len(cache.GetIndexesWithIsCompletedSuccessfully()), len(listIdOne))

		//очищаем кэш
		cache.CleanCache()
	})

	t.Run("Тест 2. Добавление новых объектов из очереди не производится если в кэше есть хотя бы одно активное задание", func(t *testing.T) {
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

		// кладем в ОЧЕРЕДЬ, не в КЭШ, объекты которые необходимо обработать
		addObjectToQueue(listIdOne)
		//размер очереди должен быть 10 объектов
		assert.Equal(t, cache.GetSizeObjectToQueue(), 10)

		//меняем статус состояния объекта isExecute = true
		cache.SetIsExecutionTrue(obj.GetID())

		chStop := make(chan cachingstoragewithqueue.HandlerOptionsStoper)
		ctx, ctxClose := context.WithCancel(context.Background())

		go func() {
			for {
				select {
				case <-ctx.Done():
					return

				case data := <-chStop:
					cache.ChangeValues(data.GetIndex(), data.GetIsSuccess())
				}
			}
		}()

		var num int
		for {
			num++

			if num == 4 {
				//пока не поменяли статус выполнения (isExecute) единственного объекта
				//на true количество объектов в кэше не должно увеличиватся
				assert.Equal(t, cache.GetCacheSize(), 1)

				//меняем статус состояния объекта isExecute = false что бы разрешить
				//добавление новых объектов
				cache.SetIsExecutionFalse(obj.GetID())
				status, ok := cache.GetIsExecution(obj.GetID())
				assert.True(t, ok)
				t.Logf("_______ change status 'isExecute' for object with id '%s' to '%t' ______", obj.GetID(), status)
			}

			//поиск и удаление самого старого объекта если размер кэша достиг максимального значения
			//выполняется удаление объекта который в настоящее время не выполняеться и ранее был успешно выполнен
			t.Logf("======= cache.GetCacheSize(): '%d' == '%d' cache.GetCacheMaxSize_Test()", cache.GetCacheSize(), cache.GetCacheMaxSize_Test())
			if cache.GetCacheSize() == cache.GetCacheMaxSize_Test() {
				t.Log("cache.GetSizeObjectToQueue() =", cache.GetSizeObjectToQueue())

				if err := cache.DeleteOldestObjectFromCache(); err != nil {
					t.Log("Delete error:", err)
				}

				continue
			}

			//пытаемся выполнить синхронную обработку, там же и добавляем новые объекты
			//из кэша, однако первые 4 прохода объекты не добавляются, так как в кэше
			//есть 1 объект мо статусом isExecute = true
			cache.SyncExecution_Test(chStop)

			//если из кэша не удалять самый старый объект, то тогда из очереди в кэш будут
			// добавлены только 9 элементов, соответственно выход из цикла когда в очереди
			// остался только 1 элемент, если раскомментировать удаление, колторое ниже,
			// тогда сравнение должно быть с 0
			if cache.GetSizeObjectToQueue() == 0 {
				break
			}
		}

		ctxClose()

		//проверяем количество добавленых в КЭШ объектов
		//добавлен список с 10 элементами так как элемент с id "1234-5678"
		//был удалён ранее
		assert.Equal(t, cache.GetCacheSize(), 10)

		// есть самый старый элемент (0000-0000)
		assert.Equal(t, cache.GetOldestObjectFromCache(), listIdOne[0])

		//есть самый последний элемент (9999-9999)
		_, isExist := cache.GetObjectFromCacheByKey(listIdOne[len(listIdOne)-1])
		assert.True(t, isExist)
	})

	t.Run("Тест 3. Добавление объектов производится ни один объект в кэше не участвует в обработке", func(t *testing.T) {
		chStop := make(chan cachingstoragewithqueue.HandlerOptionsStoper)
		ctx, ctxClose := context.WithCancel(context.Background())

		go func() {
			for {
				select {
				case <-ctx.Done():
					return

				case data := <-chStop:
					cache.ChangeValues(data.GetIndex(), data.GetIsSuccess())
				}
			}
		}()

		// кладем в очередь объекты которые необходимо обработать
		addObjectToQueue(listIdTwo)
		for {
			//поиск и удаление самого старого объекта если размер кэша достиг максимального значения
			//выполняется удаление объекта который в настоящее время не выполняеться и ранее был успешно выполнен
			if cache.GetCacheSize() >= cache.GetCacheMaxSize_Test() {
				if err := cache.DeleteOldestObjectFromCache(); err != nil {
					t.Log("Delete error:", err)
				}

				continue
			}

			cache.SyncExecution_Test(chStop)

			if cache.GetSizeObjectToQueue() == 0 {
				break
			}
		}

		ctxClose()

		t.Log("--------", cache.GetCacheSize())
		//так как размер кеша 10 объектов, в кэше уже есть 1 элемент, а список добавляемых
		//объектов состоит из 10, то при добавлении самый старый объект (первый из добавленых)
		//должен быть удалён
		assert.Equal(t, cache.GetCacheSize(), 10)
	})
}
