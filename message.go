package ibgo

import (
	"time"
)

type Message struct {
	code string
	id   string
	kind int64
	body interface{}
}

type NextValidID struct {
	ID int64
}

type ManagedAccounts struct {
	Accounts []string
}

type CurrentTime struct {
	Time time.Time
}

type HistoricalData struct {
	StartDateStr string
	EndDateStr   string
	ItemCount    int64
	Bars         []BarData
}
