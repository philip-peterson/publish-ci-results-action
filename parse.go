package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"time"

	junit "github.com/joshdk/go-junit"
)

func parseFiles(patterns []string, timeUnit string) (*Results, error) {
	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("glob %q: %w", pattern, err)
		}
		files = append(files, matches...)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files matched patterns: %v", patterns)
	}

	res := &Results{}
	for _, f := range files {
		cases, err := parseFile(f, timeUnit)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", f, err)
		}
		for _, c := range cases {
			res.add(c)
		}
	}
	return res, nil
}

func parseFile(path, timeUnit string) ([]TestCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	format, err := detectFormat(data)
	if err != nil {
		return nil, err
	}
	switch format {
	case "junit":
		return parseJUnit(data, timeUnit)
	case "nunit":
		return parseNUnit(data, timeUnit)
	case "trx":
		return parseTRX(data)
	default:
		return nil, fmt.Errorf("unknown format %q", format)
	}
}

func detectFormat(data []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", fmt.Errorf("not valid XML: %w", err)
		}
		if se, ok := tok.(xml.StartElement); ok {
			switch se.Name.Local {
			case "testsuites", "testsuite":
				return "junit", nil
			case "test-run", "test-results":
				return "nunit", nil
			case "TestRun":
				return "trx", nil
			default:
				return "", fmt.Errorf("unrecognized root element %q", se.Name.Local)
			}
		}
	}
}

func parseJUnit(data []byte, timeUnit string) ([]TestCase, error) {
	suites, err := junit.Ingest(data)
	if err != nil {
		return nil, err
	}
	var cases []TestCase
	for _, s := range suites {
		collectJUnit(&cases, s, timeUnit)
	}
	return cases, nil
}

func collectJUnit(cases *[]TestCase, s junit.Suite, timeUnit string) {
	for _, t := range s.Tests {
		dur := t.Duration
		if timeUnit == "milliseconds" {
			dur /= 1000
		}
		c := TestCase{
			Suite:     s.Name,
			Name:      t.Name,
			ClassName: t.Classname,
			Duration:  dur,
		}
		switch t.Status {
		case junit.StatusPassed:
			c.Status = StatusPassed
		case junit.StatusFailed:
			c.Status = StatusFailed
			if t.Error != nil {
				c.Message = t.Error.Error()
			}
		case junit.StatusSkipped:
			c.Status = StatusSkipped
		case junit.StatusError:
			c.Status = StatusError
			if t.Error != nil {
				c.Message = t.Error.Error()
			}
		}
		*cases = append(*cases, c)
	}
	for _, nested := range s.Suites {
		collectJUnit(cases, nested, timeUnit)
	}
}

func durationFromSeconds(v float64, unit string) time.Duration {
	if unit == "milliseconds" {
		return time.Duration(v * float64(time.Millisecond))
	}
	return time.Duration(v * float64(time.Second))
}
