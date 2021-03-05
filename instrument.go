package ibgo

import (
	"fmt"
	"time"
)

type Instrument struct {
	contract          Contract
	Detail            ContractDetail
	client            *IBClient
	lastHistoricalReq time.Time
}

func (ins *Instrument) Contract() Contract {
	return ins.contract
}

func (ins *Instrument) ConID() int64 {
	return ins.contract.ConID
}

func newInstrument(cd ContractData, c *IBClient) (ins *Instrument) {
	ins = &Instrument{contract: cd.Contract, Detail: cd.ContractDetail, client: c}
	return
}
func (c *IBClient) NewInstrument(contract Contract) (*Instrument, error) {
	if contractDataList, err := c.ReqContractDetails(contract); err != nil {
		return nil, err
	} else if len(contractDataList) == 0 {
		return nil, fmt.Errorf("No contract found")
	} else if len(contractDataList) != 1 {
		return nil, fmt.Errorf("There is ambiguity in the contract defination. Use IBClient.ReqSymbolSample method to look up")
	} else {
		contract := contractDataList[0]
		return newInstrument(contract, c), nil
	}
}

func (c *IBClient) NewInstrumentFromConID(id int) (*Instrument, error) {
	return c.NewInstrument(Contract{ConID: int64(id)})
}

func (c *IBClient) USStock(symbol string) (*Instrument, error) {
	return c.NewInstrument(Contract{Symbol: symbol, SecType: "STK", Currency: "USD", Exchange: "SMART"})
}

func (c *IBClient) USFuture(symbol string, expiration string) (*Instrument, error) {
	if to, ok := CommonFutureSymbolMap[symbol]; ok {
		symbol = to
	}
	var con Contract
	if expiration == "" {
		con = Contract{Symbol: symbol, SecType: "CONTFUT", Currency: "USD"}
	} else {
		con = Contract{Symbol: symbol, SecType: "FUT", Currency: "USD", LastTradeDateOrContractMonth: expiration}
	}
	condatas, err := c.ReqContractDetails(con)
	if err != nil {
		return nil, err
	}
	i := 0
	for _, condata := range condatas {
		if condata.Exchange != "QBALGO" {
			condatas[i] = condata
			i++
		}
	}
	condatas = condatas[:i]
	if len(condatas) == 0 {
		return nil, fmt.Errorf("ContractData request end unexpected or cancelled by user")
	}
	if len(condatas) == 1 {
		return newInstrument(condatas[0], c), nil
	} else {
		return nil, fmt.Errorf("There is ambiguity in the contract defination. Use IBClient.ReqSymbolSample method to look up")
	}
}

type TickStream struct {
	Ticks  chan Tick
	Cancel func() error
	Err    error
}

func (ins *Instrument) TickStream(tickType string) (stream *TickStream, err error) {
	id, ack, respCh, err := ins.client.reqTicker()
	if err != nil {
		return
	}
	defer close(ack)
	w := ins.client.writer
	w.writeString(outREQTICKBYTICKDATA)
	w.writeString(id)
	w.writeContract(&ins.contract)
	w.writeString(tickType)
	w.writeInt(0)
	w.writeBool(false)
	err = w.send()
	if err != nil {
		return
	}
	stream = &TickStream{make(chan Tick), nil, nil}
	stream.Cancel = func() error {
		return ins.client.reqCancel(outCANCELTICKBYTICKDATA, "", id)
	}
	go func() {
		pending := make([]Tick, 0)
		var first Tick
		var update chan Tick
		var last time.Time
		i := 1
		for {
			if len(pending) > 0 {
				first = pending[0]
				update = stream.Ticks
			}
			select {
			case msg := <-respCh:
				if msg.code[:1] == "E" {
					close(stream.Ticks)
					stream.Err = fmt.Errorf("Error%v: %v", msg.code[1:], msg.body)
					return
				}
				t := *(msg.body).(*Tick)
				if t.Time.After(last) {
					last = t.Time
					i = 1
				} else if t.Time.Before(last) {
					fmt.Println("ERROR! Tick out of order")
				} else {
					if err := t.SetShift(i); err != nil {
						close(stream.Ticks)
						stream.Err = err
					}
					i++
				}
				pending = append(pending, t)
			case update <- first:
				pending = pending[1:]
				update = nil
			}
		}
	}()
	return
}

func (ins *Instrument) reqHistorical() {
	if t := time.Now(); t.Sub(ins.lastHistoricalReq) > time.Millisecond*400 {
		ins.lastHistoricalReq = t
		return
	} else {
		t = <-time.NewTimer(ins.lastHistoricalReq.Add(time.Millisecond * 400).Sub(t)).C
		ins.lastHistoricalReq = t
		return
	}
}

func (ins *Instrument) HeadTimeStamp(whatToShow string, useRTH bool) (t time.Time, err error) {
	id, ack, respCh, err := ins.client.reqTicker()
	if err != nil {
		return
	}
	defer close(ack)
	con := ins.contract
	if con.SecType == "FUT" {
		con.SecType = "CONTFUT"
	}
	w := ins.client.writer
	w.writeString(outREQHEADTIMESTAMP)
	w.writeString(id)
	w.writeContractWithExpired(&con)
	w.writeBool(useRTH)
	w.writeString(whatToShow)
	w.writeInt(1)
	err = w.send()
	if err != nil {
		return
	}
	msg := <-respCh
	str := (msg.body).(string)
	loc, err := time.LoadLocation(ins.Detail.TimeZoneID)
	if err != nil {
		return
	}
	t, err = time.ParseInLocation("20060102  15:04:00", str, loc)
	return
}

func (ins *Instrument) HistoricalTicks(startDataTime string, endDateTime string, numberOfTicks int64, whatToShow string, useRTH bool) (ticks []Tick, err error) {
	ins.reqHistorical()
	<-ins.client.historical
	id, ack, respCh, err := ins.client.reqTicker()
	if err != nil {
		return
	}
	defer close(ack)
	if whatToShow == "BIDASK" {
		whatToShow = "BID_ASK"
	}
	con := ins.contract
	if con.SecType == "CONTFUT" {
		con.SecType = "FUT"
	}
	w := ins.client.writer
	w.writeString(outREQHISTORICALTICKS)
	w.writeString(id)
	w.writeContractWithExpired(&con)
	w.writeString(startDataTime)
	w.writeString(endDateTime)
	w.writeInt(numberOfTicks)
	w.writeString(whatToShow)
	w.writeBool(useRTH)
	w.writeBool(false)
	w.writeString("")
	err = w.send()
	if err != nil {
		return
	}
	msg := <-respCh
	if msg.code[:1] == "E" {
		err = fmt.Errorf("Error%v: %v", msg.code[1:], msg.body)
		return
	}
	ticks = (msg.body).([]Tick)
	return
}

func (ins *Instrument) HistoricalBar(endDateTime string, durationStr string, barSize string, whatToShow string, useRTH bool, keepUpToDate bool) (bars []BarData, err error) {
	ins.reqHistorical()
	<-ins.client.historical
	id, ack, respCh, err := ins.client.reqTicker()
	if err != nil {
		return
	}
	defer close(ack)
	con := ins.contract
	if con.SecType == "FUT" {
		con.SecType = "CONTFUT"
	}
	w := ins.client.writer
	w.writeString(outREQHISTORICALDATA)
	// v:="6" //serverVersion 124
	// writeString(w,v) //serverVersion 124
	w.writeString(id)
	w.writeContractWithExpired(&con)
	w.writeString(endDateTime)
	w.writeString(barSize)
	w.writeString(durationStr)
	w.writeBool(useRTH)
	w.writeString(whatToShow)
	w.writeInt(1)
	//TODO BAGS
	w.writeBool(keepUpToDate) // serverVersion 124
	w.writeString("")
	w.send()
	msg := <-respCh
	if msg.code[:1] == "E" {
		err = fmt.Errorf("Error%v: %v", msg.code[1:], msg.body)
		return
	}
	bars = (msg.body).([]BarData)
	return
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
