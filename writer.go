package ibgo

import (
	"encoding/binary"
	"net"
	"strconv"
)

type RequestBody interface {
	write(c *reqWriter)
}
type reqWriter struct {
	wt            *net.TCPConn
	buf           []byte
	serverVersion int64
}

func newWriter(c *IBClient) (w *reqWriter) {
	return &reqWriter{c.conn, make([]byte, 1024), c.serverVersion}
}
func (w *reqWriter) writeString(s string) {
	w.buf = append(w.buf, s...)
	w.buf = append(w.buf, 0)
}
func (w *reqWriter) writeInt(i int64) {
	w.buf = strconv.AppendInt(w.buf, i, 10)
	w.buf = append(w.buf, 0)
}
func (w *reqWriter) writeFloat(f float64) {
	w.buf = strconv.AppendFloat(w.buf, f, 'g', 10, 64)
	w.buf = append(w.buf, 0)
}
func (w *reqWriter) writeBool(b bool) {
	if b {
		w.writeString("1")
	} else {
		w.writeString("0")
	}
}
func (w *reqWriter) writeContract(c *Contract) {
	w.writeInt(c.ConID)
	w.writeString(c.Symbol)
	w.writeString(c.SecType)
	w.writeString(c.LastTradeDateOrContractMonth)
	w.writeFloat(c.Strike)
	w.writeString(c.Right)
	w.writeString(c.Multiplier)
	w.writeString(c.Exchange)
	w.writeString(c.PrimaryExchange)
	w.writeString(c.Currency)
	w.writeString(c.LocalSymbol)
	w.writeString(c.TradingClass)
}
func (w *reqWriter) writeContractWithExpired(c *Contract) {
	w.writeContract(c)
	w.writeBool(c.IncludeExpired)
}
func (w *reqWriter) writeContractWithFull(c *Contract) {
	w.writeContractWithExpired(c)
	w.writeString(c.SecIDType)
	w.writeString(c.SecID)
}
func (w *reqWriter) send() (err error) {
	defer func() { w.buf = w.buf[:0] }()
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(w.buf)))
	_, err = w.wt.Write(size)
	if err != nil {
		return
	}
	_, err = w.wt.Write(w.buf)
	return

}
