package cachingstoragewithqueue_test

import (
	"fmt"
	"testing"

	"github.com/av-belyakov/cachingstoragewithqueue"
	"github.com/av-belyakov/cachingstoragewithqueue/examples"
	"github.com/av-belyakov/objectsmispformat"
	"github.com/stretchr/testify/assert"
)

func TestPullMaxObjectFromQueue(t *testing.T) {
	var (
		cache *cachingstoragewithqueue.CacheStorageWithQueue[*objectsmispformat.ListFormatsMISP]

		err error
	)

	t.Run("Тест 1. Инициализация хранилища", func(t *testing.T) {
		cache, err = cachingstoragewithqueue.NewCacheStorage[*objectsmispformat.ListFormatsMISP](
			cachingstoragewithqueue.WithMaxTtl[*objectsmispformat.ListFormatsMISP](300),
			cachingstoragewithqueue.WithTimeTick[*objectsmispformat.ListFormatsMISP](3),
			cachingstoragewithqueue.WithMaxSize[*objectsmispformat.ListFormatsMISP](10),
			cachingstoragewithqueue.WithEnableAsyncProcessing[*objectsmispformat.ListFormatsMISP](4))

		assert.NoError(t, err)
	})

	t.Run("Тест 2. Добавление в очередь новых объектов", func(t *testing.T) {
		listId := [...]string{
			"a643-a483",
			"b214-b343",
			"c374-c935",
			"d851-d732",
			"e735-e421",
			"f875-f843",
			"g735-g193",
			"h035-h021",
			"i672-i123",
			"j912-j102",
			"k731-k022",
		}

		//кладем в очередь
		for _, id := range listId {
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

		assert.Equal(t, cache.GetSizeObjectToQueue(), len(listId))
	})

	t.Run("Тест 3. Получить из очереди список объектов, разово, не превышающий параметр isAsync", func(t *testing.T) {
		for i := 0; i < 4; i++ {
			objects, isEmpty := cache.PullMaxObjectFromQueue()

			if i == 0 || i == 1 {
				assert.Equal(t, len(objects), cache.GetIsAsync_Test())
				assert.False(t, isEmpty)
			} else if i == 2 {
				assert.Equal(t, len(objects), 3)
			} else {
				assert.True(t, isEmpty)
			}
		}
	})
}
