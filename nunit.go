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
	var suiteStack []string
	var cases []TestCase

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "test-suite":
				suiteStack = append(suiteStack, xmlAttr(t.Attr, "name"))
			case "test-case":
				var tc testCase
				if err := dec.DecodeElement(&tc, &t); err != nil {
					return nil, err
				}
				c := TestCase{Suite: last(suiteStack)}
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
				durStr := firstNonEmpty(tc.Duration, tc.Time)
				if f, err := strconv.ParseFloat(durStr, 64); err == nil {
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
		case xml.EndElement:
			if t.Name.Local == "test-suite" && len(suiteStack) > 0 {
				suiteStack = suiteStack[:len(suiteStack)-1]
			}
		}
	}
	return cases, nil
}

func xmlAttr(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func last(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[len(s)-1]
}
