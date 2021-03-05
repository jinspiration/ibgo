package ibgo

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Tick struct {
	Time              time.Time `json:"time"`
	Last              float64   `json:"last,omitempty"`
	Size              int64     `json:"size,omitempty"`
	Bid               float64   `json:"bid,omitempty"`
	Ask               float64   `json:"ask,omitempty"`
	BidSize           int64     `json:"bidsize,omitempty"`
	AskSize           int64     `json:"asksize,omitempty"`
	Midpoint          float64   `json:"midpoint,omitempty"`
	Mask              int64     `json:"mask"`
	Exchange          string    `json:"exchange,omitempty"`
	SpecialConditions string    `json:"specialcondition,omitempty"`
}

func (t *Tick) MarshalJSON() ([]byte, error) {
	mt := time.Time(t.Time).UnixNano() / int64(time.Millisecond)
	type alias Tick
	return json.Marshal(struct {
		Time int64 `json:"time"`
		*alias
	}{mt, (*alias)(t)})
}

type TickMask = int64

const (
	RTH TickMask = 1 << iota
	SessionStart
	SessionEnd
	PastLimit
	UnReported
	BidPastLow
	AskPastHigh
)

type BarData struct {
	Time       string
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Volume     int64
	TradeCount int64
}

func (bar *BarData) ToLineProtocol() (point string) {
	return
}
func (bar *BarData) ToCSV() (record []string) {
	record = make([]string, 7)
	record[0] = bar.Time
	record[1] = strconv.FormatFloat(bar.Open, 'g', 10, 64)
	record[2] = strconv.FormatFloat(bar.High, 'g', 10, 64)
	record[3] = strconv.FormatFloat(bar.Low, 'g', 10, 64)
	record[4] = strconv.FormatFloat(bar.Close, 'g', 10, 64)
	record[5] = strconv.FormatInt(bar.Volume, 10)
	record[6] = strconv.FormatInt(bar.TradeCount, 10)
	return
}

type TypedTick struct {
	Time     time.Time
	TickType int64
	Price    float64
	Size     int64
	Mask     int64
}

type TypedTickString struct {
	Time     time.Time
	TickType int64
	Value    string
	Mask     int64
}

// func (t *Tick) Read(m *Message) {
// 	typ := m.readInt()
// 	t.Time = time.Unix(m.readInt(), 0)
// 	switch typ {
// 	case 1, 2:
// 		t.Last = m.readFloat()
// 		t.Size = m.readInt()
// 		mask := m.readInt()
// 		t.Mask |= mask << 3
// 	case 3:
// 		t.Bid = m.readFloat()
// 		t.Ask = m.readFloat()
// 		t.BidSize = m.readInt()
// 		t.AskSize = m.readInt()
// 		mask := m.readInt()
// 		t.Mask |= mask << 5
// 	case 4:
// 		t.Midpoint = m.readFloat()
// 	}
// }

func (t *Tick) SetShift(i int) (err error) {
	// fmt.Println(t.Time.UnixNano())
	if i < 1000 {
		t.Time = t.Time.Add(time.Duration(i) * time.Millisecond)
	} else if i < 1999 {
		t.Time = t.Time.Add(time.Millisecond * 999).Add(time.Duration(i-999) * time.Microsecond)
	} else {
		err = fmt.Errorf("Too many tick arrives in one second")
	}
	return
}
func (t *Tick) SetRTH() {
	t.Mask |= RTH
}

func (t *Tick) SetSessionStart() {
	t.Mask |= SessionStart
}

func (t *Tick) SetSessionEnd() {
	t.Mask |= SessionEnd
}
