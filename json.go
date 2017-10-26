package xlog

import (
	"io"
	"os"
	"fmt"
	"encoding/json"
	"reflect"
	"encoding"
)

const JSONFormat MODE = "jsonFormat"

type jsonFormat struct {
	Adapter // needs to compose the adapter
	writer        io.Writer
	valuesDefault []interface{}
}

type JsonFormatConfig struct {
	Level        LEVEL
	BufferSize   int64
	Writer       io.Writer
	ValuesDefault []interface{}
}

func newJsonFormat() Logger {
	return &jsonFormat{
		valuesDefault: make([]interface{}, 0),
		Adapter: Adapter{
			quitChan: make(chan struct{}),
		},
	}
}

func init() {
	Register(JSONFormat, newJsonFormat)
}

func (c *jsonFormat) Level() LEVEL { return c.level }

func (c *jsonFormat) Init(v interface{}) error {
	cfg, ok := v.(JsonFormatConfig)
	if !ok {
		return ErrConfigObject{"JsonFormatConfig", v}
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

func (c *jsonFormat) ExchangeChans(chan<- error) chan *Message {
	c.errorChan = errorChan
	return c.msgChan
}

func (c *jsonFormat) Start() {
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

func (c *jsonFormat) Destroy() {
	c.quitChan <- struct{}{}
	<-c.quitChan

	close(c.msgChan)
	close(c.quitChan)
}

func (c *jsonFormat) write(msg *Message) {
	msg.Body = append(append(c.valuesDefault, "level", msg.Level), msg.Body...)

	n := (len(msg.Body) + 1) / 2 // +1 to handle case when len is odd
	m := make(map[string]interface{}, n)
	for i := 0; i < len(msg.Body); i += 2 {
		k := msg.Body[i]
		var v interface{} = ErrMissing{}
		if i+1 < len(msg.Body) {
			v = msg.Body[i+1]
		}
		merge(m, k, v)
	}

	err := json.NewEncoder(c.writer).Encode(m)
	if err != nil {
		c.errorChan <- err
		return
	}
}

func merge(dst map[string]interface{}, k, v interface{}) {
	var key string
	switch x := k.(type) {
	case string:
		key = x
	case fmt.Stringer:
		key = safeString(x)
	default:
		key = fmt.Sprint(x)
	}

	// We want json.Marshaler and encoding.TextMarshaller to take priority over
	// err.Error() and v.String(). But json.Marshall (called later) does that by
	// default so we force a no-op if it's one of those 2 case.
	switch x := v.(type) {
	case json.Marshaler:
	case encoding.TextMarshaler:
	case error:
		v = safeError(x)
	case fmt.Stringer:
		v = safeString(x)
	}

	dst[key] = v
}

func safeString(str fmt.Stringer) (s string) {
	defer func() {
		if panicVal := recover(); panicVal != nil {
			if v := reflect.ValueOf(str); v.Kind() == reflect.Ptr && v.IsNil() {
				s = "NULL"
			} else {
				panic(panicVal)
			}
		}
	}()
	s = str.String()
	return
}

func safeError(err error) (s interface{}) {
	defer func() {
		if panicVal := recover(); panicVal != nil {
			if v := reflect.ValueOf(err); v.Kind() == reflect.Ptr && v.IsNil() {
				s = nil
			} else {
				panic(panicVal)
			}
		}
	}()
	s = err.Error()
	return
}
