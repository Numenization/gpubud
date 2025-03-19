package main

import (
	"fmt"
	"log"
	"reflect"
	"time"
)

// A structure for asynchronously running several different functions on a loop.
// Functions will be added to a map of callbacks that, when the UpdateManager is started,
// will run on loop every d duration until the manager is stopped.
type UpdateManager struct {
	updateTicker *time.Ticker
	callbacks    map[uintptr]func(*Env) error
	doneChannel  chan bool
	environment  *Env
}

// Runs the callbacks for the UpdateManager
func (um *UpdateManager) emit() error {
	for _, fn := range um.callbacks {
		err := fn(um.environment)
		if err != nil {
			return fmt.Errorf("error in update manager callback: %s", err.Error())
		}
	}

	return nil
}

// Forces the UpdateManager to run all callbacks now
func (um *UpdateManager) UpdateNow() error {
	return um.emit()
}

// Adds a function to the callback map. Will run every time the UpdateManager loops
func (um *UpdateManager) Add(fn func(*Env) error) {
	ptr := reflect.ValueOf(fn).Pointer()
	um.callbacks[ptr] = fn
}

// Removes a function from the callback map
func (um *UpdateManager) Remove(fn func(*Env) error) {
	ptr := reflect.ValueOf(fn).Pointer()
	delete(um.callbacks, ptr)
}

// Starts the UpdateManager loop. Will run until an error from a callback or until the doneChannel recieves true
func (um *UpdateManager) Start() {
	log.Println("Starting update loop")

	go func() {
		for {
			select {
			case <-um.doneChannel:
				log.Println("Update loop stopped")
				return
			case <-um.updateTicker.C:
				err := um.emit()
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()
}

// Stops the UpdateManager loop by sending true through the doneChannel
func (um *UpdateManager) Stop() {
	um.doneChannel <- true
}

// Creates a new UpdateManager whose update interval is set by a duration d
func NewUpdateManager(env *Env, d time.Duration) *UpdateManager {
	return &UpdateManager{
		callbacks:    make(map[uintptr]func(*Env) error),
		doneChannel:  make(chan bool),
		updateTicker: time.NewTicker(d),
		environment:  env,
	}
}
