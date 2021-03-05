package ibgo

import (
	"time"
)

type TickAttrib struct {
	CanAutoExecute bool
	PastLimit      bool
	PreOpen        bool
}

type TickAttribBidAsk struct {
	BidPastLow  bool
	AskPastHigh bool
}

type TickAttribLast struct {
	PastLimit  bool
	UnReported bool
}

// func newBarData(buf *bytes.Buffer) *BarData {
// 	resp := &BarData{}
// 	resp.Date = readString(buf)
// 	resp.Open = readFloat(buf)
// 	resp.High = readFloat(buf)
// 	resp.Low = readFloat(buf)
// 	resp.Close = readFloat(buf)
// 	resp.Volume = readInt(buf)
// 	resp.Average = readFloat(buf)
// 	resp.BarCount = readInt(buf)
// 	return resp
// }

type RealtimeBar struct {
	Time    time.Time
	Endtime time.Time
	Date    string
	Open    float64
	High    float64
	Low     float64
	Close   float64
	Volume  int64
	Wap     int64
	Average float64
}
