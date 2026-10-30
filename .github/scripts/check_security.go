package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
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
	Finding *struct {
		OSV   string `json:"osv"`
		Trace []struct {
			Function string `json:"function"`
		} `json:"trace"`
	} `json:"finding,omitempty"`
	// Keep parsing the pre-v1 fields for compatibility with older reports.
	Vulnerability *struct {
		OSV string `json:"osv"`
	} `json:"vulnerability,omitempty"`
	CallStacks []json.RawMessage `json:"call_stacks,omitempty"`
}

func checkGovulncheck(reader io.Reader) {
	decoder := json.NewDecoder(reader)
	reachable := make(map[string]bool)
	events := 0
	for {
		var event govulnEvent
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			fail("invalid govulncheck JSON stream: %v", err)
		}
		events++
		if event.Finding != nil && len(event.Finding.Trace) > 0 && event.Finding.Trace[0].Function != "" {
			reachable[event.Finding.OSV] = true
		}
		if event.Vulnerability != nil && len(event.CallStacks) > 0 {
			reachable[event.Vulnerability.OSV] = true
		}
	}
	if events == 0 {
		fail("empty govulncheck JSON stream")
	}

	if len(reachable) > 0 {
		ids := make([]string, 0, len(reachable))
		for id := range reachable {
			ids = append(ids, id)
		}
		// Stable output makes CI logs easier to compare.
		sort.Strings(ids)
		for _, id := range ids {
			fmt.Fprintf(os.Stderr, "blocking reachable govulncheck finding: %s\n", id)
		}
		os.Exit(1)
	}
	fmt.Println("No reachable govulncheck findings.")
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
