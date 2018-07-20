package main

import (
	"sync"
)

type Set struct {
	m map[string]bool
	sync.RWMutex
}

func (s *Set) Add(item string) {
	s.Lock()
	if s.m == nil {
		s.m = make(map[string]bool)
	}
	s.m[item] = true
	s.Unlock()
}
func (s *Set) AddMany(items ...string) {

	s.Lock()
	if s.m == nil {
		s.m = make(map[string]bool)
	}
	for _, v := range items {
		s.m[v] = true
	}
	s.Unlock()
}

func (s *Set) Remove(item string) {
	s.Lock()
	delete(s.m, item)
	s.Unlock()
}

func (s *Set) Has(item string) bool {
	s.RLock()
	_, ok := s.m[item]
	s.RUnlock()
	return ok
}

func (s *Set) Len() int {
	s.RLock()
	l := len(s.m)
	s.RUnlock()
	return l
}

func (s *Set) Clear() {
	s.Lock()
	s.m = make(map[string]bool)
	s.Unlock()
}

func (s *Set) IsEmpty() bool {
	if s.Len() == 0 {
		return true
	}
	return false
}

func (s *Set) List() []string {
	var list []string
	s.RLock()
	for item := range s.m {
		list = append(list, item)
	}
	s.RUnlock()
	return list
}
