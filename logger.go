package gocore

import (
	"fmt"
	"log"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/mgutz/ansi"
)

// Debugf Comment
func (l *Logger) Debugf(msg string, args ...interface{}) {
	l.output("DEBUG", "blue", msg, args...)
}

// Infof comment
func (l *Logger) Infof(msg string, args ...interface{}) {
	l.output("INFO", "green", msg, args...)
}

// Warnf comment
func (l *Logger) Warnf(msg string, args ...interface{}) {
	l.output("WARN", "yellow", msg, args...)
}

// Errorf comment
func (l *Logger) Errorf(msg string, args ...interface{}) {
	args = append(args, getStack())
	msg = msg + "\n%s"
	l.output("ERROR", "red", msg, args...)
}

func getStack() string {
	return strings.Join(strings.Split(string(debug.Stack()), "\n")[7:], "\n")
}

func (l *Logger) output(level, colour, msg string, args ...interface{}) {
	print := true

	if level == "DEBUG" {
		match, _ := regexp.MatchString(l.conf.debug.regex, msg)
		if !l.conf.debug.enabled || !match {
			print = false
		}
	}

	if l.colour && colour != "" {
		level = ansi.Color(level, colour)
	}

	format := fmt.Sprintf("%s %s. %s: %s\n", l.packageName, l.serviceName, level, msg)

	if print {
		log.Printf(format, args...)
	}

	if len(l.conf.trace.sockets) > 0 {
		for s, r := range l.conf.trace.sockets {
			match, _ := regexp.MatchString(r, msg)
			if match {
				_, err := s.Write([]byte(fmt.Sprintf(format, args...)))
				if err != nil {
					log.Printf("Writing client error: %+v", err)
					delete(l.conf.trace.sockets, s)
				}
			}
		}
	}
}
