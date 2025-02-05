package cachingstoragewithqueue_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/cachingstoragewithqueue"
	"github.com/av-belyakov/cachingstoragewithqueue/examples"
	"github.com/av-belyakov/objectsmispformat"
)

//func TestMain(m *testing.M) {
//	os.Exit(m.Run())
//}

func TestQueueHandler(t *testing.T) {
	var (
		cache *cachingstoragewithqueue.CacheStorageWithQueue[*objectsmispformat.ListFormatsMISP]

		err error
	)

	cache, err = cachingstoragewithqueue.NewCacheStorage[*objectsmispformat.ListFormatsMISP](
		context.Background(),
		cachingstoragewithqueue.WithMaxTtl[*objectsmispformat.ListFormatsMISP](60),
		cachingstoragewithqueue.WithTimeTick[*objectsmispformat.ListFormatsMISP](3),
		cachingstoragewithqueue.WithMaxSize[*objectsmispformat.ListFormatsMISP](10))
	if err != nil {
		log.Fatalln(err)
	}

	t.Run("Тест 1. Работа с очередью", func(t *testing.T) {
		//******* добавляем ПЕРВЫЙ тестовый объект ********
		//инициируем вспомогательный тип реализующий интерфейс CacheStorageFuncHandler
		//ult *objectsmispformat.ListFormatsMISP реализует вспомогательный интерфейс
		//SpecialObjectComparator
		//вспомогательный тип и вспомогательный интерфейс задаются пользователем и могут быть любыми
		soc := examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		//для примера используем конструктор списка форматов MISP
		objectTemplate := objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "75473-63475"
		//заполняем вспомогательный тип
		soc.SetID(objectTemplate.ID)
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			//здесь некий обработчик...
			//в контексе работы с MISP здесь должен быть код отвечающий
			//за REST запросы к серверу MISP
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)

		//******* добавляем ВТОРОЙ тестовый объект ********
		soc = examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate = objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "13422-24532"
		soc.SetID(objectTemplate.ID)
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)

		//******* добавляем ТРЕТИЙ тестовый объект ********
		soc = examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate = objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "32145-13042"
		soc.SetID(objectTemplate.ID)
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)

		//проверка количества
		assert.Equal(t, cache.GetSizeObjectToQueue(), 3)

		obj, isEmpty := cache.PullObjectFromQueue()
		assert.False(t, isEmpty)
		assert.Equal(t, obj.GetID(), "75473-63475")
		//проверка возможности вызова функции обработчика
		assert.True(t, obj.GetFunc()(0))
		assert.Equal(t, cache.GetSizeObjectToQueue(), 2)

		_, _ = cache.PullObjectFromQueue()
		_, _ = cache.PullObjectFromQueue()
		assert.Equal(t, cache.GetSizeObjectToQueue(), 0)

		_, isEmpty = cache.PullObjectFromQueue()
		assert.True(t, isEmpty)
	})

	t.Run("Тест 1.1. Добавить в очередь некоторое количество объектов", func(t *testing.T) {
		cache.CleanQueue()

		//1.
		soc := examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate := objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "3255-46673"
		soc.SetID(objectTemplate.ID)
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)
		cache.PushObjectToQueue(soc) //дублирующийся объект

		//2.
		soc = examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate = objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "8483-78578"
		soc.SetID(objectTemplate.ID)
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)

		//3.
		soc = examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate = objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "3132-11223"
		soc.SetID(objectTemplate.ID)
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)

		//4.
		soc = examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate = objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "6553-13323"
		soc.SetID(objectTemplate.ID)
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)

		//5.
		soc = examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate = objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "8474-37722"
		soc.SetID(objectTemplate.ID)
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)
		objectTemplate = objectsmispformat.NewListFormatsMISP()

		//6.
		soc = examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate = objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "9123-84885"
		soc.SetID(objectTemplate.ID)
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)

		//7.
		soc = examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate = objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "1200-04993"
		soc.SetID(objectTemplate.ID)
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)

		//8.
		soc = examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate = objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "4323-29909"
		soc.SetID(objectTemplate.ID)
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)

		//9.
		soc = examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate = objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "7605-89493"
		soc.SetID(objectTemplate.ID)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)

		//10.
		soc = examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate = objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "9423-13373"
		soc.SetID(objectTemplate.ID)
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)

		//11.
		soc = examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate = objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "5238-74389"
		soc.SetID(objectTemplate.ID)
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)

		assert.Equal(t, cache.GetSizeObjectToQueue(), 12)
	})

	t.Run("Тест 2. Добавить в кэш хранилища некоторое количество объектов находящихся в очереди", func(t *testing.T) {
		//----- первый объект из очереди
		obj, isEmpty := cache.PullObjectFromQueue()
		assert.False(t, isEmpty)
		assert.Equal(t, obj.GetID(), "3255-46673")

		err := cache.AddObjectToCache(obj.GetID(), obj)
		assert.NoError(t, err)

		_, ok := cache.GetObjectFromCacheByKey(obj.GetID())
		assert.True(t, ok)

		//----- второй объект из очереди
		obj, isEmpty = cache.PullObjectFromQueue()
		assert.False(t, isEmpty)

		//должна быть ошибка так как второй в очереди объект имеет такой же
		//идентификатор как и первый
		err = cache.AddObjectToCache(obj.GetID(), obj)
		assert.Error(t, err)

		_, ok = cache.GetObjectFromCacheByKey(obj.GetID())
		assert.True(t, ok)

		cacheSize := cache.GetSizeObjectToQueue()
		for i := 0; i < cacheSize; i++ {
			obj, isEmpty = cache.PullObjectFromQueue()
			assert.False(t, isEmpty)

			err := cache.AddObjectToCache(obj.GetID(), obj)
			assert.NoError(t, err)
		}

		assert.Equal(t, cache.GetCacheSize(), 11)
	})

	t.Run("Тест 3. Найти объект с самой старой записью в кэше", func(t *testing.T) {
		index := cache.GetOldestObjectFromCache()
		assert.Equal(t, index, "3255-46673")

		obj, ok := cache.GetObjectFromCacheByKey(index)
		assert.True(t, ok)
		assert.Equal(t, obj.GetID(), "3255-46673")
	})

	t.Run("Тест 4. Проверить поиск в кэше исполняемой функции которая в настоящее время не выполняется, не была успешно выполнена и время истечения жизни объекта которой самое меньшее", func(t *testing.T) {
		index, _ := cache.GetFuncFromCacheMinTimeExpiry()
		assert.Equal(t, index, "3255-46673")
		obj, ok := cache.GetObjectFromCacheByKey(index)
		assert.True(t, ok)
		assert.Equal(t, obj.GetID(), "3255-46673")
	})

	t.Run("Тест 5. Проверить удаляются ли объекты время жизни которых, time expiry, истекло", func(t *testing.T) {
		//
		//Надо дописать этот тест
		//

		//очищаем данные из кеша
		cache.CleanCache()

		cache.AddObjectToCache_TestTimeExpiry("6447-47344", time.Unix(time.Now().Unix()-35, 0), examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]())
		time.Sleep(1 * time.Second)

		cache.AddObjectToCache_TestTimeExpiry("3845-21283", time.Unix(time.Now().Unix()-35, 0), examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]())
		time.Sleep(1 * time.Second)

		cache.AddObjectToCache_TestTimeExpiry("1734-32222", time.Unix(time.Now().Unix()-35, 0), examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]())
		time.Sleep(1 * time.Second)

		indexOldestObject := cache.GetOldestObjectFromCache()
		assert.Equal(t, indexOldestObject, "6447-47344")
		assert.Equal(t, cache.GetCacheSize(), 3)

		cache.DeleteForTimeExpiryObjectFromCache()
		assert.Equal(t, cache.GetCacheSize(), 0)
	})
}
