package gocore

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func init() {
	socketDIR, _ := Config().Get("socketDIR")
	err := os.MkdirAll(socketDIR, os.ModePerm)
	if err != nil {
		log.Printf("ERROR: Unable to make socket directory %s: %+v", socketDIR, err)
	}
}

// NewLogger comment
func NewLogger(packageName string, serviceName string, enableColours bool) *Logger {
	logger := Logger{
		packageName: strings.ToUpper(packageName),
		serviceName: strings.ToUpper(serviceName),
		colour:      enableColours,
		conf: loggerConfig{
			mu: new(sync.RWMutex),
			trace: traceSettings{
				sockets: make(map[net.Conn]string),
			},
		},
	}

	// Run a listener on a Unix socket
	go func() {
		n := fmt.Sprintf(
			"/tmp/sockets/%s.%s%d.sock",
			strings.ToUpper(packageName), strings.ToUpper(serviceName), getRand(),
		)

		ln, err := net.Listen("unix", n)
		if err != nil {
			log.Fatalf("LOGGER: listen error: %+v", err)
		}

		log.Printf("Socket created. Connect with 'nc -U %s'", n)

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

		logger.handleShutdown(ln, ch)

		for {
			fd, err := ln.Accept()
			if err != nil {
				log.Fatalf("LOGGER: Accept error: %+v", err)
			}

			logger.handleIncomingMessage(fd)
		}

	}()

	return &logger
}

func getRand() int {
	rand.Seed(time.Now().UnixNano())
	min := 100000000
	max := 999999999
	return rand.Intn(max-min) + min
}
