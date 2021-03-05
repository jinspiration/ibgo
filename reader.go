package ibgo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var ErrBadMsgLength = fmt.Errorf("Bad Msg Length")

type msgReader struct {
	buf           []byte
	curSize       int
	rd            *net.TCPConn
	r, w          int
	serverVersion int64
}

func newReader(conn *net.TCPConn) *msgReader {
	return &msgReader{buf: make([]byte, 4096), rd: conn}
}

func (rd *msgReader) fill() error {
	if rd.r > 0 {
		copy(rd.buf, rd.buf[rd.r:rd.w])
		rd.w -= rd.r
		rd.r = 0
	}
	if rd.w >= len(rd.buf) {
		panic("eReader: tired to fill full buffer")
	}
	for i := 100; i > 0; i-- {
		n, err := rd.rd.Read(rd.buf[rd.w:])
		if n < 0 {
			panic("eReader: negative read")
		}
		rd.w += n
		if err != nil {
			return err
		}
		if n > 0 {
			return nil
		}
	}
	return io.ErrNoProgress
}

func (rd *msgReader) peekCode() (code string, err error) {
	if err = rd.readSize(); err != nil {
		return
	}
	token, err := rd.nextToken()
	if err != nil {
		return
	}
	code = string(token[:len(token)-1])
	return
}

func (rd *msgReader) readSize() (err error) {
	for rd.w-rd.r < 4 {
		err = rd.fill()
		if err != nil {
			return
		}
	}
	rd.curSize = int(binary.BigEndian.Uint32(rd.buf[rd.r : rd.r+4]))
	rd.r += 4
	return
}

func (rd *msgReader) nextString() (str string) {
	token, err := rd.nextToken()
	if err != nil {
		panic(err)
	}
	l := len(token)
	strH := (*reflect.StringHeader)(unsafe.Pointer(&str))
	strH.Data = uintptr(unsafe.Pointer(&token[0]))
	strH.Len = l - 1
	rd.curSize -= l
	rd.r += l
	return str
}

func (rd *msgReader) nextToken() (token []byte, err error) {
	s := 0
	for {
		if i := bytes.IndexByte(rd.buf[rd.r+s:rd.w], 0); i >= 0 {
			i += s
			token = rd.buf[rd.r : rd.r+i+1]
			return
		}
		s = rd.w - rd.r
		if s >= len(rd.buf) {
			err = fmt.Errorf("eReader: buffer full. Too large of a single field")
			return
		}
		if err = rd.fill(); err != nil {
			return
		}
	}
}

func (rd *msgReader) discard() error {
	token, err := rd.nextToken()
	if err != nil {
		return err
	}
	l := len(token)
	rd.r += l
	rd.curSize -= l
	return nil
}

func (rd *msgReader) readString() string {
	token, err := rd.nextToken()
	if err != nil {
		panic(err)
	}
	l := len(token)
	rd.curSize -= l
	rd.r += l
	return string(token[:l-1])
}

func (rd *msgReader) readInt() int64 {
	s := rd.nextString()
	if s == "" {
		return math.MaxInt64
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}

func (rd *msgReader) readFloat() float64 {
	s := rd.nextString()
	if s == "" {
		return math.MaxFloat64
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return f
}

func (rd *msgReader) readBool() bool {
	s := rd.nextString()
	return s == "1"
}

func (rd *msgReader) readUnix() time.Time {
	u := rd.readInt()
	return time.Unix(u, 0)
}

func (rd *msgReader) readBar() (bar BarData) {
	bar = BarData{}
	bar.Time = rd.readString()
	bar.Open = rd.readFloat()
	bar.High = rd.readFloat()
	bar.Low = rd.readFloat()
	bar.Close = rd.readFloat()
	bar.Volume = rd.readInt()
	rd.discard()
	// bar.Average = rd.readFloat()
	bar.TradeCount = rd.readInt()
	return
}

func (rd *msgReader) readFull() (dst []byte, err error) {
	if rd.curSize == 0 {
		err = rd.readSize()
		if err != nil {
			return
		}
	}
	dst = make([]byte, rd.curSize)
	p := copy(dst[len(dst)-rd.curSize:], rd.buf[rd.r:rd.w])
	rd.r += p
	rd.curSize -= p
	for rd.curSize > 0 && err == nil {
		err = rd.fill()
		p := copy(dst[len(dst)-rd.curSize:], rd.buf[rd.r:rd.w])
		rd.r += p
		rd.curSize -= p
	}
	if rd.curSize == 0 {
		return dst, nil
	}
	return
}

func (rd *msgReader) readMessage() (m *Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic during message reading: %s", r)
		}
	}()
	code, err := rd.peekCode()
	if decoderF, ok := decoderMap[code]; ok {
		m = &Message{}
		decoderF(m, rd)
		return
	}
	var unknown []byte
	unknown, err = rd.readFull()
	if err != nil {
		return
	}
	fmt.Println("Unknown msg code:", code, "whole msg:", string(unknown))
	return
}

func readLastTradeDate(rd *msgReader, c *ContractData) {
	if str := rd.readString(); str != "" {
		if sl := strings.Split(str, " "); len(sl) > 1 {
			c.LastTradeTime = str
			c.LastTradeDateOrContractMonth = sl[0]
		} else if len(sl) > 0 {
			c.LastTradeDateOrContractMonth = sl[0]
		}
	}
}

func (rd *msgReader) readSessions(layout string, zone string) []Session {
	strs := strings.Split(rd.readString(), ";")
	sessions := make([]Session, 0)
	loc, _ := time.LoadLocation(zone)
	for _, str := range strs {
		if s := strings.Split(str, "-"); len(s) == 1 && strings.HasSuffix(s[0], "CLOSED") {
			continue
		} else if len(s) == 2 {
			start, _ := time.ParseInLocation(layout, s[0], loc)
			end, _ := time.ParseInLocation(layout, s[1], loc)
			sessions = append(sessions, Session{start, end})
		}
	}
	return sessions
}

type decoderFunc = func(*Message, *msgReader)

var decoderMap = map[string]decoderFunc{
	inNEXTVALIDID:           decodeNextValidID,
	inMANAGEDACCTS:          decodeManagedAccounts,
	inCURRENTTIME:           decodeCurrentTime,
	inERRMSG:                decodeErrorMsg,
	inCONTRACTDATA:          decodeContractData,
	inCONTRACTDATAEND:       decodeContractDataEnd,
	inTICKBYTICK:            decodeTickByTick,
	inHEADTIMESTAMP:         decodeHeadTimeStamp,
	inHISTORICALDATA:        decodeHistoricalData,
	inHISTORICALTICKS:       decodeHistoricalTick,
	inHISTORICALTICKSLAST:   decodeHistoricalTick,
	inHISTORICALTICKSBIDASK: decodeHistoricalTick,
}

// func (rd *msgReader) readMessage() (m *Message, err error) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			err = fmt.Errorf("Panic during message reading: %s", r)
// 		}
// 	}()
// 	token, err := rd.peekCode()
// 	m = &Message{}
// 	switch string(token) {
// 	case inNEXTVALIDID:
// 		decodeNextValidID(m, rd)
// 	case inMANAGEDACCTS:
// 		decodeManagedAccounts(m, rd)
// 	case inCURRENTTIME:
// 		decodeCurrentTime(m, rd)
// 	case inCONTRACTDATA:
// 		decodeContractData(m, rd)
// 	case inCONTRACTDATAEND:
// 		decodeContractDataEnd(m, rd)
// 	default:
// 		// fmt.Println(b.curSize, b.r, b.w)
// 		var unknown []byte
// 		unknown, err = rd.readFull()
// 		if err != nil {
// 			return
// 		}
// 		fmt.Println("Unknown msg code:", string(token), "whole msg:", string(unknown))
// 		return nil, err
// 	}
// 	if rd.curSize != 0 {
// 		err = ErrBadMsgLength
// 	}
// 	return
// }
