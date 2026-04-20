package main

import (
	"bytes"
	"encoding/xml"
	"strconv"
	"strings"
)

func parseNUnit(data []byte, timeUnit string) ([]TestCase, error) {
	type failure struct {
		Message string `xml:"message"`
	}
	type testCase struct {
		Name     string   `xml:"name,attr"`
		FullName string   `xml:"fullname,attr"`
		Result   string   `xml:"result,attr"`  // NUnit 3: Passed/Failed/Skipped/Inconclusive
		Success  string   `xml:"success,attr"` // NUnit 2: True/False
		Duration string   `xml:"duration,attr"`
		Time     string   `xml:"time,attr"` // NUnit 2
		Failure  *failure `xml:"failure"`
	}

	dec := xml.NewDecoder(bytes.NewReader(data))
	var cases []TestCase

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "test-case" {
			continue
		}
		var tc testCase
		if err := dec.DecodeElement(&tc, &se); err != nil {
			return nil, err
		}
		c := TestCase{}
		if tc.FullName != "" {
			if i := strings.LastIndex(tc.FullName, "."); i >= 0 {
				c.ClassName = tc.FullName[:i]
				c.Name = tc.FullName[i+1:]
			} else {
				c.Name = tc.FullName
			}
		} else {
			c.Name = tc.Name
		}
		if f, err := strconv.ParseFloat(firstNonEmpty(tc.Duration, tc.Time), 64); err == nil {
			c.Duration = durationFromSeconds(f, timeUnit)
		}
		switch tc.Result {
		case "Passed", "Success":
			c.Status = StatusPassed
		case "Failed":
			c.Status = StatusFailed
			if tc.Failure != nil {
				c.Message = tc.Failure.Message
			}
		case "Skipped", "Inconclusive", "Ignored":
			c.Status = StatusSkipped
		default:
			switch tc.Success {
			case "True":
				c.Status = StatusPassed
			case "False":
				c.Status = StatusFailed
				if tc.Failure != nil {
					c.Message = tc.Failure.Message
				}
			default:
				c.Status = StatusSkipped
			}
		}
		cases = append(cases, c)
	}
	return cases, nil
}
