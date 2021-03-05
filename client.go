package ibgo

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

const ReconnectIntervalSeconds = 5
const RequestsPerSecond = 50
const MinOrderID = int64(1 << 10)
const MinTickerID = int64(1 << 4)

// ConnectionState
const (
	DISCONNECTED = iota
	CONNECTING
	OFFLINE
	CONNECTED
)

var ReconnectInterval = time.Second * time.Duration(ReconnectIntervalSeconds)

type IBClient struct {
	Address        string
	ClientID       int64
	readonly       bool
	status         int
	registry       map[string]chan *Message
	conn           *net.TCPConn
	reader         *msgReader
	writer         *reqWriter
	systemChan     chan *Message
	unRegisterChan chan string
	runningErr     []error
	stop           chan struct{}
	done           chan error
	terminate      chan struct{}
	nextID         chan int64
	ticker         chan *quote
	order          chan *quote
	static         chan *quote
	historical     chan bool
	handshakeInfo
}

type handshakeInfo struct {
	serverVersion int64
	connTime      time.Time
	mgAccounts    []string
}

type quote struct {
	id  int64
	ack chan struct{}
}

type cancelRequest struct {
	id  string
	ack chan struct{}
}

func NewClient(addr string, id int64, allowTrade bool) (*IBClient, error) {
	c := &IBClient{
		Address:  addr,
		ClientID: id,
		readonly: !allowTrade,
	}
	if err := c.Connect(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *IBClient) Connect() error {
	raddr, err := net.ResolveTCPAddr("tcp", c.Address)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return err
	}
	c.status = CONNECTING
	c.conn = conn
	c.reader = newReader(c.conn)
	c.writer = newWriter(c)
	c.nextID = make(chan int64, 1)
	c.runningErr = []error{}
	handshakeDone := make(chan error)
	defer close(handshakeDone)
	go c.handshake(handshakeDone)
	select {
	case err := <-handshakeDone:
		if err != nil {
			return err
		}
	case <-time.After(time.Second * 2):
		return errTimeOut
	}
	c.status = CONNECTED
	fmt.Println(c.handshakeInfo)
	c.systemChan = make(chan *Message)
	c.registry = make(map[string]chan *Message)
	c.unRegisterChan = make(chan string)
	c.ticker = make(chan *quote)
	c.order = make(chan *quote)
	c.static = make(chan *quote)
	c.historical = make(chan bool)
	// c.registry["-1"] = c.systemChan
	go c.systemInfo()
	go c.limiter()
	go c.historicalLimiter()
	c.done, _ = goWithDone(c.mainLoop)
	return nil
}

func (c *IBClient) KeepAlive() {
	go func() {
		for {
			fmt.Println("Client stopped with error", <-c.Done())
			for {
				time.Sleep(ReconnectInterval)
				if err := c.Connect(); err == nil {
					break
				}
			}
		}
	}()
}

func (c *IBClient) Terminate() error {
	if c.terminate == nil {
		return errors.New("Client is not running")
	}
	close(c.terminate)
	return nil
}

func (c *IBClient) Done() chan error {
	return c.done
}

func (c *IBClient) mainLoop(done chan error, cancel chan struct{}) {
	c.terminate = make(chan struct{})
	receiverDone, _ := goWithDone(c.receiver)
	select {
	case <-c.terminate:
		c.conn.Close()
		fmt.Println("receiver closed", <-receiverDone)
		done <- nil
	case err := <-receiverDone:
		c.runningErr = append(c.runningErr, err)
		fmt.Println("Client fail due to receiver error", err)
		c.conn.Close()
		done <- err
	}

}

func (c *IBClient) readSingleMsg() (b []byte, err error) {
	l := make([]byte, 4)
	if _, err = io.ReadFull(c.conn, l); err != nil {
		return
	}
	size := int(binary.BigEndian.Uint32(l))
	b = make([]byte, size)
	if _, err = io.ReadFull(c.conn, b); err != nil {
		return
	}
	return
}
func (c *IBClient) receiver(done chan error, cancel chan struct{}) {
	msgCh := make(chan *Message, 10)
	readerDone, _ := goWithDone(func(done chan error, cancel chan struct{}) {
		for {
			msg, err := c.reader.readMessage()
			if err != nil {
				done <- err
			}
			if msg != nil {
				msgCh <- msg
			}
		}
	})
	for {
		select {
		case err := <-readerDone:
			fmt.Println("Reader stopped with error", err)
			done <- err
			return
		case cancelID := <-c.unRegisterChan:
			idChan := c.registry[cancelID]
			delete(c.registry, cancelID)
			close(idChan)
		case msg := <-msgCh:
			respCh, ok := c.registry[msg.id]
			if ok {
				// fmt.Println("received", msg.id, respCh)
				respCh <- msg
			} else {
				fmt.Println("not registered", msg.code, msg.id, msg.body)
			}
			// printMessage(msg)
			// msg.open(c.handshakeInfo.serverVersion)
			// if msg.ID != "IGNORE" {
			// 	idChan, ok := c.registry[msg.ID]
			// 	if !ok {
			// 		c.systemChan <- msg
			// 		msg.Error = fmt.Errorf("ID %s isn't registerd", msg.ID)
			// 	} else if idChan == nil {
			// 		msg.Error = fmt.Errorf("Listener of ID %s has quit", msg.ID)
			// 		c.systemChan <- msg
			// 	} else {
			// 		idChan <- msg
			// 	}
			// }
		}
	}
}

// func (c *IBClient) REQ(request Request) (msgChan chan *Message, cancelReq func() error, err error) {
// 	var q *quote
// 	var id string
// 	var ack chan struct{}
// 	// acquire request quote
// 	rType := request.typ()
// 	switch rType.kind {
// 	case "STATIC":
// 		return
// 	case "TICKER":
// 		q = <-c.ticker
// 		id, ack = strconv.FormatInt(q.id, 10), q.ack
// 	case "ORDER":
// 		if c.readonly {
// 			return nil, nil, ErrReadonly
// 		}
// 		q = <-c.order
// 		id, ack = strconv.FormatInt(q.id, 10), q.ack
// 	case "HISTORICAL":
// 		//TODO add additional limiter
// 		q = <-c.ticker
// 		id, ack = strconv.FormatInt(q.id, 10), q.ack
// 		return
// 	default:
// 		return nil, nil, fmt.Errorf("Bad Request Type %v", request.typ())
// 	}
// 	// expect 0 response msg
// 	if rType.expectMsgCnt == 0 {
// 		if err := c.sendReq(request, id); err != nil {
// 			close(ack)
// 			return nil, nil, ErrIOError
// 		}
// 		return
// 	}
// 	// expect response msg
// 	if _, ok := c.registry[id]; ok {
// 		close(ack)
// 		// could this happen somehow?
// 		return nil, nil, ErrDuplicateReqID
// 	}
// 	// ready to send
// 	msgChan = make(chan *Message)
// 	idChan := make(chan *Message)
// 	c.registry[id] = idChan
// 	if err := c.sendReq(request, id); err != nil {
// 		close(ack)
// 		return nil, nil, ErrIOError
// 	}
// 	close(ack)
// 	var cancel chan chan error
// 	if rType.cancelable {
// 		cancel = make(chan chan error)
// 		cancelReq = func() error {
// 			ech := make(chan error)
// 			fmt.Println("cancelling", cancel)
// 			cancel <- ech
// 			fmt.Println("cancelling")
// 			return <-ech
// 		}
// 	} else {
// 		cancelReq = func() error {
// 			return fmt.Errorf("Request %s is not cancelable", id)
// 		}
// 	}
// 	// succesfully sent ready to return
// 	// request lifecicle loop
// 	go func() {
// 		// unregister id close msgChan & idChan
// 		defer func() {
// 			close(msgChan)
// 			c.unRegisterChan <- id
// 			for _, ok := <-idChan; ok; {
// 			}
// 		}()
// 		var first *Message
// 		// waiting for first msg to handle timeout aka req fail silently
// 		select {
// 		//TOCHECK some historical data request might take long time to response what about other situation?
// 		case <-time.After(time.Second * 5):
// 			msgChan <- MsgTimeoutErr
// 			return
// 		case first = <-idChan:
// 		}
// 		// expect 1 msg
// 		if rType.expectMsgCnt == 1 {
// 			if first.Code == inERRMSG {
// 				e := &ErrMsg{}
// 				e.read(first)
// 				first.Error = e.toError()
// 			}
// 			msgChan <- first
// 			return
// 		}
// 		pending := make([]*Message, 0)
// 		pending = append(pending, first)
// 		var update chan *Message
// 		fmt.Println("cancelling channel", cancel)
// 		for {
// 			if len(pending) > 0 {
// 				first = pending[0]
// 				if first.Code == inERRMSG {
// 					e := &ErrMsg{}
// 					e.read(first)
// 					first.Error = e.toError()
// 					msgChan <- first
// 					return
// 				}
// 				// multi parts response end here
// 				if rType.expectMsgCnt == 2 && rType.endCode == first.Code {
// 					msgChan <- first
// 					return
// 				}
// 				update = msgChan
// 			}
// 			select {
// 			case ech := <-cancel:
// 				if _, ok := c.registry[id]; !ok {
// 					ech <- fmt.Errorf("Request %s has already finished", id)
// 					return
// 				}
// 				fmt.Println("canceling signal received")
// 				if req, ok := request.(CancelRequest); ok {
// 					cq := <-c.static
// 					ech <- c.sendCancel(req, id)
// 					close(cq.ack)
// 					return
// 				} else {
// 					log.Fatal("Failed to cancel", id)
// 				}
// 			case newMsg := <-idChan:
// 				if newMsg.Error != nil {
// 					idChan = nil
// 				}
// 				pending = append(pending, newMsg)
// 			case update <- first:
// 				update = nil
// 				pending = pending[1:]
// 			}
// 		}
// 	}()
// 	return
// }

func (c *IBClient) limiter() {
	nextOrder := &quote{MinOrderID, make(chan struct{})}
	nextTicker := &quote{MinTickerID, make(chan struct{})}
	nextStatic := &quote{0, make(chan struct{})}
	limit := make(chan bool, RequestsPerSecond)
	for i := 0; i < RequestsPerSecond; i++ {
		limit <- true
	}
	for {
		<-limit
		select {
		// nextID update event will slow down the limiter by 1 quote which is fine
		case id := <-c.nextID:
			if id > nextOrder.id {
				nextOrder.id = id
			}
		case c.static <- nextStatic:
			ack := nextStatic.ack
			<-ack
			nextStatic = &quote{0, make(chan struct{})}
			time.AfterFunc(time.Second, func() { limit <- true })
		case c.ticker <- nextTicker:
			id, ack := nextTicker.id, nextTicker.ack
			<-ack
			id++
			if id == MinOrderID {
				id = MinTickerID
			}
			nextTicker = &quote{id, make(chan struct{})}
			time.AfterFunc(time.Second, func() { limit <- true })
		case c.order <- nextOrder:
			id, ack := nextOrder.id, nextOrder.ack
			<-ack
			id++
			nextOrder = &quote{id, make(chan struct{})}
			time.AfterFunc(time.Second, func() { limit <- true })
		}
	}
}

func (c *IBClient) historicalLimiter() {
	limit := make(chan bool, 60)
	for i := 0; i < 60; i++ {
		limit <- true
	}
	for {
		<-limit
		c.historical <- true
		time.AfterFunc(time.Minute*10, func() { limit <- true })
	}
}

func (c *IBClient) handshake(done chan error) {
	if err := c.sendHandShake(); err != nil {
		done <- nil
		return
	}
	fmt.Println("handshake sent")
handshake:
	for {
		bytesMsg, err := c.reader.readFull()
		if err != nil {
			done <- err
			return
		}
		if fields := bytes.Split(bytesMsg, []byte{0}); len(fields) == 3 {
			v, _ := strconv.ParseInt(string(fields[0]), 10, 64)
			t, _ := bytesToTime(fields[1])
			c.connTime, c.serverVersion = t, v
			c.reader.serverVersion = v
			c.writer.serverVersion = v
			break handshake
		}
	}
	if err := c.startAPI(); err != nil {
		done <- err
		return
	}
	fmt.Println("start api sent")
	for i := 0; i != 3; {
		msg, err := c.reader.readMessage()
		if err != nil {
			done <- err
		}
		switch msg.code {
		case inNEXTVALIDID:
			i |= 1
			c.nextID <- (msg.body).(*NextValidID).ID
		case inMANAGEDACCTS:
			i |= 2
			c.mgAccounts = (msg.body).(*ManagedAccounts).Accounts
		default:
		}
	}
	fmt.Println("startapi success")
	done <- nil
}

func (c *IBClient) startAPI() error {
	c.writer.writeString(outSTARTAPI)
	c.writer.writeString("2")
	c.writer.writeInt(c.ClientID)
	c.writer.writeString("")
	return c.writer.send()

}
func (c *IBClient) sendSingleMsg(msg []byte) (err error) {
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(msg)))
	_, err = c.conn.Write(size)
	if err != nil {
		return
	}
	_, err = c.conn.Write(msg)
	return
}
func (c *IBClient) sendHandShake() (err error) {
	head := []byte("API\x00")
	clientVersion := []byte(fmt.Sprintf("v%d..%d", 100, 151))
	_, err = c.conn.Write(head)
	if err != nil {
		return
	}
	err = c.sendSingleMsg(clientVersion)
	return
}

func (c *IBClient) sendStartAPI() (err error) {
	var startAPIBytes []byte
	const v = "2"
	startAPIBytes = makeMsgBytes(outSTARTAPI, v, c.ClientID, "")
	_, err = c.conn.Write(startAPIBytes)
	return
}

func (c *IBClient) systemInfo() {
	// for resp := range c.systemChan {
	// }
}
