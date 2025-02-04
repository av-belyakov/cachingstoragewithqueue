package cachingstoragewithqueue_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/cachingstoragewithqueue"
	"github.com/av-belyakov/objectsmispformat"
)

var (
	cache *cachingstoragewithqueue.CacheStorageWithQueue[*objectsmispformat.ListFormatsMISP]

	err error
)

type SpecialObjectComparator interface {
	ComparisonID(string) bool
	ComparisonEvent(*objectsmispformat.EventsMispFormat) bool
	ComparisonReports(*objectsmispformat.EventReports) bool
	ComparisonAttributes([]*objectsmispformat.AttributesMispFormat) bool
	ComparisonObjects(map[int]*objectsmispformat.ObjectsMispFormat) bool
	ComparisonObjectTags(*objectsmispformat.ListEventObjectTags) bool
	SpecialObjectGetter
}

type SpecialObjectGetter interface {
	GetID() string
	GetEvent() *objectsmispformat.EventsMispFormat
	GetReports() *objectsmispformat.EventReports
	GetAttributes() []*objectsmispformat.AttributesMispFormat
	GetObjects() map[int]*objectsmispformat.ObjectsMispFormat
	GetObjectTags() *objectsmispformat.ListEventObjectTags
}

// SpecialObjectForCache является вспомогательным типом который реализует интерфейс
// CacheStorageFuncHandler[T any] где в методе Comparison(objFromCache T) bool необходимо
// реализовать подробное сравнение объекта типа T
type SpecialObjectForCache[T SpecialObjectComparator] struct {
	object      T
	handlerFunc func(int) bool
	id          string
}

// NewSpecialObjectForCache конструктор вспомогательного типа реализующий интерфейс CacheStorageFuncHandler[T any]
func NewSpecialObjectForCache[T SpecialObjectComparator]() *SpecialObjectForCache[T] {
	return &SpecialObjectForCache[T]{}
}

func (o *SpecialObjectForCache[T]) SetID(v string) {
	o.id = v
}

func (o *SpecialObjectForCache[T]) GetID() string {
	return o.id
}

func (o *SpecialObjectForCache[T]) SetObject(v T) {
	o.object = v
}

func (o *SpecialObjectForCache[T]) GetObject() T {
	return o.object
}

func (o *SpecialObjectForCache[T]) SetFunc(f func(int) bool) {
	o.handlerFunc = f
}

func (o *SpecialObjectForCache[T]) GetFunc() func(int) bool {
	return o.handlerFunc
}

func (o *SpecialObjectForCache[T]) Comparison(objFromCache T) bool {
	if !o.object.ComparisonID(objFromCache.GetID()) {
		return false
	}

	if !o.object.ComparisonEvent(objFromCache.GetEvent()) {
		return false
	}

	if !o.object.ComparisonReports(objFromCache.GetReports()) {
		return false
	}

	if !o.object.ComparisonAttributes(objFromCache.GetAttributes()) {
		return false

	}

	if !o.object.ComparisonObjects(objFromCache.GetObjects()) {
		return false
	}

	if !o.object.ComparisonObjectTags(o.object.GetObjectTags()) {
		return false
	}

	return true
}

func TestMain(m *testing.M) {
	cache, err = cachingstoragewithqueue.NewCacheStorage[*objectsmispformat.ListFormatsMISP](
		context.Background(),
		cachingstoragewithqueue.WithMaxTtl[*objectsmispformat.ListFormatsMISP](60),
		cachingstoragewithqueue.WithTimeTick[*objectsmispformat.ListFormatsMISP](3),
		cachingstoragewithqueue.WithMaxSize[*objectsmispformat.ListFormatsMISP](10))
	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(m.Run())
}

func TestQueueHandler(t *testing.T) {
	t.Run("Тест 1. Работа с очередью", func(t *testing.T) {
		//******* добавляем ПЕРВЫЙ тестовый объект ********
		//инициируем вспомогательный тип реализующий интерфейс CacheStorageFuncHandler
		//ult *objectsmispformat.ListFormatsMISP реализует вспомогательный интерфейс
		//SpecialObjectComparator
		//вспомогательный тип и вспомогательный интерфейс задаются пользователем и могут быть любыми
		soc := NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
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
		soc = NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
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
		soc = NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
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
		assert.Equal(t, cache.SizeObjectToQueue(), 3)

		obj, isEmpty := cache.PullObjectFromQueue()
		assert.False(t, isEmpty)
		assert.Equal(t, obj.GetID(), "75473-63475")
		//проверка возможности вызова функции обработчика
		assert.True(t, obj.GetFunc()(0))
		assert.Equal(t, cache.SizeObjectToQueue(), 2)

		_, _ = cache.PullObjectFromQueue()
		_, _ = cache.PullObjectFromQueue()
		assert.Equal(t, cache.SizeObjectToQueue(), 0)

		_, isEmpty = cache.PullObjectFromQueue()
		assert.True(t, isEmpty)
	})

	t.Run("Тест 1.1. Добавить в очередь некоторое количество объектов", func(t *testing.T) {
		cache.CleanQueue()

		//1.
		soc := NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
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
		soc = NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
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
		soc = NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
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
		soc = NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
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
		soc = NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
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
		soc = NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
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
		soc = NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
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
		soc = NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
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
		soc = NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate = objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "7605-89493"
		soc.SetID(objectTemplate.ID)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)

		//10.
		soc = NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
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
		soc = NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		objectTemplate = objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "5238-74389"
		soc.SetID(objectTemplate.ID)
		soc.SetObject(objectTemplate)
		soc.SetFunc(func(int) bool {
			fmt.Println("function with ID:", soc.GetID())

			return true
		})
		cache.PushObjectToQueue(soc)

		assert.Equal(t, cache.SizeObjectToQueue(), 12)
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

		cacheSize := cache.SizeObjectToQueue()
		for i := 0; i < cacheSize; i++ {
			obj, isEmpty = cache.PullObjectFromQueue()
			assert.False(t, isEmpty)

			err := cache.AddObjectToCache(obj.GetID(), obj)
			assert.NoError(t, err)
		}

		assert.Equal(t, cache.GetCacheSize(), 10)
	})

	t.Run("Тест 3. Найти объект с самой старой записью в кэше", func(t *testing.T) {
		index := cache.GetOldestObjectFromCache()
		assert.Equal(t, index, "8483-78578")

		obj, ok := cache.GetObjectFromCacheByKey(index)
		assert.True(t, ok)
		assert.Equal(t, obj.GetID(), "8483-78578")
	})

	t.Run("Тест 4. Проверить удаляются ли объекты время жизни которых, time expiry, истекло", func(t *testing.T) {
		//
		//Надо дописать этот тест
		//

		//очищаем данные из кеша
		cache.CleanCache_Test()

		cache.AddObjectToCache_TestTimeExpiry("6447-47344", time.Unix(time.Now().Unix()-35, 0), NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]())
		time.Sleep(1 * time.Second)

		cache.AddObjectToCache_TestTimeExpiry("3845-21283", time.Unix(time.Now().Unix()-35, 0), NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]())
		time.Sleep(1 * time.Second)

		cache.AddObjectToCache_TestTimeExpiry("1734-32222", time.Unix(time.Now().Unix()-35, 0), NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]())
		time.Sleep(1 * time.Second)

		indexOldestObject := cache.GetOldestObjectFromCache()
		assert.Equal(t, indexOldestObject, "6447-47344")
		assert.Equal(t, cache.GetCacheSize(), 3)

		cache.DeleteForTimeExpiryObjectFromCache()
		assert.Equal(t, cache.GetCacheSize(), 0)
	})
}
