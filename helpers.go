package ibgo

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"
)

// type routineFunc = func(done chan error, args ...interface{})

var errTimeOut = errors.New("Routine Timeout")
var errUnexpected = errors.New("Unexpected error")

func goWithDone(fn func(chan error, chan struct{})) (chan error, chan struct{}) {
	done := make(chan error)
	cancel := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Exception caught: ", r)
				fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))

				done <- errUnexpected
			}
		}()
		fn(done, cancel)
	}()
	return done, cancel
}

func goWithTimeout(fn func(chan error), timeout int) chan error {
	done := make(chan error)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Exception caught: ", r)
				done <- errUnexpected
			}
		}()
		fn(done)
	}()
	time.AfterFunc(time.Second*time.Duration(timeout), func() { safeSendErr(done, errTimeOut) })
	return done
	// innerCtx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
	// defer cancel()
}

func safeSend(ch chan interface{}, value interface{}) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = true
		}
	}()

	ch <- value  // panic if ch is closed
	return false // <=> closed = false; return
}

func safeSendErr(ch chan error, err error) {
	defer func() {
		recover()
	}()
	// fmt.Println("trying to send err", err)
	ch <- err // panic if ch is closed
	fmt.Println("err sent", err)
}
