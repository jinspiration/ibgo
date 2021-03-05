package ibgo

import "time"

type TickLast struct {
	Time  time.Time
	Price float64
	Size  int64
	TickAttribLast
	Exchange          string
	SpecialConditions string
}

type TickBidAsk struct {
	Time     time.Time
	BidPrice float64
	AskPrice float64
	BidSize  int64
	AskSize  int64
	TickAttribBidAsk
}

type TickMidPoint struct {
	Time     time.Time
	Midpoint float64
}

// type HeadTimeStamp struct {
// 	Message
// }

// func (resp *HeadTimeStamp) id() string { return readString(resp.buf) }

// func (resp *HeadTimeStamp) read() {}

// type TickPrice struct {
// 	Message
// 	serverVersion int64
// 	TickType      int
// 	Price         float64
// 	Size          int64
// 	TickAttrib
// }

// func (resp *TickPrice) id() string {
// 	readString(resp.buf)
// 	return readString(resp.buf)
// }

// func (resp *TickPrice) read() {
// 	resp.TickType = int(readInt(resp.buf))
// 	resp.Price = readFloat(resp.buf)
// 	resp.Size = readInt(resp.buf)
// 	attrMask := readInt(resp.buf)
// 	attrib := TickAttrib{}
// 	attrib.CanAutoExecute = attrMask == 1

// 	if resp.serverVersion >= vMINSERVERVERPASTLIMIT {
// 		attrib.CanAutoExecute = attrMask&0x1 != 0
// 		attrib.PastLimit = attrMask&0x2 != 0
// 		if resp.serverVersion >= vMINSERVERVERPREOPENBIDASK {
// 			attrib.PreOpen = attrMask&0x4 != 0
// 		}
// 	}
// 	resp.TickAttrib = attrib
// }

// type TickSize struct {
// 	Message
// }

// type HistoricalData struct {
// 	HistoricalDataUpdate
// 	StartDate string
// 	EndDate   string
// 	ItemCount int64
// }

// func (resp *HistoricalData) id() string {
// 	// TOCHECK
// 	// if serverVersion<vMINSERVERVERSYNTREALTIMEBARS{
// 	// 	readString(resp.buf)
// 	// }
// 	return readString(resp.buf)
// }

// func (resp *HistoricalData) read() {
// 	resp.StartDate = readString(resp.buf)
// 	resp.EndDate = readString(resp.buf)
// 	resp.ItemCount = readInt(resp.buf)
// 	resp.BarData = *newBarData(resp.buf)
// }

// type HistoricalDataUpdate struct {
// 	Message
// 	BarData
// }

// func (resp *HistoricalDataUpdate) id() string {
// 	return readString(resp.buf)
// }

// func (resp *HistoricalDataUpdate) read() {
// 	resp.BarData = *newBarData(resp.buf)
// }

// type TickGeneric struct {
// 	Message
// 	TickType int
// 	Value    float64
// }

// func (resp *TickGeneric) id() string {
// 	readString(resp.buf)
// 	return readString(resp.buf)
// }

// func (resp *TickGeneric) read() {
// 	resp.TickType = int(readInt(resp.buf))
// 	resp.Value = readFloat(resp.buf)
// }

// type TickString struct {
// 	Message
// 	TickType int
// 	Value    string
// }

// func (resp *TickString) id() string {
// 	readString(resp.buf)
// 	return readString(resp.buf)
// }

// func (resp *TickString) read() {
// 	resp.TickType = int(readInt(resp.buf))
// 	resp.Value = readString(resp.buf)
// }

// type TickEFP struct {
// 	Message
// 	TickType                 int
// 	BasisPoints              float64
// 	FormattedBasisPoints     string
// 	TotalDividends           float64
// 	HoldDays                 int64
// 	FutureLastTradeDate      string
// 	DividentImpact           float64
// 	DividendsTolastTradeDate float64
// }

// type MarketDepth struct {
// 	Message
// 	Position  int64
// 	Operation int64
// 	Side      int64
// 	Price     float64
// 	Size      int64
// }

// func (resp *MarketDepth) id() string {
// 	readString(resp.buf)
// 	return readString(resp.buf)
// }

// func (resp *MarketDepth) read() {
// 	resp.Position = readInt(resp.buf)
// 	resp.Operation = readInt(resp.buf)
// 	resp.Side = readInt(resp.buf)
// 	resp.Price = readFloat(resp.buf)
// 	resp.Size = readInt(resp.buf)
// }

// type MarketDepthl2 struct {
// 	Message
// 	Position     int64
// 	MarketMaker  string
// 	Operation    int64
// 	Side         int64
// 	Price        float64
// 	Size         int64
// 	IsSmartDepth bool
// }

// func (resp *MarketDepthl2) id() string {
// 	readString(resp.buf)
// 	return readString(resp.buf)
// }

// func (resp *MarketDepthl2) read() {
// 	resp.Position = readInt(resp.buf)
// 	resp.MarketMaker = readString(resp.buf)
// 	resp.Operation = readInt(resp.buf)
// 	resp.Side = readInt(resp.buf)
// 	resp.Price = readFloat(resp.buf)
// 	resp.Size = readInt(resp.buf)
// 	resp.IsSmartDepth = readBool(resp.buf)
// }
