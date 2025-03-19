package main

import (
	"fmt"
	"log"
	"reflect"
	"time"
)

type UpdateManager struct {
	updateTicker *time.Ticker
	callbacks    map[uintptr]func(*Env) error
	doneChannel  chan bool
	environment  *Env
}

func (um *UpdateManager) emit() error {
	for _, fn := range um.callbacks {
		err := fn(um.environment)
		if err != nil {
			return fmt.Errorf("error in update manager callback: %s", err.Error())
		}
	}

	return nil
}

func (um *UpdateManager) UpdateNow() error {
	return um.emit()
}

func (um *UpdateManager) Add(fn func(*Env) error) {
	ptr := reflect.ValueOf(fn).Pointer()
	um.callbacks[ptr] = fn
}

func (um *UpdateManager) Remove(fn func(*Env) error) {
	ptr := reflect.ValueOf(fn).Pointer()
	delete(um.callbacks, ptr)
}

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

func (um *UpdateManager) Stop() {
	um.doneChannel <- true
}

func NewUpdateManager(env *Env, d time.Duration) *UpdateManager {
	return &UpdateManager{
		callbacks:    make(map[uintptr]func(*Env) error),
		doneChannel:  make(chan bool),
		updateTicker: time.NewTicker(d),
		environment:  env,
	}
}
