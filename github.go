package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/v62/github"
)

const commentMarker = "<!-- publish-unit-test-results -->"

func publishCheckRun(ctx context.Context, gh *github.Client, cfg *Config, r *Results) error {
	now := github.Timestamp{Time: time.Now()}
	_, _, err := gh.Checks.CreateCheckRun(ctx, cfg.Owner, cfg.Repo, github.CreateCheckRunOptions{
		Name:        cfg.CheckName,
		HeadSHA:     cfg.SHA,
		Status:      github.String("completed"),
		Conclusion:  github.String(checkRunConclusion(r, cfg.FailOn)),
		CompletedAt: &now,
		Output: &github.CheckRunOutput{
			Title:   github.String(checkRunTitle(r)),
			Summary: github.String(resultsSummaryMarkdown(r)),
		},
	})
	return err
}


func updatePRComment(ctx context.Context, gh *github.Client, cfg *Config, r *Results) error {
	prNum, err := findPR(ctx, gh, cfg)
	if err != nil {
		return err
	}
	if prNum == 0 {
		log.Printf("PR comment skipped: no open PR found for SHA %s", cfg.SHA)
		return nil
	}

	body := commentMarker + "\n## " + cfg.CheckName + "\n\n" + resultsSummaryMarkdown(r)

	comments, _, err := gh.Issues.ListComments(ctx, cfg.Owner, cfg.Repo, prNum, nil)
	if err != nil {
		return err
	}

	for _, c := range comments {
		if strings.Contains(c.GetBody(), commentMarker) {
			if cfg.CommentMode == "changes" && c.GetBody() == body {
				return nil
			}
			_, _, err = gh.Issues.EditComment(ctx, cfg.Owner, cfg.Repo, c.GetID(), &github.IssueComment{
				Body: github.String(body),
			})
			return err
		}
	}

	_, _, err = gh.Issues.CreateComment(ctx, cfg.Owner, cfg.Repo, prNum, &github.IssueComment{
		Body: github.String(body),
	})
	return err
}

func findPR(ctx context.Context, gh *github.Client, cfg *Config) (int, error) {
	if cfg.PRNumber != 0 {
		log.Printf("PR comment: using PR #%d from GITHUB_REF", cfg.PRNumber)
		return cfg.PRNumber, nil
	}

	log.Printf("PR comment: looking up open PR for SHA %s via API", cfg.SHA)
	prs, _, err := gh.PullRequests.ListPullRequestsWithCommit(ctx, cfg.Owner, cfg.Repo, cfg.SHA, nil)
	if err != nil {
		return 0, err
	}
	for _, pr := range prs {
		if pr.GetState() == "open" {
			log.Printf("PR comment: found PR #%d", pr.GetNumber())
			return pr.GetNumber(), nil
		}
	}
	log.Printf("PR comment: API returned %d PRs for SHA %s, none open", len(prs), cfg.SHA)
	return 0, nil
}
