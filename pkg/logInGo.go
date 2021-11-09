package LogInGo

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

/*
 * 	Lig is a type that defines a logger
 * 	There are two ways to write logs using LIG :-
 * 	* Syncronously - via the Write Method
 * 	* Asyncronously - via the channel returned by the MessageChannel accessor
 */

type Lig struct {
	dest               io.Writer
	m                  *sync.Mutex
	msgCh              chan string
	errorCh            chan error
	shutdownCh         chan struct{}
	shutdownCompleteCh chan struct{}
}

/*
 *	New returns an Lig object with dest set to w
 *	In the event that w provided is nil, the default "os.Stdout" is used as the writer
 *  Message and error channels are initialized with empty buffers and can be retrieved
 *  using their respective accessor methods
 */

func New(w io.Writer, msgBufferSize int, errorBufferSize int) *Lig {
	if w == nil {
		w = os.Stdout
	}
	if msgBufferSize == 0 {
		msgBufferSize = 1
	}
	if errorBufferSize == 0 {
		errorBufferSize = 1
	}

	return &Lig{
		dest:               w,
		m:                  &sync.Mutex{},
		msgCh:              make(chan string, msgBufferSize),
		shutdownCh:         make(chan struct{}),
		shutdownCompleteCh: make(chan struct{}),
		errorCh:            make(chan error, errorBufferSize),
	}
}

func (al Lig) formatMessage(msg string) string {
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	return fmt.Sprintf("[%v] - %v", time.Now().Format("2006-01-02 15:04:05"), msg)
}

/*
 *
 * Start begins the message loop for the asynchronous logger
 * It should be initiated as a goroutine to prevent the caller from being blocked
 *
 */

func (al Lig) Start() struct{} {
	wg := &sync.WaitGroup{}
	for {
		select {
		case msg := <-al.msgCh:
			wg.Add(1)
			go al.write(msg, wg)
		case <-al.shutdownCh:
			wg.Wait()
			// al.shutdown()	
		}
	}
}

/*
 *  MessageChannel returns a channel that accepts messages that should be written to the log.
 */
func (al Lig) MessageChannel() chan string {
	return al.msgCh
}

/*
 * 	ErrorChannel returns a channel that will be populated when an error is raised during a write operation.
 * 	This channel should always be monitored in some way to prevent deadlock goroutines from being generated
 * 	when errors occur.
 */

func (al Lig) ErrorChannel() chan error {
	return al.errorCh
}

/*
 * 	Write writes curent message to dest iowriter object.
 *	Since, this is run as a go-routine, thread safety is ensured 
 *	using mutex m in the Lig object.
 */

func (al Lig) write(msg string, wg *sync.WaitGroup) {
	defer wg.Done()
	al.m.Lock()
	defer al.m.Unlock()
	_, err := al.dest.Write([]byte(al.formatMessage(msg)))
	if err != nil {
		go func(err error) {
			al.errorCh <- err
		}(err)
	}
}