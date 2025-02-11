package cachingstoragewithqueue_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/cachingstoragewithqueue"
	"github.com/av-belyakov/cachingstoragewithqueue/examples"
	"github.com/av-belyakov/objectsmispformat"
)

func TestSizeObjectToQueue(t *testing.T) {
	t.Run("Тест метода GetSizeObjectToQueue который возвращает размер очереди", func(t *testing.T) {
		cache, err := cachingstoragewithqueue.NewCacheStorage[*objectsmispformat.ListFormatsMISP](
			cachingstoragewithqueue.WithMaxTtl[*objectsmispformat.ListFormatsMISP](60),
			cachingstoragewithqueue.WithTimeTick[*objectsmispformat.ListFormatsMISP](3),
			cachingstoragewithqueue.WithMaxSize[*objectsmispformat.ListFormatsMISP](10))
		assert.NoError(t, err)

		soc := examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
		//для примера используем конструктор списка форматов MISP
		objectTemplate := objectsmispformat.NewListFormatsMISP()
		objectTemplate.ID = "12723-37428"
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

		assert.Equal(t, cache.GetSizeObjectToQueue(), 1)
	})
}
