package main

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"strconv"
	"time"
)

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
		log.Println(data)
		log.Println("price data unmarshal error")
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
			log.Panicln(reflect.TypeOf(rawData[6]))
			log.Println(rawData[6])
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