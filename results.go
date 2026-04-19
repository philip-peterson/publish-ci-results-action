package main

import "time"

type Status int

const (
	StatusPassed Status = iota
	StatusFailed
	StatusSkipped
	StatusError
)

type TestCase struct {
	Suite     string
	Name      string
	ClassName string
	File      string
	Line      int
	Duration  time.Duration
	Status    Status
	Message   string
}

type Results struct {
	Cases    []TestCase
	Total    int
	Passed   int
	Failed   int
	Skipped  int
	Errors   int
	Duration time.Duration
}

func (r *Results) add(c TestCase) {
	r.Cases = append(r.Cases, c)
	r.Total++
	r.Duration += c.Duration
	switch c.Status {
	case StatusPassed:
		r.Passed++
	case StatusFailed:
		r.Failed++
	case StatusSkipped:
		r.Skipped++
	case StatusError:
		r.Errors++
	}
}
