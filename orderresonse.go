package ibgo

type OrderStatus struct {
	Message
	serverVersion int64
	// TODO
}

func (resp *OrderStatus) id() string {
	resp.serverVersion = readInt(resp.buf)
	return readString(resp.buf)
}

func (resp *OrderStatus) read() {}

type OpenOrder struct {
	Message
	serverVersion int64
	version       int64
}
