package ibgo

// import (
// 	"bufio"
// 	"encoding/binary"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"net"
// 	"strconv"
// 	"strings"
// 	"time"
// )

// /*
// 	open
// 		goSocket -> handshake -> startAPI -> update Server info -> go goReader
// 		-> go goWriter
// 	gracefully close
// 		router close s.wc -> goWriter notify goReader -> goReader clean up; close s.rc
// 		 ->  router notified
// 	reconnect
// 		#1 goReader EOF -> goSocket return done -> router control
// 		TODO

// 		#2 writer error -> goWriter notify goReader -> goReader clean up
// 		 ->  goSocket return done -> router control

// 		TODO
// 		#3 reader error -> goRedaer notify goWriter -> goWriter close
// 		 -> goWriter notify goReader -> goReader clean up
// 		 ->  goSocket return done -> router control
// */
// type ibSocket struct {
// 	clientID     string
// 	address      string
// 	conn         net.Conn
// 	outgoingChan chan *Request
// 	incomingChan chan *Response
// 	scannerChan  chan *Response
// 	terminal     chan bool
// 	writer       *bufio.Writer
// 	scanner      *bufio.Scanner
// 	serverInfo   *handshakeInfo
// 	err          error
// }

// type socketDone struct {
// 	err       error
// 	everRun   bool
// 	startTime time.Time
// 	endTime   time.Time
// }
// type handshakeInfo struct {
// 	serverVersion int
// 	connTime      time.Time
// 	nextID        int64
// 	mgAccounts    []string
// }

// func (s *ibSocket) run() bool {
// 	fmt.Println("starting socket")
// 	connOK := make(chan bool)
// 	defer close(connOK)
// 	s.serverInfo = &handshakeInfo{}
// 	go func() {
// 		conn, err := net.Dial("tcp", s.address)
// 		if err != nil {
// 			s.setErr(err)
// 			connOK <- false
// 			return
// 		}
// 		defer conn.Close()
// 		s.conn = conn
// 		s.writer = bufio.NewWriter(s.conn)
// 		s.scanner = bufio.NewScanner(s.conn)
// 		s.scanner.Split(scanFields)
// 		if s.scannerChan == nil {
// 			s.scannerChan = make(chan *Response, 10)
// 		}
// 		scannerDone := s.runScanner()

// 		if err := s.sendHandShake(); err != nil {
// 			s.setErr(err)
// 			connOK <- false
// 			return
// 		}
// 		// conn.Close()
// 		for c := true; c; {
// 			select {
// 			case <-time.After(time.Second * 5):
// 				s.setErr(errors.New("Timeout during handshake"))
// 				connOK <- false
// 				return
// 			case r := <-s.scannerChan:
// 				// if ok {
// 				if sl, l := splitMsgBytes(r.bytes); l == 2 {
// 					v, _ := strconv.Atoi(string(sl[0]))
// 					t, _ := bytesToTime(sl[1])
// 					s.serverInfo.connTime, s.serverInfo.serverVersion = t, v
// 					c = false
// 				} else {
// 					s.incomingChan <- r
// 				}
// 			// } else {
// 			case <-scannerDone:
// 				// fmt.Println("before send", s.err)
// 				connOK <- false
// 				return
// 				// }
// 			}
// 		}
// 		if err := s.sendStartAPI(); err != nil {
// 			s.setErr(err)
// 			connOK <- false
// 			return
// 		}
// 		for i := 0; i != 3; {
// 			select {
// 			case <-time.After(time.Second * 5):
// 				s.setErr(errors.New("Timeout during startapi"))
// 				connOK <- false
// 				return
// 			case r := <-s.scannerChan:
// 				// if !ok {

// 				sl, _ := splitMsgBytes(r.bytes)
// 				code := string(sl[0])
// 				if code == inNEXTVALIDID {
// 					// fmt.Println("id got", string(sl[1]), string(sl[2]))
// 					i |= 1
// 					s.serverInfo.nextID, _ = strconv.ParseInt(string(sl[2]), 10, 64)
// 				} else if code == inMANAGEDACCTS {
// 					// fmt.Println("acct got", string(sl[1]), string(sl[2]))
// 					s.serverInfo.mgAccounts = strings.Split(string(sl[2]), ",")
// 					i |= 2
// 				} else {
// 					// fmt.Println("other got", string(sl[1]), string(sl[2]))
// 					s.incomingChan <- r
// 				}
// 			case <-scannerDone:
// 				connOK <- false
// 				return
// 			}
// 		}
// 		connOK <- true
// 		incomingDone := s.runIncoming()
// 		outgoingDone := s.runOutgoing()
// 		s.terminal = make(chan bool)
// 		for {
// 			select {
// 			case <-incomingDone:
// 				fmt.Println("incoming done")
// 			case <-outgoingDone:
// 				fmt.Println("outgoing done")
// 			case <-scannerDone:
// 				fmt.Println("scanner done")
// 				s.terminal <- true
// 				s.terminal <- true
// 				<-incomingDone
// 				fmt.Println("incoming done1")
// 				<-outgoingDone
// 				fmt.Println("outgoing done1")
// 			}
// 		}
// 	}()
// 	return <-connOK
// }

// func (s *ibSocket) setErr(err error) {
// 	if s.err == nil {
// 		s.err = err
// 	}
// }

// func (s *ibSocket) sendHandShake() error {
// 	head := []byte("API\x00")
// 	clientVersion := []byte(fmt.Sprintf("v%d..%d", 100, 151))
// 	sizeofCV := make([]byte, 4)
// 	binary.BigEndian.PutUint32(sizeofCV, uint32(len(clientVersion)))
// 	s.writer.Write(head)
// 	s.writer.Write(sizeofCV)
// 	s.writer.Write(clientVersion)
// 	if err := s.writer.Flush(); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (s *ibSocket) sendStartAPI() error {
// 	var startAPIBytes []byte
// 	const v = "2"
// 	startAPIBytes = makeMsgBytes(outSTARTAPI, v, s.clientID, "")
// 	if _, err := s.writer.Write(startAPIBytes); err != nil {
// 		return err
// 	}
// 	if err := s.writer.Flush(); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (s *ibSocket) runScanner() chan bool {
// 	done := make(chan bool)
// 	go func() {
// 		defer close(done)
// 		fmt.Println("Scanner started")
// 		for s.scanner.Scan() {
// 			resBytes := make([]byte, len(s.scanner.Bytes()))
// 			copy(resBytes, s.scanner.Bytes())
// 			s.scannerChan <- newResponseBytes(resBytes)
// 		}
// 		err := s.scanner.Err()
// 		if err == nil {
// 			err = io.EOF
// 		}
// 		s.setErr(err)
// 		fmt.Println("Scanner stopped", s.err)
// 	}()

// 	return done
// }
// func (s *ibSocket) runIncoming() chan bool {
// 	done := make(chan bool)
// 	go func() {
// 		defer close(done)
// 		fmt.Println("Incoming started")
// 		for {
// 			select {
// 			case r := <-s.scannerChan:
// 				s.incomingChan <- r
// 			case <-s.terminal:
// 				fmt.Println("terminal signal in")
// 				return
// 			}
// 		}
// 	}()
// 	return done
// }

// func (s *ibSocket) runOutgoing() chan bool {
// 	done := make(chan bool)
// 	go func() {
// 		defer close(done)
// 		fmt.Println("Outgoing started")
// 		for {
// 			select {
// 			case r := <-s.outgoingChan:
// 				req := r.bytes
// 				if _, err := s.writer.Write(req); err != nil {
// 					r.ack <- err
// 					s.setErr(err)
// 					break
// 				}
// 				if err := s.writer.Flush(); err != nil {
// 					r.ack <- err
// 					s.setErr(err)
// 					break
// 				}
// 				fmt.Println(req, "sent")
// 				close(r.ack)
// 			case <-s.terminal:
// 				fmt.Println("terminal signal out")
// 				return
// 			}

// 		}
// 	}()
// 	return done
// }
