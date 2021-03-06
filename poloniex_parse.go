package main

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
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

type mainResp struct {
	data []priceData
}

type priceData struct {
	dataType string
	id string
	sideCode float64
	price float64
	amount float64
	timestamp time.Time
}

const tradeTypeLetter = "t"

func (r *mainResp) UnmarshalJSON(data []byte) error {
	var channelID int
	var seqNum int
	firstStageData := []interface{}{&channelID, &seqNum, &r.data}

	var err error
	if err = json.Unmarshal(data, &firstStageData); err != nil {
	   return err
	}

	return nil
}

func (pd *priceData) UnmarshalJSON(data []byte) error {
	var rawData []interface{}
	var err error
	if err = json.Unmarshal(data, &rawData); err != nil {
		return err
	}
	// ["t", "<trade id>", <1 for buy 0 for sell>, "<price>", "<size>", <timestamp>, "<epoch_ms>"]
	var ok bool
	pd.dataType, ok = rawData[0].(string)
	if !ok {
		return errors.New("can not parse data type")
	}

	if pd.dataType == tradeTypeLetter {
		pd.id, ok = rawData[1].(string)
		if !ok {
			return errors.New("can not parse trade id")
		}

		pd.sideCode, ok = rawData[2].(float64)
		if !ok {
			return errors.New("can not parse trade id")
		}

		priceStr, ok := rawData[3].(string)
		if !ok {
			return errors.New("can not parse price")
		}
		pd.price, err = strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return err
		}

		amountStr, ok := rawData[4].(string)
		if !ok {
			return errors.New("can not parse amount")
		}
		pd.amount, err = strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return err
		}

		unixTimeMsStr, ok := rawData[6].(string)
		if !ok {
			return errors.New("can not parse time")
		}
		unixTimeMs, err := strconv.ParseInt(unixTimeMsStr, 10, 64)
		if err != nil {
			return err
		}
		pd.timestamp = time.UnixMilli(unixTimeMs)
	}

	return nil
}

func parseResponse(inCh <-chan []byte, outCh chan<- RecentTrade, poloniexPair string) {
	resp := new(mainResp)
	for r := range inCh {
		err := json.Unmarshal(r, &resp)
		if err != nil {
			log.Printf("unmarshal error: %s\n", err.Error())
			continue
		}

		for _, row := range resp.data {
			if row.dataType != tradeTypeLetter {
				continue
			}
			pair, err := getRevertedPair(poloniexPair)
			if err != nil {
				log.Println(err.Error())
			}

			outCh <- RecentTrade{
				Id: row.id,
				Pair: pair,
				Price: row.price,
				Amount: row.amount,
				Side: getSideText(row.sideCode),
				Timestamp: row.timestamp,
			}
		}	
	}
}