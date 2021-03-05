package ibgo

import (
	"fmt"
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

type ContractDataEnd struct{}

func (c Contract) String() string {
	return fmt.Sprintf("ID: %d\tSymbol: %s\tSecType: %v\tCurrency: %v\tExchange: %v\tPrimaryExchaneg: %v\tLocalSymbol: %v\tExpire: %v\tMultiplier: %v", c.ConID, c.Symbol, c.SecType, c.Currency, c.Exchange, c.PrimaryExchange, c.LocalSymbol, c.LastTradeDateOrContractMonth, c.Multiplier)
}

func (c ContractData) String() string {
	return c.Contract.String()
}

var ContractDetailTimeLayout = "20060102:1504"
var HistoricaldataTimeLayout = "20060102 15:04:00"
var HeadTimeStampTimeLayout = "20060102  15:04:00"

type Session struct {
	Start time.Time
	End   time.Time
}
