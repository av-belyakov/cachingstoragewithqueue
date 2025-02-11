package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/av-belyakov/cachingstoragewithqueue"
	"github.com/av-belyakov/cachingstoragewithqueue/examples"
	"github.com/av-belyakov/objectsmispformat"
	"github.com/google/uuid"
)

var (
	cache *cachingstoragewithqueue.CacheStorageWithQueue[*objectsmispformat.ListFormatsMISP]
	/*listExample []string = []string{
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
	}*/

	err error
)

func main() {
	// добавление в очередь новых объектов
	/*addObjectToQueue := func(lid []string) {
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
	}*/

	addObjectToQueue := func(ctx context.Context) {
		tick := time.NewTicker(4 * time.Second)

		go func() {
			<-ctx.Done()
			tick.Stop()
		}()

		for range tick.C {
			soc := examples.NewSpecialObjectForCache[*objectsmispformat.ListFormatsMISP]()
			//для примера используем конструктор списка форматов MISP
			objectTemplate := objectsmispformat.NewListFormatsMISP()
			objectTemplate.ID = uuid.NewString()

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

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		chSignal := make(chan os.Signal, 1)
		log.Printf("system call:%+v", <-chSignal)

		cancel()
	}()

	cache, err = cachingstoragewithqueue.NewCacheStorage[*objectsmispformat.ListFormatsMISP](
		cachingstoragewithqueue.WithMaxTtl[*objectsmispformat.ListFormatsMISP](300),
		cachingstoragewithqueue.WithTimeTick[*objectsmispformat.ListFormatsMISP](3),
		cachingstoragewithqueue.WithMaxSize[*objectsmispformat.ListFormatsMISP](10),
		cachingstoragewithqueue.WithEnableAsyncProcessing[*objectsmispformat.ListFormatsMISP](4))
	if err != nil {
		log.Fatal(err)
	}

	//addObjectToQueue(listExample)
	go addObjectToQueue(ctx)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	log.Println("Package 'cachestoragewithqueue' is start")

	cache.StartAutomaticExecution(ctx)
}
