package vfsm

import (
	"errors"
	"fmt"
	"sync"
)

type (
	StateDescriptor struct {
		Name        string
		Description string
		Next        []string
	}
	StateObjectDescriptor struct {
		Name        string
		Description string
		States      []StateDescriptor
	}
	StatefulObjectIf interface {
		IsStopped() bool
		Error() error
		Descriptor() StateObjectDescriptor
		Current() string
		Shutdown() error
		Start() error
	}

	VFSM[T StatefulObjectIf] struct {
		body         T
		mutex        sync.Mutex
		running      bool
		current      *State[T]
		states       map[string]*State[T]
		primaryState string
	}
)

var errAlreadyRun = errors.New("already run")
var errNextStateNotFound = errors.New("next state not found")

func NewVFSM[T StatefulObjectIf](origin string, states ...*State[T]) (vptr *VFSM[T], err error) {
	vptr = &VFSM[T]{}
	for _, sptr := range states {
		if _, found := vptr.states[sptr.id]; found {
			return nil, fmt.Errorf("duplicated state %s", sptr.id)
		}
		vptr.states[sptr.id] = sptr
	}
	return vptr, nil
}

func (f *VFSM[T]) Start(params map[string]any) error {
	if f.running {
		return errAlreadyRun
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()

	st, found := f.states[f.primaryState]
	if !found {
		return errNextStateNotFound
	}

	err := f.body.Start()
	if err != nil {
		return err
	}
	if err = st.enter(f.body); err != nil {
		return err
	}
	f.current = st
	return err
}

func (f *VFSM[T]) move(n string, needExit bool) error {
	if len(n) == 0 {
		return nil
	}
	if needExit {
		f.current.Exit(f.body)
	}
	var st *State[T]
	var found bool
	if st, found = f.states[n]; !found {
		return errNextStateNotFound
	}
	var err = st.enter(f.body)
	if err == nil {
		return err
	}

	return err
}

func (f *VFSM[T]) Move(params map[string]any) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	var err error
	var next string
	if next, err = f.current.Next(f.body, params); err != nil {
		err = f.move(next, false)
		if err != nil {
			return err
		}
	} else {
		f.move(next, true)
	}
	return err
}
