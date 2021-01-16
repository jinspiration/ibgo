package ibgo

import (
	"fmt"
	"strings"
	"time"
)

type Contract struct {
	ConID                        int64
	Symbol                       string
	SecType                      string
	LastTradeDateOrContractMonth string
	Strike                       float64
	Right                        string
	Multiplier                   string
	Exchange                     string
	PrimaryExchange              string
	Currency                     string
	LocalSymbol                  string
	TradingClass                 string
	IncludeExpired               bool   //Used for Historicaldata and Contract detail
	SecIDType                    string //Used for Contract Detail	CUSIP;SEDOL;ISIN;RIC
	SecID                        string //Used for Contract Detail

	// combos legs
	ComboLegsDescription string
	ComboLegs            []ComboLeg
	// UnderComp            *UnderComp

	DeltaNeutralContract *DeltaNeutralContract
}

type ContractDescription struct {
	Contract
	DerivativeSecTypes []string
}

type ComboLeg struct {
	ContractID int64
	Ratio      int64
	Action     string
	Exchange   string
	OpenClose  int64

	// for stock legs when doing short sale
	ShortSaleSlot      int64
	DesignatedLocation string
	ExemptCode         int64 `default:"-1"`
}

type DeltaNeutralContract struct {
	ContractID int64
	Delta      float64
	Price      float64
}

type ContractData struct {
	Contract
	ContractDetail
}
type ContractDetail struct {
	MarketName         string
	MinTick            float64
	OrdeTypes          string
	ValidExchange      string
	PriceManifier      int64
	UnderConID         int64
	LongName           string
	ContractMonth      string
	Industry           string
	Category           string
	Subcategory        string
	TimeZoneID         string
	TradingHours       []Session
	LiquidHours        []Session
	EVRule             string
	EVMultiplier       int64
	MDSizeMultiplier   int64
	AggGroup           int64
	UnderSymbol        string
	UnderSecType       string
	MarketRuleIDs      string
	SecIDList          []TagValue
	RealExpirationDate string
	LastTradeTime      string
}

type BondContractData struct {
	Contract
	ContractDetail
	BondContractDetail
}
type BondContractDetail struct {
	CUSIP             string
	Ratings           string
	DescAppend        string
	BondType          string
	CouponType        string
	Callable          bool
	Putable           bool
	Coupon            int64
	Convertible       bool
	Maturity          string
	IssueDate         string
	NextOptionDate    string
	NextOptionType    string
	NextOptionPartial bool
	Notes             string
}

func (c *ContractData) Read(m *Message) {
	c.Symbol = m.readString()
	c.SecType = m.readString()
	readLastTradeDate(m, c)
	c.Strike = m.readFloat()
	c.Right = m.readString()
	c.Exchange = m.readString()
	c.Currency = m.readString()
	c.LocalSymbol = m.readString()
	c.MarketName = m.readString()
	c.TradingClass = m.readString()
	c.ConID = m.readInt()
	c.MinTick = m.readFloat()
	c.MDSizeMultiplier = m.readInt() //serverVersion 110
	c.Multiplier = m.readString()
	c.OrdeTypes = m.readString()
	c.ValidExchange = m.readString()
	c.PriceManifier = m.readInt()
	c.UnderConID = m.readInt()                                              //version 4
	c.LongName = m.readString()                                             //version 5
	c.PrimaryExchange = m.readString()                                      //version 5
	c.ContractMonth = m.readString()                                        //version 6
	c.Industry = m.readString()                                             //version 6
	c.Category = m.readString()                                             //version 6
	c.Subcategory = m.readString()                                          //version 6
	c.TimeZoneID = m.readString()                                           //version 6
	c.TradingHours = m.readSessions(ContractDetailTimeLayout, c.TimeZoneID) //version 6
	c.LiquidHours = m.readSessions(ContractDetailTimeLayout, c.TimeZoneID)  //version 6
	c.EVRule = m.readString()                                               //version 8
	c.EVMultiplier = m.readIntUnset()                                       //version 8
	if n := int(m.readInt()); n > 0 {                                       //version 7
		c.SecIDList = make([]TagValue, n)
		for i := 0; i < n; i++ {
			c.SecIDList[i] = TagValue{Tag: m.readString(), Value: m.readString()}
		}
	}
	c.AggGroup = m.readInt()              //serverVersion 121
	c.UnderSymbol = m.readString()        //serverVersion 122
	c.UnderSecType = m.readString()       //serverVersion 122
	c.MarketRuleIDs = m.readString()      //serverVersion 126
	c.RealExpirationDate = m.readString() //serverVersion 134
}

func (c *Contract) String() string {
	return fmt.Sprintf("ID:%v\tSymbol:%v\tSecType:%v\nCurrency:%v\tExchange:%v\tPrimaryExchaneg:%v\nLocalSymbol:%v\tExpire:%v\tMultiplier:%v\n", c.ConID, c.Symbol, c.SecType, c.Currency, c.Exchange, c.PrimaryExchange, c.LocalSymbol, c.LastTradeDateOrContractMonth, c.Multiplier)
}
func (c *ContractData) String() string {
	return fmt.Sprintf("%v", c.Contract)
	// return fmt.Sprintf("ID:%v\tSymbol:%v\tSecType:%v\nCurrency:%v\tExchange:%v\tPrimaryExchaneg:%v\nLocalSymbol:%v\tExpire:%v\tMultiplier:%v\n", c.ConID, c.Symbol, c.SecType, c.Currency, c.Exchange, c.PrimaryExchange, c.LocalSymbol, c.LastTradeDateOrContractMonth, c.Multiplier)
}

var ContractDetailTimeLayout = "20060102:1504"
var HistoricaldataTimeLayout = "20060102 15:04:00"

type Session struct {
	Start time.Time
	End   time.Time
}

func (m *Message) readSessions(layout string, zone string) []Session {
	strs := strings.Split(m.readString(), ";")
	sessions := make([]Session, len(strs))
	loc, _ := time.LoadLocation(zone)
	for i, str := range strs {
		if s := strings.Split(str, "-"); len(s) == 1 && strings.HasSuffix(s[0], "CLOSED") {
			sessions[i] = Session{}
		} else if len(s) == 2 {
			start, _ := time.ParseInLocation(layout, s[0], loc)
			end, _ := time.ParseInLocation(layout, s[1], loc)
			sessions[i] = Session{start, end}
		}
	}
	return sessions
}
