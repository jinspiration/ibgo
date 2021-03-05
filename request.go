package ibgo

import (
	"bytes"
)

type RequestType struct {
	kind         string
	cancelable   bool
	endCode      string
	expectMsgCnt int
}

type Request interface {
	writeBody(w *bytes.Buffer, id string, serverVersion int64)
	typ() RequestType
}

type CancelRequest interface {
	Request
	writeCancel(w *bytes.Buffer, id string)
}

/*
	Market Data
*/

type ReqMktData struct {
	Contract           *Contract
	GenericTickList    string
	Snapshot           bool
	RegulatorySnapshot bool
}

func (req ReqMktData) typ() RequestType { return RequestType{"TICKER", true, "", 3} }

// func (req ReqMktData) writeBody(w *bytes.Buffer, id string, serverVersion int64) {
// 	v := "11"
// 	writeString(w, outREQMKTDATA)
// 	writeString(w, v)
// 	writeString(w, id)
// 	writeContract(w, req.Contract)
// 	// TODO combo and deltaneutral
// 	writeBool(w, false)
// 	writeString(w, req.GenericTickList)
// 	writeBool(w, req.Snapshot)
// 	writeBool(w, req.RegulatorySnapshot)
// 	writeString(w, "")
// }

func (req ReqMktData) writeCancel(w *bytes.Buffer, id string) {
	v := "2"
	writeString(w, outCANCELMKTDATA)
	writeString(w, v)
	writeString(w, id)
}

type ReqTickByTickData struct {
	Contract      *Contract
	TickType      string
	NumberofTicks int64
	IgnoreSize    bool
}

func (req ReqTickByTickData) typ() RequestType { return RequestType{"TICKER", true, "", 3} }

// func (req ReqTickByTickData) writeBody(w *bytes.Buffer, id string, serverVersion int64) {
// 	writeString(w, outREQTICKBYTICKDATA)
// 	writeString(w, id)
// 	writeContract(w, req.Contract)
// 	writeString(w, req.TickType)
// 	if serverVersion >= vMINSERVERVERTICKBYTICKIGNORESIZE {
// 		writeInt(w, req.NumberofTicks)
// 		writeBool(w, req.IgnoreSize)
// 	}
// }

func (req ReqTickByTickData) writeCancel(w *bytes.Buffer, id string) {
	writeString(w, outCANCELTICKBYTICKDATA)
	writeString(w, id)
}

/*
	Historical Data
*/

type ReqHeadTimeStamp struct {
	// Request
	Contract   *Contract
	WhatToShow string
	UseRTH     int64
	FormatDate int64
}

func (req ReqHeadTimeStamp) typ() RequestType { return RequestType{} }

// func (req ReqHeadTimeStamp) writeBody(w *bytes.Buffer, id string, serverVersion int64) {
// 	writeString(w, outREQHEADTIMESTAMP)
// 	writeString(w, id)
// 	writeContractWithExpired(w, req.Contract)
// 	writeInt(w, req.UseRTH)
// 	writeString(w, req.WhatToShow)
// 	writeInt(w, req.FormatDate)
// }

type ReqContractDetail struct {
	Contract Contract
}

func (req ReqContractDetail) typ() RequestType {
	return RequestType{"TICKER", true, inCONTRACTDATAEND, 2}
}

// func (req ReqContractDetail) writeBody(w *bytes.Buffer, id string, serverVersion int64) {
// 	v := "8"
// 	writeString(w, outREQCONTRACTDATA)
// 	writeString(w, v)
// 	writeString(w, id)
// 	writeContractFull(w, &req.Contract)
// }

type ReqHistoricalTicks struct {
	Contract      *Contract
	StartDateTime string
	EndDateTime   string
	NumberofTicks int64
	WhatToShow    string
	UseRTH        bool
	IgnoreSize    bool
}

func (req ReqHistoricalTicks) typ() RequestType { return RequestType{"HISTORICAL", true, "", 3} }

// func (req ReqHistoricalTicks) writeBody(w *bytes.Buffer, id string, serverVersion int64) {
// 	writeString(w, outREQHISTORICALTICKS)
// 	writeString(w, id)
// 	writeContractWithExpired(w, req.Contract)
// 	writeString(w, req.StartDateTime)
// 	writeString(w, req.EndDateTime)
// 	writeInt(w, req.NumberofTicks)
// 	writeString(w, req.WhatToShow)
// 	writeBool(w, req.UseRTH)
// 	writeBool(w, req.IgnoreSize)
// 	writeString(w, "")
// }

type ReqHistoricalData struct {
	Contract       *Contract
	EndDateTime    string
	DurationStr    string
	BarSizeSetting string
	WhatToShow     string
	UseRTH         int64
	FormatData     int64
	KeepUpToDate   bool
}

func (req ReqHistoricalData) typ() RequestType { return RequestType{"HISTORICAL", true, "", 3} }

// func (req ReqHistoricalData) writeBody(w *bytes.Buffer, id string, serverVersion int64) {
// 	// v:="6" //serverVersion 124
// 	writeString(w, outREQHISTORICALDATA)
// 	// writeString(w,v) //serverVersion 124
// 	writeString(w, id)
// 	writeContractWithExpired(w, req.Contract)
// 	writeString(w, req.EndDateTime)
// 	writeString(w, req.BarSizeSetting)
// 	writeString(w, req.DurationStr)
// 	writeInt(w, req.UseRTH)
// 	writeString(w, req.WhatToShow)
// 	writeInt(w, req.FormatData)
// 	//TODO BAGS
// 	writeBool(w, req.KeepUpToDate) // serverVersion 124
// 	writeString(w, "")
// }

//static request
type ReqCurrentTime struct{}

func (req ReqCurrentTime) typ() RequestType {
	return RequestType{"*" + outREQCURRENTTIME, false, "", 1}
}
func (req ReqCurrentTime) writeBody(w *bytes.Buffer, id string, serverVersion int64) {
	// req.buf.Write()
	v := "1"
	writeString(w, outREQCURRENTTIME)
	writeString(w, v)
}

type ReqSetServerLogLevel struct {
	Loglevel int64
}
