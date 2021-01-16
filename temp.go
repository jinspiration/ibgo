package ibgo

import (
	"fmt"
	"strings"
	"time"
)

func ReadAll(respChan chan *Message) {
	for resp := range respChan {
		Printmsg(resp)
	}
	fmt.Println("response ended")
}
func PrintAll(respChan chan *Message) {
	// var sum, avg int
	var resp *Message
	// resp = <-respChan
	for i := 1; ; i++ {
		resp = <-respChan
		// sum += int(time.Since(resp.Time))
		// avg = sum / i
		// fmt.Println(time.Duration(avg))
		fmt.Println(time.Since(resp.Time))
	}
}
func Printmsg(r *Message) {
	sl, _ := splitMsgBytesWithLength(r.bytes)
	ss := make([]string, len(sl))
	for i, b := range sl {
		ss[i] = string(b)
	}
	fmt.Println(r.Time.Format("2006 Jan 2 15:04:05"), strings.Join(ss, " "))
}
