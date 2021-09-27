package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

type RecentTrade struct {
    Id string // ID транзакции
    Pair string // Торговая пара (из списка выше)
    Price float64 // Цена транзакции
    Amount float64 // Объём транзакции
    Side string // Как биржа засчитала эту сделку (как buy или как sell)
    Timestamp time.Time // Время транзакции
}

const poloniexWSUrl string = "wss://api2.poloniex.com"
var poloniexParListJson = []byte(`{"poloniex":["BTC_USDT", "TRX_USDT", "ETH_USDT"]}`)

func main() {
	pairList, err := getPolonexPairList(poloniexParListJson)
	if err != nil {
		log.Fatalln(err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	for _, pair := range pairList {
		pConn, err := newPoloniexConnectWs(ctx, poloniexWSUrl)
		if err != nil {
			log.Fatalln(err)
		}
		defer func(pConn *poloniex) {
			if err := pConn.close(); err != nil {
				log.Println(err)
			}
		}(pConn)

		rawDataCh := make(chan []byte)
		parseResultCh := make(chan RecentTrade)
		
		wg.Add(1)
		go func(rawDataCh chan []byte, parseResultCh chan RecentTrade) {
			defer wg.Done()
			defer close(parseResultCh)
			parseResponse(rawDataCh, parseResultCh, pair)
		}(rawDataCh, parseResultCh)

		wg.Add(1)
		go func() {
			defer wg.Done()
			for trade := range parseResultCh {
				processTrade(trade)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(rawDataCh)
			pConn.doRead(ctx, rawDataCh)
		}()
	
		cmd, err := getSubscribeCommand(pair)
		if err != nil {
			log.Printf("can not get command for pair: %s, error: %v", pair, err)
			continue
		}

		err = pConn.sendCommand(cmd)
		if err != nil {
			log.Printf("can not subscribe with command: %s, error: %v", cmd, err)
			continue
		}
	}

	go func() {
		interruptCh := make(chan os.Signal, 1)
		signal.Notify(interruptCh, os.Interrupt)
		<-interruptCh
		log.Printf("graceful shutdown")
		cancel()
	}()
	
	wg.Wait()
}

func processTrade(trade RecentTrade) {
	log.Println(trade)
}
