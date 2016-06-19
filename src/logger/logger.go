package logger

import (
	"os"
	"bufio"
	"sync"
)

type Logger interface {
	Log(msg string)	
}


type logger struct {
	buf *bufio.Writer
	wMutex *sync.Mutex
}

var instance *logger
var mutex sync.Mutex



func GetLogger() Logger {

	defer mutex.Unlock()
	mutex.Lock()

	if instance == nil {
		instance = new(logger)
		f, err := os.Create("./raft_log.txt")
		
		if err != nil {
			panic(err)
		}

		instance.buf = bufio.NewWriter(f)
		instance.wMutex = &sync.Mutex{}
	}
	
	return instance
}

func (l *logger) Log(msg string) {
	defer l.wMutex.Unlock()
	l.wMutex.Lock()
	l.buf.WriteString(msg)
	l.buf.Flush()
}

