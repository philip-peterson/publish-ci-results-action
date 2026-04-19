package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func checkRunConclusion(r *Results, failOn string) string {
	switch failOn {
	case "test failures":
		if r.Failed > 0 || r.Errors > 0 {
			return "failure"
		}
	case "errors":
		if r.Errors > 0 {
			return "failure"
		}
	}
	if r.Total == 0 {
		return "neutral"
	}
	return "success"
}

func checkRunTitle(r *Results) string {
	var parts []string
	if r.Passed > 0 {
		parts = append(parts, fmt.Sprintf("%d passed", r.Passed))
	}
	if r.Failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", r.Failed))
	}
	if r.Errors > 0 {
		parts = append(parts, fmt.Sprintf("%d errors", r.Errors))
	}
	if r.Skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", r.Skipped))
	}
	if len(parts) == 0 {
		return "No tests found"
	}
	return strings.Join(parts, ", ")
}

func resultsSummaryMarkdown(r *Results) string {
	var b strings.Builder
	fmt.Fprintf(&b, "| | Tests |\n|---|---|\n")
	fmt.Fprintf(&b, "| ✅ Passed | %d |\n", r.Passed)
	if r.Failed > 0 {
		fmt.Fprintf(&b, "| ❌ Failed | %d |\n", r.Failed)
	}
	if r.Errors > 0 {
		fmt.Fprintf(&b, "| 🔥 Errors | %d |\n", r.Errors)
	}
	if r.Skipped > 0 {
		fmt.Fprintf(&b, "| ⏭️ Skipped | %d |\n", r.Skipped)
	}
	fmt.Fprintf(&b, "| **Total** | **%d** |\n", r.Total)
	fmt.Fprintf(&b, "\nDuration: %s\n", formatDuration(r.Duration))

	var failures []TestCase
	for _, c := range r.Cases {
		if c.Status == StatusFailed || c.Status == StatusError {
			failures = append(failures, c)
		}
	}
	if len(failures) > 0 {
		fmt.Fprintf(&b, "\n<details>\n<summary>%d failed</summary>\n\n", len(failures))
		for _, f := range failures {
			name := f.Name
			if f.ClassName != "" {
				name = f.ClassName + "." + f.Name
			}
			if f.Message != "" {
				fmt.Fprintf(&b, "**`%s`**\n```\n%s\n```\n\n", name, f.Message)
			} else {
				fmt.Fprintf(&b, "**`%s`**\n\n", name)
			}
		}
		fmt.Fprint(&b, "</details>\n")
	}
	return b.String()
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func writeJobSummary(cfg *Config, r *Results) error {
	summaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryFile == "" {
		return nil
	}
	content := "## " + cfg.CheckName + "\n\n" + resultsSummaryMarkdown(r)
	return os.WriteFile(summaryFile, []byte(content), 0644)
}
