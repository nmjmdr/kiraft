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
}

var instance *logger
var mutex sync.Mutex


func GetLogger() Logger {

	defer mutex.Unlock()
	mutex.Lock()

	if instance == nil {
		instance = new(logger)
		f, err := os.Create("./raft_log1.txt")
		
		if err != nil {
			panic(err)
		}

		instance.buf = bufio.NewWriter(f)		
	}
	
	return instance
}

func (l *logger) Log(msg string) {
	defer mutex.Unlock()
	mutex.Lock()
	l.buf.WriteString(msg)
	l.buf.Flush()
}

