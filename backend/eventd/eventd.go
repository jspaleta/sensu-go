package eventd

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var (
	logger = logrus.WithFields(logrus.Fields{
		"component": "eventd",
	})
)

// Eventd handles incoming sensu events and stores them in etcd.
type Eventd struct {
	Store        store.Store
	MessageBus   messaging.MessageBus
	HandlerCount int

	eventChan    chan interface{}
	errChan      chan error
	shutdownChan chan struct{}
	wg           *sync.WaitGroup
}

// Start eventd.
func (e *Eventd) Start() error {
	if e.Store == nil {
		return errors.New("no store found")
	}

	if e.MessageBus == nil {
		return errors.New("no message bus found")
	}

	if e.HandlerCount == 0 {
		e.HandlerCount = 10
	}

	e.errChan = make(chan error, 1)
	e.shutdownChan = make(chan struct{}, 1)

	ch := make(chan interface{}, 100)
	e.eventChan = ch

	err := e.MessageBus.Subscribe(messaging.TopicEventRaw, "eventd", ch)
	if err != nil {
		return err
	}

	e.wg = &sync.WaitGroup{}
	e.wg.Add(e.HandlerCount)
	e.startHandlers()

	return nil
}

func (e *Eventd) startHandlers() {
	for i := 0; i < e.HandlerCount; i++ {
		go func() {
			var event *types.Event
			defer e.wg.Done()

			for {
				select {
				case <-e.shutdownChan:
					return

				case msg, ok := <-e.eventChan:
					// The message bus will close channels when it's shut down which means
					// we will end up reading from a closed channel. If it's closed,
					// return from this goroutine and emit a fatal error. It is then
					// the responsility of eventd's parent to shutdown eventd.
					//
					// NOTE: Should that be the case? If eventd is signalling that it has,
					// effectively, shutdown, why would something else be responsible for
					// shutting it down.
					if !ok {
						e.errChan <- errors.New("event channel closed")
						return
					}

					event, ok = msg.(*types.Event)
					if !ok {
						logger.Errorf("received non-Event on event channel")
						continue
					}

					if event.Check == nil || event.Entity == nil {
						logger.Error("eventd - error handling event: event invalid")
						continue
					}

					if err := event.Check.Validate(); err != nil {
						logger.Errorf("eventd - error handling event: %s", err.Error())
						continue
					}

					if err := event.Entity.Validate(); err != nil {
						logger.Errorf("eventd - error handling event: %s", err.Error())
						continue
					}

					prevEvent, err := e.Store.GetEventByEntityCheck(event.Entity.ID, event.Check.Name)
					if err != nil {
						logger.Errorf("eventd - error handling event: %s", err.Error())
						continue
					}

					if prevEvent == nil {
						err = e.Store.UpdateEvent(event)
						if err != nil {
							logger.Errorf("eventd - error handling event: %s", err.Error())
						}
						continue
					}

					if prevEvent.Check == nil {
						logger.Errorf("eventd - error handling event: invalid previous event")
						continue
					}

					event.Check.MergeWith(prevEvent.Check)

					err = e.Store.UpdateEvent(event)
					if err != nil {
						logger.Errorf("eventd - error handling event: %s", err.Error())
					}

					eventBytes, err := json.Marshal(event)
					if err != nil {
						logger.Errorf("error handling event: %s", err.Error())
					}
					e.MessageBus.Publish(messaging.TopicEvent, eventBytes)
				}
			}
		}()
	}
}

// Stop eventd.
func (e *Eventd) Stop() error {
	logger.Info("shutting down eventd")
	close(e.shutdownChan)
	e.wg.Wait()
	return nil
}

// Status returns an error if eventd is unhealthy.
func (e *Eventd) Status() error {
	return nil
}

// Err returns a channel to listen for terminal errors on.
func (e *Eventd) Err() <-chan error {
	return e.errChan
}
