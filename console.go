package xlog

import (
	"io"
	"os"
	"fmt"

	"github.com/go-logfmt/logfmt"
	"github.com/fatih/color"
)

const CONSOLE MODE = "console"

var consoleColors = []func(a ...interface{}) string{
	color.New(color.FgBlue).SprintFunc(),   // Trace
	color.New(color.FgGreen).SprintFunc(),  // Info
	color.New(color.FgYellow).SprintFunc(), // Warn
	color.New(color.FgRed).SprintFunc(),    // Error
	color.New(color.FgHiRed).SprintFunc(),  // Fatal
}

type console struct {
	Adapter // needs to compose the adapter
	writer        io.Writer
	valuesDefault []interface{}
}

type ConsoleConfig struct {
	Level         LEVEL
	BufferSize    int64
	Writer        io.Writer
	ValuesDefault []interface{}
}

func newConsole() Logger {
	return &console{
		valuesDefault: make([]interface{}, 0),
		Adapter: Adapter{
			quitChan: make(chan struct{}),
		},
	}
}

func init() {
	Register(CONSOLE, newConsole)
}

func (c *console) Level() LEVEL { return c.level }

func (c *console) Init(v interface{}) error {
	cfg, ok := v.(ConsoleConfig)
	if !ok {
		return ErrConfigObject{"ConsoleConfig", v}
	}

	if cfg.ValuesDefault != nil {
		c.valuesDefault = cfg.ValuesDefault
	}

	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}

	c.writer = cfg.Writer
	c.level = cfg.Level
	c.msgChan = make(chan *Message, cfg.BufferSize)

	return nil
}

func (c *console) write(msg *Message) {
	msg.Body = append(c.valuesDefault, msg.Body...)
	bs, err := logfmt.MarshalKeyvals(msg.Body...)
	if err != nil {
		c.errorChan <- err
		return
	}

	body := consoleColors[msg.Level](formats[msg.Level], string(bs))
	fmt.Fprintln(c.writer, body)
}

func (c *console) ExchangeChans(chan<- error) chan *Message {
	c.errorChan = errorChan
	return c.msgChan
}

func (c *console) Start() {
LOOP:
	for {
		select {
		case msg := <-c.msgChan:
			c.write(msg)
		case <-c.quitChan:
			break LOOP
		}
	}

	for {
		if len(c.msgChan) == 0 {
			break
		}

		c.write(<-c.msgChan)
	}
	c.quitChan <- struct{}{} // Notify the cleanup is done.
}

func (c *console) Destroy() {
	c.quitChan <- struct{}{}
	<-c.quitChan

	close(c.msgChan)
	close(c.quitChan)
}
