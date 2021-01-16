package ibgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Instrument struct {
	contract Contract
	detail   ContractDetail
	client   *IBClient
}

func (ins *Instrument) Contract() Contract {
	return ins.contract
}

func (ins *Instrument) ConID() int64 {
	return ins.contract.ConID
}

func (c *IBClient) NewInstrument(contract Contract) (*Instrument, error) {
	if contract, err := ensureOneContract(c, contract); err != nil {
		return nil, err
	} else {
		return &Instrument{contract: contract.Contract, detail: contract.ContractDetail, client: c}, nil
	}
}

func (c *IBClient) NewInstrumentFromConid(id int) (*Instrument, error) {
	return c.NewInstrument(Contract{ConID: int64(id)})
}

func (c *IBClient) USStock(symbol string) (*Instrument, error) {
	if contract, err := ensureOneContract(c, Contract{Symbol: symbol, SecType: "STK", Currency: "USD", Exchange: "SMART"}); err != nil {
		return nil, err
	} else {
		return &Instrument{contract: contract.Contract, detail: contract.ContractDetail, client: c}, nil
	}
}

func (c *IBClient) USFuture(symbol string, expiration string) (*Instrument, error) {
	t := "FUT"
	e := expiration
	if expiration == "CONTFUT" || expiration == "" {
		t = "CONTFUT"
	}
	if expiration == "CONTFUT" {
		e = ""
	}
	if to, ok := CommonFutureSymbolMap[symbol]; ok {
		symbol = to
	}
	respCh, _, _ := c.REQ(&ReqContractDetail{Contract{Symbol: symbol, SecType: t, Currency: "USD", LastTradeDateOrContractMonth: e}})
	cs, err := HandleContractDetail(respCh)
	if err != nil {
		return nil, err
	}
	if len(cs) == 0 {
		return nil, fmt.Errorf("ContractData request end unexpected or cancelled by user")
	}
	var cdata ContractData
	if len(cs) == 1 {
		cdata = cs[0]
	} else {
		candidates := make([]*ContractData, 0)
		for i := range cs {
			c := cs[i]
			fmt.Println(candidates)
			if _, ok := KnowUSFutureExchange[c.Exchange]; ok {
				candidates = append(candidates, &c)
			}
		}
		if len(candidates) == 1 {
			cdata = *candidates[0]
			// fmt.Println(cdata.Contract)
		} else {
			return nil, fmt.Errorf("there is ambiguity in the contract defination. Check out the log for detail. Use IBClient.REQ method to find out ConID of the desired contract")
		}
	}
	if cdata.SecType == "CONTFUT" && expiration == "" {
		cdata.SecType = "FUT"
	}
	return &Instrument{contract: cdata.Contract, detail: cdata.ContractDetail, client: c}, nil
}

func (ins *Instrument) TickStream(typ string) (chan Tick, error) {
	if typ != "Last" && typ != "AllLast" && typ != "BidAsk" && typ != "MidPoint" {
		return nil, errors.New("value erro")
	}
	msgCh, _, err := ins.client.REQ(ReqTickByTickData{&ins.contract, typ, 0, false})
	if err != nil {
		return nil, err
	}
	for m := range msgCh {
		if m.Code == inTICKBYTICK {
		}
	}
	return nil, nil
}
func HandleContractDetail(ch chan *Message) ([]ContractData, error) {
	cs := make([]ContractData, 0)
	for m := range ch {
		if m.Error != nil {
			return nil, m.Error
		}
		if m.Code == inCONTRACTDATAEND {
			return cs, nil
		}
		c := ContractData{}
		c.Read(m)
		cs = append(cs, c)
	}
	return nil, fmt.Errorf("ContractData request end unexpected or cancelled by user")
}

func ensureOneContract(c *IBClient, ct Contract) (*ContractData, error) {
	respCh, _, _ := c.REQ(ReqContractDetail{ct})
	if cs, err := HandleContractDetail(respCh); err != nil {
		return nil, err
	} else if len(cs) == 0 {
		return nil, fmt.Errorf("ContractData request end unexpected or cancelled by user")
	} else if len(cs) != 1 {
		var id int64
		for _, c := range cs {
			if id != 0 && c.ConID != id {
				return nil, fmt.Errorf("there is ambiguity in the contract defination. Check out the log for detail. Use IBClient.REQ method to find out ConID of the desired contract")
			}
			id = c.ConID
		}
		return &cs[0], nil
	} else {
		return &cs[0], nil
	}
}

func HanddleTickbyTick(ch chan *Message) {
	var last time.Time
	var i int
	for m := range ch {
		// Printmsg(m)
		if m.Code != inTICKBYTICK {
			Printmsg(m)
			break
		}
		t := Tick{}
		t.Read(m)
		if t.Time.After(last) {
			// fmt.Println("after")
			i = 1
			last = t.Time
		} else if t.Time.Before(last) {
			// fmt.Println("before")
			// fmt.Println("ERROR!")
		} else {
			// fmt.Println("equal", i)
			t.SetMilli(i)
			i++
		}

		t.SetRTH()
		b, err := json.Marshal(&t)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	}
}

var KnowUSFutureExchange = map[string]struct{}{
	"GLOBEX": {},
	"CME":    {},
	"CBOE":   {},
	"CBOT":   {},
	"COME":   {},
	"ICE":    {},
	"NYMEX":  {},
}
var CommonFutureSymbolMap = map[string]string{
	"6B":  "GBP",
	"6C":  "CAD",
	"6J":  "JPY",
	"6S":  "CHF",
	"6E":  "EUR",
	"6A":  "AUD",
	"VX":  "VIX",
	"BTC": "BRR",
}
