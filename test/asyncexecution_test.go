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

func TestAsynExecution(t *testing.T) {
	listExample := []string{
		"aa83245",
		"bb43522",
		"cc19345",
		"dd75621",
		"ee72452",
		"ff31442",
		"hh90134",
		"ii61342",
		"gg32352",
		"kk81345",
		"ll67341",
		"mm83142",
		"nn14421",
		"oo46231",
		"pp51239",
	}

	cache, err := cachingstoragewithqueue.NewCacheStorage[*objectsmispformat.ListFormatsMISP](
		context.Background(),
		cachingstoragewithqueue.WithMaxTtl[*objectsmispformat.ListFormatsMISP](300),
		cachingstoragewithqueue.WithTimeTick[*objectsmispformat.ListFormatsMISP](3),
		cachingstoragewithqueue.WithMaxSize[*objectsmispformat.ListFormatsMISP](10),
		cachingstoragewithqueue.WithEnableAsyncProcessing[*objectsmispformat.ListFormatsMISP](4))
	assert.NoError(t, err)

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

	t.Run("Тест 1. Проверяем асинхронную обработку по группам", func(t *testing.T) {
		addObjectToQueue(listExample)
		assert.Equal(t, cache.GetSizeObjectToQueue(), len(listExample))

		chStop := make(chan cachingstoragewithqueue.HandlerOptionsStoper)
		chDone := make(chan struct{})

		go func() {
			var count int
			for data := range chStop {
				count++

				fmt.Printf("received from chan, index:'%s', isSuccess:'%t'\n", data.GetIndex(), data.GetIsSuccess())

				cache.ChangeValues(data.GetIndex(), data.GetIsSuccess())
				if count == len(listExample) {
					chDone <- struct{}{}

					return
				}
			}
		}()

		for {
			fmt.Println("___ cache.GetSizeObjectToQueue(): ", cache.GetSizeObjectToQueue())

			if cache.GetSizeObjectToQueue() == 0 {
				break
			}

			go cache.AsyncExecution_Test(chStop)
		}

		<-chDone

		t.Log("cache.GetCacheSize():", cache.GetCacheSize())

		assert.Equal(t, cache.GetSizeObjectToQueue(), 0)
		assert.Equal(t, cache.GetCacheSize(), len(listExample))
	})
}
