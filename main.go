package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	results, err := parseFiles(cfg.Files, cfg.TimeUnit)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.Token})
	httpClient := oauth2.NewClient(ctx, ts)

	var gh *github.Client
	if cfg.APIURL != "" && cfg.APIURL != "https://api.github.com" {
		gh, err = github.NewClient(httpClient).WithEnterpriseURLs(cfg.APIURL+"/", cfg.APIURL+"/")
		if err != nil {
			return fmt.Errorf("github client: %w", err)
		}
	} else {
		gh = github.NewClient(httpClient)
	}

	if cfg.CheckRun {
		if err := publishCheckRun(ctx, gh, cfg, results); err != nil {
			return fmt.Errorf("check run: %w", err)
		}
	}

	if cfg.JobSummary {
		if err := writeJobSummary(cfg, results); err != nil {
			return fmt.Errorf("job summary: %w", err)
		}
	}

	if cfg.CommentMode != "off" {
		if err := updatePRComment(ctx, gh, cfg, results); err != nil {
			fmt.Fprintf(os.Stderr, "warning: PR comment: %v\n", err)
		}
	}

	switch cfg.FailOn {
	case "test failures":
		if results.Failed > 0 || results.Errors > 0 {
			os.Exit(1)
		}
	case "errors":
		if results.Errors > 0 {
			os.Exit(1)
		}
	}

	return nil
}
