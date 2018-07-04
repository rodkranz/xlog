package xlog

import "fmt"

type Logger interface {
	// Level returns minimum level of given logger.
	Level() LEVEL
	// Init accepts a config struct specific for given logger and performs any necessary initialization.
	Init(interface{}) error
	// ExchangeChans accepts error channel, and returns message receive channel.
	ExchangeChans(chan<- error) chan *Message
	// Start starts message processing.
	Start()
	// Destroy releases all resources.
	Destroy()
}

type Adapter struct {
	level     LEVEL
	msgChan   chan *Message
	quitChan  chan struct{}
	errorChan chan<- error
}

type Factory func() Logger

var factories = map[MODE]Factory{}

func Register(mode MODE, f Factory) {
	if f == nil {
		panic("xlog: register function is nil")
	}
	if factories[mode] != nil {
		panic("xlog: register duplicated mode '" + mode + "'")
	}
	factories[mode] = f
}

type receiver struct {
	Logger
	mode    MODE
	msgChan chan *Message
}

var (
	// receivers is a list of loggers with their message channel for broadcasting.
	receivers []*receiver

	errorChan = make(chan error, 5)
	quitChan  = make(chan struct{})
)

func init() {
	// Start background error handling goroutine.
	go func() {
		for {
			select {
			case err := <-errorChan:
				fmt.Printf("xlog: unable to write message: %v\n", err)
			case <-quitChan:
				return
			}
		}
	}()
}

func New(mode MODE, cfg interface{}) error {
	factory, ok := factories[mode]
	if !ok {
		return fmt.Errorf("unknown mode '%s'", mode)
	}

	logger := factory()
	if err := logger.Init(cfg); err != nil {
		return err
	}

	msgChan := logger.ExchangeChans(errorChan)

	hasFound := false
	for i := range receivers {
		if receivers[i].mode == mode {
			hasFound = true

			// Release previous logger.
			receivers[i].Destroy()

			// Update info to new one.
			receivers[i].Logger = logger
			receivers[i].msgChan = msgChan
			break
		}
	}

	if !hasFound {
		receivers = append(receivers, &receiver{
			Logger:  logger,
			mode:    mode,
			msgChan: msgChan,
		})
	}

	go logger.Start()
	return nil
}

func Delete(mode MODE) {
	foundIdx := -1
	for i := range receivers {
		if receivers[i].mode == mode {
			foundIdx = i
			receivers[i].Destroy()
		}
	}

	if foundIdx >= 0 {
		newList := make([]*receiver, len(receivers)-1)
		copy(newList, receivers[:foundIdx])
		copy(newList[foundIdx:], receivers[foundIdx+1:])
		receivers = newList
	}
}
