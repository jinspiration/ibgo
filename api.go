package ibgo

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

var ErrDisconnected = errors.New("Client disconnected")

func (c *IBClient) reqStatic(key string) (ack chan struct{}, respCh chan *Message, err error) {
	if c.status < 3 {
		err = ErrDisconnected
		return
	}
	select {
	case q := <-c.static:
		respCh = make(chan *Message)
		if key != "" {
			c.registry["*"+key] = respCh
		}
		ack = q.ack
		return
	case <-time.After(time.Second):
		err = errTimeOut
		return
	}
}

func (c *IBClient) reqTicker() (id string, ack chan struct{}, respCh chan *Message, err error) {
	if c.status < 3 {
		err = ErrDisconnected
		return
	}
	select {
	case q := <-c.ticker:
		id, ack = strconv.FormatInt(q.id, 10), q.ack
		respCh = make(chan *Message)
		c.registry[id] = respCh
		ack = q.ack
		return
	case <-time.After(time.Second):
		err = errTimeOut
		return
	}
}

func (c *IBClient) reqOrder() (id string, ack chan struct{}, respCh chan *Message, err error) {
	if c.status < 3 {
		err = ErrDisconnected
		return
	}
	select {
	case q := <-c.order:
		id, ack = strconv.FormatInt(q.id, 10), q.ack
		respCh = make(chan *Message)
		c.registry[id] = respCh
		ack = q.ack
		return
	case <-time.After(time.Second):
		err = errTimeOut
		return
	}
}

func (c *IBClient) reqCancel(code string, version string, id string) (err error) {
	if c.status < 3 {
		err = ErrDisconnected
		return
	}
	var ack chan struct{}
	select {
	case q := <-c.static:
		ack = q.ack
	case <-time.After(time.Second):
		err = errTimeOut
		return
	}
	defer close(ack)
	c.writer.writeString(code)
	if version != "" {
		c.writer.writeString(version)
	}
	c.writer.writeString(id)
	err = c.writer.send()
	return
}

func (c *IBClient) ReqCurrentTime() (t time.Time, err error) {
	ack, respCh, err := c.reqStatic(inCURRENTTIME)
	if err != nil {
		return
	}
	defer close(ack)
	c.writer.writeString(outREQCURRENTTIME)
	c.writer.writeString("1")
	err = c.writer.send()
	if err != nil {
		return
	}
	msg := <-respCh
	t = (msg.body).(*CurrentTime).Time
	return
}

func (c *IBClient) ReqContractDetails(con Contract) (response []ContractData, err error) {
	id, ack, respCh, err := c.reqTicker()
	if err != nil {
		return
	}
	defer close(ack)
	c.writer.writeString(outREQCONTRACTDATA)
	c.writer.writeString("8")
	c.writer.writeString(id)
	c.writer.writeContractWithFull(&con)
	err = c.writer.send()
	if err != nil {
		return
	}
	response = make([]ContractData, 0)
	for msg := range respCh {
		if msg.code[0] == 'E' {
			err = fmt.Errorf("Error[%v]:%v", msg.code, msg.body)
			return
		}
		if msg.code == inCONTRACTDATAEND {
			return
		}
		data := (msg.body).(*ContractData)
		response = append(response, *data)
	}
	return
}

func (c *IBClient) ReqHistoricalData(con *Contract, endDateTime string, durationStr string, barSizeSetting string, whatToShow string, useRTH bool, formatDate int64, keepUpToDate bool) (bars []BarData, err error) {
	id, ack, respCh, err := c.reqTicker()
	if err != nil {
		return
	}
	defer close(ack)
	c.writer.writeString(outREQHISTORICALDATA)
	// v:="6" //serverVersion 124
	// writeString(w,v) //serverVersion 124
	c.writer.writeString(id)
	c.writer.writeContractWithExpired(con)
	c.writer.writeString(endDateTime)
	c.writer.writeString(barSizeSetting)
	c.writer.writeString(durationStr)
	c.writer.writeBool(useRTH)
	c.writer.writeString(whatToShow)
	c.writer.writeInt(formatDate)
	//TODO BAGS
	c.writer.writeBool(keepUpToDate) // serverVersion 124
	c.writer.writeString("")
	c.writer.send()
	msg := <-respCh
	bars = (msg.body).(*HistoricalData).Bars
	return
}
