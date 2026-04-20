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
	Name      string
	ClassName string
	Duration  time.Duration
	Status    Status
	Message   string
}

type Results struct {
	Failures []TestCase
	Total    int
	Passed   int
	Failed   int
	Skipped  int
	Errors   int
	Duration time.Duration
}

func (r *Results) add(c TestCase) {
	r.Total++
	r.Duration += c.Duration
	switch c.Status {
	case StatusPassed:
		r.Passed++
	case StatusFailed:
		r.Failed++
		r.Failures = append(r.Failures, c)
	case StatusSkipped:
		r.Skipped++
	case StatusError:
		r.Errors++
		r.Failures = append(r.Failures, c)
	}
}
