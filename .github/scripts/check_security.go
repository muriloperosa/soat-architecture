package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fail("usage: check_security.go <govulncheck|gosec> <report.json>")
	}

	file, err := os.Open(os.Args[2])
	if err != nil {
		fail("cannot open report: %v", err)
	}
	defer file.Close()

	switch os.Args[1] {
	case "govulncheck":
		checkGovulncheck(file)
	case "gosec":
		checkGosec(file)
	default:
		fail("unknown scanner %q", os.Args[1])
	}
}

type govulnEvent struct {
	Vulnerability *struct {
		OSV string `json:"osv"`
	} `json:"vulnerability,omitempty"`
	OSV *struct {
		ID               string `json:"id"`
		DatabaseSpecific struct {
			Severity string `json:"severity"`
		} `json:"database_specific"`
		Severity []struct {
			Type  string `json:"type"`
			Score string `json:"score"`
		} `json:"severity"`
	} `json:"osv,omitempty"`
	CallStacks []json.RawMessage `json:"call_stacks,omitempty"`
	Scanning   *struct {
		Message string `json:"message"`
	} `json:"scanning,omitempty"`
}

func checkGovulncheck(reader io.Reader) {
	decoder := json.NewDecoder(reader)
	severityByID := make(map[string]string)
	reachable := make(map[string]bool)
	for {
		var event govulnEvent
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			fail("invalid govulncheck JSON stream: %v", err)
		}
		if event.OSV != nil {
			severity := strings.ToUpper(event.OSV.DatabaseSpecific.Severity)
			if severity == "" {
				for _, item := range event.OSV.Severity {
					if strings.EqualFold(item.Type, "CVSS_V3") {
						severity = severityFromCVSS(item.Score)
						break
					}
				}
			}
			severityByID[event.OSV.ID] = severity
		}
		if event.Vulnerability != nil && len(event.CallStacks) > 0 {
			reachable[event.Vulnerability.OSV] = true
		}
	}

	findings := 0
	for id := range reachable {
		severity := severityByID[id]
		if severity == "HIGH" || severity == "CRITICAL" {
			findings++
			fmt.Fprintf(os.Stderr, "blocking reachable %s vulnerability: %s\n", severity, id)
		}
	}
	if findings > 0 {
		os.Exit(1)
	}
	fmt.Println("No high or critical govulncheck findings.")
}

func severityFromCVSS(score string) string {
	var value float64
	if _, err := fmt.Sscanf(score, "%f", &value); err != nil {
		return ""
	}
	switch {
	case value >= 9.0:
		return "CRITICAL"
	case value >= 7.0:
		return "HIGH"
	default:
		return ""
	}
}

type gosecReport struct {
	Issues []struct {
		Severity   string `json:"severity"`
		Confidence string `json:"confidence"`
		RuleID     string `json:"rule_id"`
		Details    string `json:"details"`
		File       string `json:"file"`
		Line       string `json:"line"`
	} `json:"Issues"`
}

func checkGosec(reader io.Reader) {
	var report gosecReport
	if err := json.NewDecoder(reader).Decode(&report); err != nil {
		fail("invalid gosec JSON report: %v", err)
	}

	findings := 0
	for _, issue := range report.Issues {
		severity := strings.ToUpper(issue.Severity)
		confidence := strings.ToUpper(issue.Confidence)
		if (severity == "HIGH" || severity == "CRITICAL") && confidence == "HIGH" {
			findings++
			fmt.Fprintf(os.Stderr, "blocking gosec finding %s (%s/%s): %s:%s %s\n", issue.RuleID, severity, confidence, issue.File, issue.Line, issue.Details)
		}
	}
	if findings > 0 {
		os.Exit(1)
	}
	fmt.Println("No high-confidence high or critical gosec findings.")
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(2)
}
