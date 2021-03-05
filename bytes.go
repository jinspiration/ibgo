package ibgo

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

const sepByte byte = '\x00'

// func readSize(r *bufio.Reader) (size int, err error) {
// 	l := make([]byte, 4)
// 	if _, err = r.Read(l); err != nil {
// 		return
// 	}
// 	size = int(binary.BigEndian.Uint32(l))
// 	return
// }

// func readSingleMsg(r *bufio.Reader) (b []byte, err error) {
// 	l := make([]byte, 4)
// 	if _, err = r.Read(l); err != nil {
// 		return
// 	}
// 	size := int(binary.BigEndian.Uint32(l))
// 	b = make([]byte, size)
// 	if _, err = io.ReadFull(r, b); err != nil {
// 		return
// 	}
// 	return
// }

// func (resp *Message) readString() string {
// 	b, err := resp.buf.ReadBytes(sepByte)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return string(b[:len(b)-1])
// }

// func (resp *Message) readInt() int64 {
// 	s := resp.readString()
// 	i, err := strconv.ParseInt(s, 10, 64)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return i
// }

// func (resp *Message) readFloat() float64 {
// 	s := resp.readString()
// 	f, err := strconv.ParseFloat(s, 10)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return f
// }

// func (resp *Message) readBool() bool {
// 	s := resp.readString()
// 	return !(s == "0")
// }

// func (resp *Message) readTime() time.Time {
// 	s := resp.readString()
// 	unix, err := strconv.ParseInt(s, 10, 64)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return time.Unix(unix, 0)
// }

// func (resp *Message) readIntUnset() int64 {
// 	s := resp.readString()
// 	if s == "" {
// 		return math.MaxInt64
// 	}
// 	i, err := strconv.ParseInt(s, 10, 64)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return i
// }

// func (resp *Message) readFloatUnset() float64 {
// 	s := resp.readString()
// 	if s == "" {
// 		return math.MaxFloat64
// 	}
// 	f, err := strconv.ParseFloat(s, 10)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return f
// }

// func (resp *Message) readStringList(sep string) []string {
// 	s := resp.readString()
// 	return strings.Split(s, sep)
// }

func readString(r *bufio.Reader) string {
	b, err := r.ReadBytes(sepByte)
	if err != nil {
		panic(err)
	}
	return string(b[:len(b)-1])
}

func readInt(r *bufio.Reader) int64 {
	s := readString(r)
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}

func readFloat(r *bufio.Reader) float64 {
	s := readString(r)
	f, err := strconv.ParseFloat(s, 10)
	if err != nil {
		panic(err)
	}
	return f
}

func readBool(r *bufio.Reader) bool {
	s := readString(r)
	return !(s == "0")
}

func readIntUnset(r *bufio.Reader) int64 {
	s := readString(r)
	if s == "" {
		return math.MaxInt64
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}

func readFloatUnset(r *bufio.Reader) float64 {
	s := readString(r)
	if s == "" {
		return math.MaxFloat64
	}
	f, err := strconv.ParseFloat(s, 10)
	if err != nil {
		panic(err)
	}
	return f
}

func readStringList(r *bufio.Reader, sep string) []string {
	s := readString(r)
	return strings.Split(s, sep)
}

func splitMsgBytesWithLength(data []byte) ([][]byte, int) {
	fields := bytes.Split(data, []byte{sepByte})
	return fields[:len(fields)-1], len(fields) - 1
}

func bytesToStrSlice(data []byte) []string {
	fields := bytes.Split(data, []byte{sepByte})
	slice := make([]string, len(fields))
	for i, field := range fields {
		slice[i] = string(field)
	}
	return slice
}

func bytesToTime(b []byte) (time.Time, error) {
	// format := "20060102 15:04:05 Mountain Standard Time"
	// 214 208 185 250 177 234 215 188 202 177 188 228
	fmt.Println("bytetotime", string(b))
	format := "20060102 15:04:05 MST"
	t := string(b)
	localtime, err := time.ParseInLocation(format, t, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return localtime, nil
}

func bytesUnixToTime(b []byte) (time.Time, error) {
	unix, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return time.Time{}, nil
	}
	return time.Unix(unix, 0), nil
}

func writeString(w *bytes.Buffer, str string) {
	_, err := w.Write([]byte(str))
	if err != nil {
		panic(err)
	}
	if err = w.WriteByte(sepByte); err != nil {
		panic(err)
	}
}

func writeInt(w *bytes.Buffer, i int64) {
	writeString(w, strconv.FormatInt(i, 10))
	// req.bytes = append(req.bytes, +"\x00")...)
}

func writeFloat(w *bytes.Buffer, f float64) {
	writeString(w, strconv.FormatFloat(f, 'g', 10, 64))
	// req.bytes = append(req.bytes, []byte(strconv.FormatFloat(f, 'g', 10, 64)+"\x00")...)
}

func writeBool(w *bytes.Buffer, b bool) {
	if b {
		writeString(w, "1")
	}
	writeString(w, "0")
}

// func writeContract(w *bytes.Buffer, c *Contract) {
// 	writeInt(w, c.ConID)
// 	writeString(w, c.Symbol)
// 	writeString(w, c.SecType)
// 	writeString(w, c.LastTradeDateOrContractMonth)
// 	writeFloat(w, c.Strike)
// 	writeString(w, c.Right)
// 	writeString(w, c.Multiplier)
// 	writeString(w, c.Exchange)
// 	writeString(w, c.PrimaryExchange)
// 	writeString(w, c.Currency)
// 	writeString(w, c.LocalSymbol)
// 	writeString(w, c.TradingClass)
// }
// func writeContractWithExpired(w *bytes.Buffer, c *Contract) {
// 	writeContract(w, c)
// 	writeBool(w, c.IncludeExpired)
// }

// func writeContractFull(w *bytes.Buffer, c *Contract) {
// 	writeContractWithExpired(w, c)
// 	writeString(w, c.SecIDType)
// 	writeString(w, c.SecID)
// }

func makeMsgBytes(fields ...interface{}) []byte {
	msgBytes := make([]byte, 4, 8*len(fields)+4) // pre alloc memory

	for _, f := range fields {
		switch v := f.(type) {
		case int64:
			msgBytes = strconv.AppendInt(msgBytes, v, 10)
		case float64:
			msgBytes = strconv.AppendFloat(msgBytes, v, 'g', 10, 64)
		case string:
			msgBytes = append(msgBytes, []byte(v)...)

		case bool:
			if v {
				msgBytes = append(msgBytes, '1')
			} else {
				msgBytes = append(msgBytes, '0')
			}
		case int:
			msgBytes = strconv.AppendInt(msgBytes, int64(v), 10)
		case []byte:
			msgBytes = append(msgBytes, v...)
		default:
			// log.Panic("failed to covert the field", zap.Reflect("field", f)) // never reach here
		}

		msgBytes = append(msgBytes, sepByte)
	}

	// add the size header
	binary.BigEndian.PutUint32(msgBytes, uint32(len(msgBytes)-4))
	return msgBytes
}

func scanFields(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// fmt.Println("scanning")
	// fmt.Println(len(data))
	if atEOF {
		return 0, nil, io.EOF
	}

	if len(data) < 4 {
		return 0, nil, nil
	}
	totalSize := int(binary.BigEndian.Uint32(data[:4])) + 4

	if totalSize > len(data) {
		return 0, nil, nil
	}

	// msgBytes := make([]byte, totalSize-4, totalSize-4)
	// copy(msgBytes, data[4:totalSize])
	// not copy here, copied by callee more reasonable
	return totalSize, data[4:totalSize], nil
}
