package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Token       string
	Files       []string
	CheckName   string
	FailOn      string // "nothing" | "errors" | "test failures"
	CommentMode string // "always" | "changes" | "off"
	CheckRun    bool
	JobSummary  bool
	TimeUnit    string // "seconds" | "milliseconds"

	APIURL    string
	Owner     string
	Repo      string
	SHA       string
}

func loadConfig() (*Config, error) {
	cfg := &Config{
		Token:       firstNonEmpty(os.Getenv("INPUT_GITHUB_TOKEN"), os.Getenv("GITHUB_TOKEN")),
		CheckName:   envOr("INPUT_CHECK_NAME", "Test Results"),
		FailOn:      envOr("INPUT_FAIL_ON", "test failures"),
		CommentMode: envOr("INPUT_COMMENT_MODE", "always"),
		CheckRun:    envBool("INPUT_CHECK_RUN", true),
		JobSummary:  envBool("INPUT_JOB_SUMMARY", true),
		TimeUnit:    envOr("INPUT_TIME_UNIT", "seconds"),
		APIURL:      envOr("GITHUB_API_URL", "https://api.github.com"),
		SHA:         os.Getenv("GITHUB_SHA"),
	}

	repo := os.Getenv("GITHUB_REPOSITORY")
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid GITHUB_REPOSITORY: %q", repo)
	}
	cfg.Owner, cfg.Repo = parts[0], parts[1]

	filesRaw := firstNonEmpty(os.Getenv("INPUT_FILES"), os.Getenv("INPUT_REPORT_PATHS"))
	for _, line := range strings.Split(filesRaw, "\n") {
		for _, f := range strings.Fields(line) {
			cfg.Files = append(cfg.Files, f)
		}
	}
	if len(cfg.Files) == 0 {
		return nil, fmt.Errorf("config: INPUT_FILES is required")
	}

	if cfg.FailOn == "test failures" {
		if v := os.Getenv("INPUT_FAIL_ON_FAILURE"); v != "" {
			if b, err := strconv.ParseBool(v); err == nil {
				if b {
					cfg.FailOn = "test failures"
				} else {
					cfg.FailOn = "nothing"
				}
			}
		}
	}

	return cfg, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
