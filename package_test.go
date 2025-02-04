package cachingstoragewithqueue_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/placeholder_misp/internal/cachestorage"
	"github.com/av-belyakov/placeholder_misp/internal/datamodels"
)

var (
	cache *cachestorage.CacheExecutedObjects[*datamodels.ListFormatsMISP]

	err error
)

type SpecialObjectComparator interface {
	ComparisonID(string) bool
	ComparisonEvent(*datamodels.EventsMispFormat) bool
	ComparisonReports(*datamodels.EventReports) bool
	ComparisonAttributes([]*datamodels.AttributesMispFormat) bool
	ComparisonObjects(map[int]*datamodels.ObjectsMispFormat) bool
	ComparisonObjectTags(*datamodels.ListEventObjectTags) bool
	SpecialObjectGetter
}

type SpecialObjectGetter interface {
	GetID() string
	GetEvent() *datamodels.EventsMispFormat
	GetReports() *datamodels.EventReports
	GetAttributes() []*datamodels.AttributesMispFormat
	GetObjects() map[int]*datamodels.ObjectsMispFormat
	GetObjectTags() *datamodels.ListEventObjectTags
}

type SpecialObjectForCache[T SpecialObjectComparator] struct {
	object      T
	handlerFunc func(int) bool
}

func NewSpecialObjectForCache[T SpecialObjectComparator]() *SpecialObjectForCache[T] {
	return &SpecialObjectForCache[T]{}
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
	cache, err = cachestorage.NewCacheStorage[*datamodels.ListFormatsMISP](
		context.Background(),
		cachestorage.WithMaxTtl[*datamodels.ListFormatsMISP](60),
		cachestorage.WithTimeTick[*datamodels.ListFormatsMISP](3),
		cachestorage.WithMaxSize[*datamodels.ListFormatsMISP](10))

	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(m.Run())
}

func TestQueueHandler(t *testing.T) {
	t.Run("Тест 1. Работа с очередью", func(t *testing.T) {
		cache.PushObjectToQueue(datamodels.NewListFormatsMISP())
		cache.PushObjectToQueue(datamodels.NewListFormatsMISP())
		cache.PushObjectToQueue(datamodels.NewListFormatsMISP())

		assert.Equal(t, cache.SizeObjectToQueue(), 3)

		_, siEmpty := cache.PullObjectFromQueue()
		assert.False(t, siEmpty)
		assert.Equal(t, cache.SizeObjectToQueue(), 2)

		_, _ = cache.PullObjectFromQueue()
		_, _ = cache.PullObjectFromQueue()
		assert.Equal(t, cache.SizeObjectToQueue(), 0)

		_, siEmpty = cache.PullObjectFromQueue()
		assert.True(t, siEmpty)
	})

	t.Run("Тест 1.1. Добавить в очередь некоторое количество объектов", func(t *testing.T) {
		cache.CleanQueue()

		objectTemplate := datamodels.NewListFormatsMISP()
		objectTemplate.ID = "3255-46673"
		cache.PushObjectToQueue(objectTemplate)
		cache.PushObjectToQueue(objectTemplate) //дублирующийся объект

		objectTemplate = datamodels.NewListFormatsMISP()
		objectTemplate.ID = "8483-78578"
		cache.PushObjectToQueue(objectTemplate)

		objectTemplate = datamodels.NewListFormatsMISP()
		objectTemplate.ID = "3132-11223"
		cache.PushObjectToQueue(objectTemplate)

		objectTemplate = datamodels.NewListFormatsMISP()
		objectTemplate.ID = "6553-13323"
		cache.PushObjectToQueue(objectTemplate)

		objectTemplate = datamodels.NewListFormatsMISP()
		objectTemplate.ID = "8474-37722"
		cache.PushObjectToQueue(objectTemplate)

		objectTemplate = datamodels.NewListFormatsMISP()
		objectTemplate.ID = "9123-84885"
		cache.PushObjectToQueue(objectTemplate)

		objectTemplate = datamodels.NewListFormatsMISP()
		objectTemplate.ID = "1200-04993"
		cache.PushObjectToQueue(objectTemplate)

		objectTemplate = datamodels.NewListFormatsMISP()
		objectTemplate.ID = "4323-29909"
		cache.PushObjectToQueue(objectTemplate)

		objectTemplate = datamodels.NewListFormatsMISP()
		objectTemplate.ID = "7605-89493"
		cache.PushObjectToQueue(objectTemplate)

		objectTemplate = datamodels.NewListFormatsMISP()
		objectTemplate.ID = "9423-13373"
		cache.PushObjectToQueue(objectTemplate)

		objectTemplate = datamodels.NewListFormatsMISP()
		objectTemplate.ID = "5238-74389"
		cache.PushObjectToQueue(objectTemplate)

		assert.Equal(t, cache.SizeObjectToQueue(), 12)
	})

	t.Run("Тест 2. Добавить в кэш хранилища некоторое количество объектов находящихся в очереди", func(t *testing.T) {
		//----- первый объект из очереди
		obj, isEmpty := cache.PullObjectFromQueue()
		assert.False(t, isEmpty)
		assert.Equal(t, obj.ID, "3255-46673")

		specialObject := NewSpecialObjectForCache[*datamodels.ListFormatsMISP]()
		specialObject.SetObject(obj)
		specialObject.SetFunc(func(int) bool {
			//здесь некий обработчик...
			//в контексе работы с MISP здесь должен быть код отвечающий
			//за REST запросы к серверу MISP
			fmt.Println("function with ID:", obj.ID)

			return true
		})

		err := cache.AddObjectToCache(specialObject.object.ID, specialObject)
		assert.NoError(t, err)

		_, ok := cache.GetObjectFromCacheByKey(specialObject.object.ID)
		assert.True(t, ok)

		//----- второй объект из очереди
		obj, isEmpty = cache.PullObjectFromQueue()
		assert.False(t, isEmpty)

		specialObject = NewSpecialObjectForCache[*datamodels.ListFormatsMISP]()
		specialObject.SetObject(obj)
		specialObject.SetFunc(func(int) bool {
			//здесь некий обработчик...
			//в контексе работы с MISP здесь должен быть код отвечающий
			//за REST запросы к серверу MISP
			fmt.Println("function with ID:", obj.ID)

			return true
		})

		//должна быть ошибка так как второй в очереди объект имеет такой же
		//идентификатор как и первый
		err = cache.AddObjectToCache(specialObject.object.ID, specialObject)
		assert.Error(t, err)

		_, ok = cache.GetObjectFromCacheByKey(obj.ID)
		assert.True(t, ok)

		cacheSize := cache.SizeObjectToQueue()
		for i := 0; i < cacheSize; i++ {
			obj, isEmpty = cache.PullObjectFromQueue()
			assert.False(t, isEmpty)

			specialObject = NewSpecialObjectForCache[*datamodels.ListFormatsMISP]()
			specialObject.SetObject(obj)
			specialObject.SetFunc(func(int) bool {
				//здесь некий обработчик...
				//в контексе работы с MISP здесь должен быть код отвечающий
				//за REST запросы к серверу MISP
				fmt.Println("function with ID:", obj.ID)

				return true
			})

			err := cache.AddObjectToCache(specialObject.object.ID, specialObject)
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

		cache.AddObjectToCache_TestTimeExpiry("6447-47344", time.Unix(time.Now().Unix()-35, 0), NewSpecialObjectForCache[*datamodels.ListFormatsMISP]())
		time.Sleep(1 * time.Second)

		cache.AddObjectToCache_TestTimeExpiry("3845-21283", time.Unix(time.Now().Unix()-35, 0), NewSpecialObjectForCache[*datamodels.ListFormatsMISP]())
		time.Sleep(1 * time.Second)

		cache.AddObjectToCache_TestTimeExpiry("1734-32222", time.Unix(time.Now().Unix()-35, 0), NewSpecialObjectForCache[*datamodels.ListFormatsMISP]())
		time.Sleep(1 * time.Second)

		indexOldestObject := cache.GetOldestObjectFromCache()
		assert.Equal(t, indexOldestObject, "6447-47344")
		assert.Equal(t, cache.GetCacheSize(), 3)

		cache.DeleteForTimeExpiryObjectFromCache()
		assert.Equal(t, cache.GetCacheSize(), 0)
	})
}
