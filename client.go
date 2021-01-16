package ibgo

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const ReconnectIntervalSeconds = 5
const RequestsPerSecond = 50
const MinOrderID = int64(1 << 10)
const MinTickerID = int64(1 << 4)

var ReconnectInterval = time.Second * time.Duration(ReconnectIntervalSeconds)

type IBClient struct {
	Address        string
	ClientID       int
	readonly       bool
	registry       map[string]chan *Message
	conn           net.Conn
	outBuffer      *bytes.Buffer
	writer         *bufio.Writer
	scanner        *bufio.Scanner
	handshakeInfo  *handshakeInfo
	pending        []Message
	systemChan     chan *Message
	unRegisterChan chan string
	runningErr     []error
	stop           chan struct{}
	done           chan error
	terminate      chan struct{}
	ticker         chan *quote
	order          chan *quote
	static         chan *quote
}

type handshakeInfo struct {
	serverVersion int64
	connTime      time.Time
	nextID        int64
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

func NewClient(addr string, id int, allowTrade bool) (*IBClient, error) {
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
	conn, err := net.Dial("tcp", c.Address)
	if err != nil {
		return err
	}
	c.conn = conn
	c.outBuffer = bytes.NewBuffer(make([]byte, 1024))
	c.writer = bufio.NewWriter(c.conn)
	c.scanner = bufio.NewScanner(c.conn)
	c.scanner.Split(scanFields)
	c.handshakeInfo = &handshakeInfo{}
	c.runningErr = []error{}
	c.pending = []Message{}
	handshakeDone := goWithTimeout(c.handshake, 2)
	defer close(handshakeDone)
	if err := <-handshakeDone; err != nil {
		return err
	}
	fmt.Println(c.handshakeInfo)
	c.systemChan = make(chan *Message)
	c.registry = make(map[string]chan *Message)
	c.unRegisterChan = make(chan string)
	c.ticker = make(chan *quote)
	c.order = make(chan *quote)
	c.static = make(chan *quote)
	c.registry["-1"] = c.systemChan
	go c.systemInfo()
	go c.limiter(c.handshakeInfo.nextID)
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

func (c *IBClient) receiver(done chan error, cancel chan struct{}) {
	msgChan := make(chan *Message, 10)
	readerDone, _ := goWithDone(func(done chan error, cancel chan struct{}) {
		fmt.Println("Reader started")
		for c.scanner.Scan() {
			msgChan <- newMessage(c.scanner.Bytes())
		}
		err := c.scanner.Err()
		// if err == nil {
		// 	err = io.EOF
		// }
		done <- err
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
		case msg := <-msgChan:
			// printMessage(msg)
			msg.open(c.handshakeInfo.serverVersion)
			if msg.ID != "IGNORE" {
				idChan, ok := c.registry[msg.ID]
				if !ok {
					c.systemChan <- msg
					msg.Error = fmt.Errorf("ID %s isn't registerd", msg.ID)
				} else if idChan == nil {
					msg.Error = fmt.Errorf("Listener of ID %s has quit", msg.ID)
					c.systemChan <- msg
				} else {
					idChan <- msg
				}
			}
		}
	}
}

func (c *IBClient) REQ(request Request) (msgChan chan *Message, cancelReq func() error, err error) {
	var q *quote
	var id string
	var ack chan struct{}
	// acquire request quote
	rType := request.typ()
	switch rType.kind {
	case "STATIC":
		return
	case "TICKER":
		q = <-c.ticker
		id, ack = strconv.FormatInt(q.id, 10), q.ack
	case "ORDER":
		if c.readonly {
			return nil, nil, ErrReadonly
		}
		q = <-c.order
		id, ack = strconv.FormatInt(q.id, 10), q.ack
	case "HISTORICAL":
		//TODO add additional limiter
		q = <-c.ticker
		id, ack = strconv.FormatInt(q.id, 10), q.ack
		return
	default:
		return nil, nil, fmt.Errorf("Bad Request Type %v", request.typ())
	}
	// expect 0 response msg
	if rType.expectMsgCnt == 0 {
		if err := c.sendReq(request, id); err != nil {
			close(ack)
			return nil, nil, ErrIOError
		}
		return
	}
	// expect response msg
	if _, ok := c.registry[id]; ok {
		close(ack)
		// could this happen somehow?
		return nil, nil, ErrDuplicateReqID
	}
	// ready to send
	msgChan = make(chan *Message)
	idChan := make(chan *Message)
	c.registry[id] = idChan
	if err := c.sendReq(request, id); err != nil {
		close(ack)
		return nil, nil, ErrIOError
	}
	close(ack)
	var cancel chan chan error
	if rType.cancelable {
		cancel = make(chan chan error)
		cancelReq = func() error {
			ech := make(chan error)
			fmt.Println("cancelling", cancel)
			cancel <- ech
			fmt.Println("cancelling")
			return <-ech
		}
	} else {
		cancelReq = func() error {
			return fmt.Errorf("Request %s is not cancelable", id)
		}
	}
	// succesfully sent ready to return
	// request lifecicle loop
	go func() {
		// unregister id close msgChan & idChan
		defer func() {
			close(msgChan)
			c.unRegisterChan <- id
			for _, ok := <-idChan; ok; {
			}
		}()
		var first *Message
		// waiting for first msg to handle timeout aka req fail silently
		select {
		//TOCHECK some historical data request might take long time to response what about other situation?
		case <-time.After(time.Second * 5):
			msgChan <- MsgTimeoutErr
			return
		case first = <-idChan:
		}
		// expect 1 msg
		if rType.expectMsgCnt == 1 {
			if first.Code == inERRMSG {
				e := &ErrMsg{}
				e.read(first)
				first.Error = e.toError()
			}
			msgChan <- first
			return
		}
		pending := make([]*Message, 0)
		pending = append(pending, first)
		var update chan *Message
		fmt.Println("cancelling channel", cancel)
		for {
			if len(pending) > 0 {
				first = pending[0]
				if first.Code == inERRMSG {
					e := &ErrMsg{}
					e.read(first)
					first.Error = e.toError()
					msgChan <- first
					return
				}
				// multi parts response end here
				if rType.expectMsgCnt == 2 && rType.endCode == first.Code {
					msgChan <- first
					return
				}
				update = msgChan
			}
			select {
			case ech := <-cancel:
				if _, ok := c.registry[id]; !ok {
					ech <- fmt.Errorf("Request %s has already finished", id)
					return
				}
				fmt.Println("canceling signal received")
				if req, ok := request.(CancelRequest); ok {
					cq := <-c.static
					ech <- c.sendCancel(req, id)
					close(cq.ack)
					return
				} else {
					log.Fatal("Failed to cancel", id)
				}
			case newMsg := <-idChan:
				if newMsg.Error != nil {
					idChan = nil
				}
				pending = append(pending, newMsg)
			case update <- first:
				update = nil
				pending = pending[1:]
			}
		}
	}()
	return
}

func (c *IBClient) limiter(serverMinOrderID int64) {
	nextOrder := &quote{MinOrderID, make(chan struct{})}
	nextTicker := &quote{MinTickerID, make(chan struct{})}
	nextStatic := &quote{0, make(chan struct{})}
	if serverMinOrderID > 2<<10 {
		nextOrder.id = serverMinOrderID
	}
	limit := make(chan bool, RequestsPerSecond)
	for i := 0; i < RequestsPerSecond; i++ {
		limit <- true
	}
	for {
		select {
		case <-limit:
			select {
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
}

func (c IBClient) sendReq(req Request, id string) error {
	c.outBuffer.Reset()
	req.writeBody(c.outBuffer, id, c.handshakeInfo.serverVersion)
	if err := binary.Write(c.writer, binary.BigEndian, int32(c.outBuffer.Len())); err != nil {
		return err
	}
	if _, err := c.writer.Write(c.outBuffer.Bytes()); err != nil {
		return err
	}
	if err := c.writer.Flush(); err != nil {
		return err
	}
	return nil
}

func (c IBClient) sendCancel(req CancelRequest, id string) error {
	c.outBuffer.Reset()
	req.writeCancel(c.outBuffer, id)
	if err := binary.Write(c.writer, binary.BigEndian, int32(c.outBuffer.Len())); err != nil {
		return err
	}
	if _, err := c.writer.Write(c.outBuffer.Bytes()); err != nil {
		return err
	}
	if err := c.writer.Flush(); err != nil {
		return err
	}
	return nil
}

func (c *IBClient) handshake(done chan error) {
	t := time.Now()
	if err := c.sendHandShake(); err != nil {
		safeSendErr(done, err)
		return
	}
	fmt.Println("handshake sent")
handshake:
	for {
		if c.scanner.Scan() {
			rb := newMessage(c.scanner.Bytes())
			if sl, l := splitMsgBytesWithLength(rb.bytes); l == 2 {
				fmt.Println(time.Since(t))
				v, _ := strconv.ParseInt(string(sl[0]), 10, 64)
				t, _ := bytesToTime(sl[1])
				c.handshakeInfo.connTime, c.handshakeInfo.serverVersion = t, v
				break handshake
			}
			c.pending = append(c.pending, *rb)
		}
		err := c.scanner.Err()
		if err == nil {
			err = io.EOF
		}
		safeSendErr(done, err)
	}
	if err := c.sendStartAPI(); err != nil {
		safeSendErr(done, err)
	}
	fmt.Println("start api sent")
	for i := 0; i != 3; {
		if c.scanner.Scan() {
			rb := newMessage(c.scanner.Bytes())
			sl, _ := splitMsgBytesWithLength(rb.bytes)
			code := string(sl[0])
			if code == inNEXTVALIDID {
				i |= 1
				c.handshakeInfo.nextID, _ = strconv.ParseInt(string(sl[2]), 10, 64)
			} else if code == inMANAGEDACCTS {
				c.handshakeInfo.mgAccounts = strings.Split(string(sl[2]), ",")
				i |= 2
			} else {
				c.pending = append(c.pending, *rb)
			}
		} else {
			err := c.scanner.Err()
			if err == nil {
				err = io.EOF
			}
			safeSendErr(done, err)
		}
	}
	safeSendErr(done, nil)
}

func (c *IBClient) sendHandShake() error {
	head := []byte("API\x00")
	clientVersion := []byte(fmt.Sprintf("v%d..%d", 100, 151))
	sizeofCV := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeofCV, uint32(len(clientVersion)))
	c.writer.Write(head)
	c.writer.Write(sizeofCV)
	c.writer.Write(clientVersion)
	if err := c.writer.Flush(); err != nil {
		return err
	}
	return nil
}

func (c *IBClient) sendStartAPI() error {
	var startAPIBytes []byte
	const v = "2"
	startAPIBytes = makeMsgBytes(outSTARTAPI, v, c.ClientID, "")
	if _, err := c.writer.Write(startAPIBytes); err != nil {
		return err
	}
	if err := c.writer.Flush(); err != nil {
		return err
	}
	return nil
}

func (c *IBClient) systemInfo() {
	for resp := range c.systemChan {
		Printmsg(resp)
	}
}
