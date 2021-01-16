package ibgo

import (
	"encoding/json"
	"time"
)

type Tick struct {
	Time     time.Time `json:"time"`
	Last     float64   `json:"last,omitempty"`
	Size     int64     `json:"size,omitempty"`
	Bid      float64   `json:"bid,omitempty"`
	Ask      float64   `json:"ask,omitempty"`
	BidSize  int64     `json:"bidsize,omitempty"`
	AskSize  int64     `json:"asksize,omitempty"`
	Midpoint float64   `json:"midpoint,omitempty"`
	Mask     int64     `json:"mask"`
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

func (t *Tick) Read(m *Message) {
	typ := m.readInt()
	t.Time = time.Unix(m.readInt(), 0)
	switch typ {
	case 1, 2:
		t.Last = m.readFloat()
		t.Size = m.readInt()
		mask := m.readInt()
		t.Mask |= mask << 3
	case 3:
		t.Bid = m.readFloat()
		t.Ask = m.readFloat()
		t.BidSize = m.readInt()
		t.AskSize = m.readInt()
		mask := m.readInt()
		t.Mask |= mask << 5
	case 4:
		t.Midpoint = m.readFloat()
	}
}

func (t *Tick) SetMilli(n int) {
	// fmt.Println(t.Time.UnixNano())
	t.Time = t.Time.Add(time.Duration(n) * time.Millisecond)
	// fmt.Println(t.Time.UnixNano())
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
