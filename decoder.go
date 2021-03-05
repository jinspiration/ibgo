package ibgo

import (
	"strings"
)

func decodeNextValidID(m *Message, rd *msgReader) {
	m.code = rd.readString()
	rd.discard()
	m.id = "*" + inNEXTVALIDID
	m.body = &NextValidID{rd.readInt()}
}

func decodeManagedAccounts(m *Message, rd *msgReader) {
	m.code = rd.readString()
	rd.discard()
	m.id = "*" + inMANAGEDACCTS
	sl := strings.Split(rd.readString(), ",")
	m.body = &ManagedAccounts{sl}
}

func decodeCurrentTime(m *Message, rd *msgReader) {
	m.code = rd.readString()
	rd.discard()
	m.id = "*" + inCURRENTTIME
	m.body = &CurrentTime{rd.readUnix()}
}

func decodeErrorMsg(m *Message, rd *msgReader) {
	rd.discard()
	rd.discard()
	m.id = rd.readString()
	m.code = "E" + rd.readString()
	m.body = rd.readString()
}
func decodeContractData(m *Message, rd *msgReader) {
	m.code = rd.readString()
	rd.discard()
	m.id = rd.readString()
	c := &ContractData{}
	m.body = c
	c.Symbol = rd.readString()
	c.SecType = rd.readString()
	readLastTradeDate(rd, c)
	c.Strike = rd.readFloat()
	c.Right = rd.readString()
	c.Exchange = rd.readString()
	c.Currency = rd.readString()
	c.LocalSymbol = rd.readString()
	c.MarketName = rd.readString()
	c.TradingClass = rd.readString()
	c.ConID = rd.readInt()
	c.MinTick = rd.readFloat()
	c.MDSizeMultiplier = rd.readInt() //serverVersion 110
	c.Multiplier = rd.readString()
	c.OrdeTypes = rd.readString()
	c.ValidExchange = rd.readString()
	c.PriceManifier = rd.readInt()
	c.UnderConID = rd.readInt()                                              //version 4
	c.LongName = rd.readString()                                             //version 5
	c.PrimaryExchange = rd.readString()                                      //version 5
	c.ContractMonth = rd.readString()                                        //version 6
	c.Industry = rd.readString()                                             //version 6
	c.Category = rd.readString()                                             //version 6
	c.Subcategory = rd.readString()                                          //version 6
	c.TimeZoneID = rd.readString()                                           //version 6
	c.TradingHours = rd.readSessions(ContractDetailTimeLayout, c.TimeZoneID) //version 6
	c.LiquidHours = rd.readSessions(ContractDetailTimeLayout, c.TimeZoneID)  //version 6
	c.EVRule = rd.readString()                                               //version 8
	c.EVMultiplier = rd.readInt()                                            //version 8
	if n := int(rd.readInt()); n > 0 {                                       //version 7
		c.SecIDList = make([]TagValue, n)
		for i := 0; i < n; i++ {
			c.SecIDList[i] = TagValue{Tag: rd.readString(), Value: rd.readString()}
		}
	}
	c.AggGroup = rd.readInt()              //serverVersion 121
	c.UnderSymbol = rd.readString()        //serverVersion 122
	c.UnderSecType = rd.readString()       //serverVersion 122
	c.MarketRuleIDs = rd.readString()      //serverVersion 126
	c.RealExpirationDate = rd.readString() //serverVersion 134
}

func decodeContractDataEnd(m *Message, rd *msgReader) {
	m.code = rd.readString()
	rd.discard()
	m.id = rd.readString()
}

func decodeTickByTick(m *Message, rd *msgReader) {
	m.code = rd.readString()
	m.id = rd.readString()
	switch tickType := rd.readInt(); tickType {
	case 1, 2:
		t := &Tick{}
		m.body = t
		t.Time = rd.readUnix()
		t.Last = rd.readFloat()
		t.Size = rd.readInt()
		mask := rd.readInt()
		t.Mask |= mask << 3
		t.Exchange = rd.readString()
		t.SpecialConditions = rd.readString()
	case 3:
		t := &Tick{}
		m.body = t
		t.Time = rd.readUnix()
		t.Bid = rd.readFloat()
		t.BidSize = rd.readInt()
		t.Ask = rd.readFloat()
		t.AskSize = rd.readInt()
		mask := rd.readInt()
		t.Mask |= mask << 5
	case 4:
		t := &Tick{}
		m.body = t
		t.Time = rd.readUnix()
		t.Midpoint = rd.readFloat()
	}
}

func decodeHeadTimeStamp(m *Message, rd *msgReader) {
	m.code = rd.readString()
	m.id = rd.readString()
	m.body = rd.readString()
}
func decodeHistoricalData(m *Message, rd *msgReader) {
	m.code = rd.readString()
	m.id = rd.readString()
	// h := &HistoricalData{}
	// m.body = h
	// h.StartDateStr = rd.readString()
	// h.EndDateStr = rd.readString()
	// h.ItemCount = rd.readInt()
	rd.discard()
	rd.discard()
	bars := make([]BarData, int(rd.readInt()))
	m.body = bars
	// h.Bars = make([]BarData, int(h.ItemCount))
	for i := 0; i < len(bars); i++ {
		bars[i] = rd.readBar()
	}
}

func decodeHistoricalTick(m *Message, rd *msgReader) {
	m.code = rd.readString()
	m.id = rd.readString()
	l := int(rd.readInt())
	ticks := make([]Tick, l)
	m.body = ticks
	switch m.code {
	case inHISTORICALTICKS:
		for i := 0; i < l; i++ {
			ticks[i].Time = rd.readUnix()
			ticks[i].Midpoint = rd.readFloat()
			ticks[i].Size = rd.readInt()
		}
	case inHISTORICALTICKSBIDASK:
		for i := 0; i < l; i++ {
			ticks[i].Time = rd.readUnix()
			ticks[i].Mask |= rd.readInt() << 5
			ticks[i].Bid = rd.readFloat()
			ticks[i].Ask = rd.readFloat()
			ticks[i].BidSize = rd.readInt()
			ticks[i].AskSize = rd.readInt()
		}
	case inHISTORICALTICKSLAST:
		for i := 0; i < l; i++ {
			ticks[i].Time = rd.readUnix()
			ticks[i].Mask |= rd.readInt() << 3
			ticks[i].Last = rd.readFloat()
			ticks[i].Size = rd.readInt()
			ticks[i].Exchange = rd.readString()
			ticks[i].SpecialConditions = rd.readString()
		}
	}

}
