package node

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type command struct {
	c   string
	a   string
	ctx context.Context
	cf  context.CancelFunc
}

func (s *command) string() string {
	return fmt.Sprintf("%s:%s", s.c, s.a)
}

type queue struct {
	c        chan *command
	commands map[string]*command
	mutex    sync.Mutex
}

func (s *queue) init(len int) *queue {
	s.c = make(chan *command, len)
	s.commands = make(map[string]*command)
	return s
}

func (s *queue) loop() {
	for {
		time.Sleep(5 * time.Minute)
		s.mutex.Lock()
		for k, v := range s.commands {
			select {
			case <-v.ctx.Done():
				delete(s.commands, k)
			default:
				continue
			}
		}
		s.mutex.Unlock()
	}
}

func (s *queue) push(c *command) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	select {
	case s.c <- c:
		s.commands[c.string()] = c
		return nil
	default:
		return errors.New("queue fulled")
	}
}

func (s *queue) get(c string) *command {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if v, ok := s.commands[c]; ok {
		return v
	}
	return nil
}

// func (s *queue) pop() *command {
// 	s.mutex.Lock()
// 	defer s.mutex.Unlock()
// 	c := <-s.c
// 	delete(s.commands, c.string())
// 	return c
// }

func (s *queue) del(c string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.commands, c)
}
