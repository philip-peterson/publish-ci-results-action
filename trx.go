package main

import (
	"bytes"
	"encoding/xml"
	"strconv"
	"strings"
	"time"
)

func parseTRX(data []byte) ([]TestCase, error) {
	// Strip default namespace so plain struct tags match.
	data = bytes.ReplaceAll(data,
		[]byte(`xmlns="http://microsoft.com/schemas/VisualStudio/TeamTest/2010"`),
		[]byte{})

	type testMethod struct {
		ClassName string `xml:"className,attr"`
	}
	type unitTest struct {
		ID         string     `xml:"id,attr"`
		TestMethod testMethod `xml:"TestMethod"`
	}
	type errorInfo struct {
		Message    string `xml:"Message"`
		StackTrace string `xml:"StackTrace"`
	}
	type output struct {
		ErrorInfo *errorInfo `xml:"ErrorInfo"`
	}
	type result struct {
		TestID   string  `xml:"testId,attr"`
		TestName string  `xml:"testName,attr"`
		Duration string  `xml:"duration,attr"`
		Outcome  string  `xml:"outcome,attr"`
		Output   *output `xml:"Output"`
	}
	type testRun struct {
		Definitions []unitTest `xml:"TestDefinitions>UnitTest"`
		Results     []result   `xml:"Results>UnitTestResult"`
	}

	var run testRun
	if err := xml.Unmarshal(data, &run); err != nil {
		return nil, err
	}

	classNames := make(map[string]string, len(run.Definitions))
	for _, d := range run.Definitions {
		classNames[d.ID] = d.TestMethod.ClassName
	}

	cases := make([]TestCase, 0, len(run.Results))
	for _, r := range run.Results {
		c := TestCase{
			Name:      r.TestName,
			ClassName: classNames[r.TestID],
			Duration:  parseTRXDuration(r.Duration),
		}
		switch r.Outcome {
		case "Passed":
			c.Status = StatusPassed
		case "Failed":
			c.Status = StatusFailed
			if r.Output != nil && r.Output.ErrorInfo != nil {
				c.Message = r.Output.ErrorInfo.Message
			}
		case "NotExecuted", "Inconclusive":
			c.Status = StatusSkipped
		default:
			c.Status = StatusError
		}
		cases = append(cases, c)
	}
	return cases, nil
}

// parseTRXDuration parses TRX duration format "HH:MM:SS.fffffff".
func parseTRXDuration(s string) time.Duration {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	sec, _ := strconv.ParseFloat(parts[2], 64)
	return time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute +
		time.Duration(sec*float64(time.Second))
}
