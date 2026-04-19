package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v62/github"
)

const commentMarker = "<!-- publish-unit-test-results -->"

func publishCheckRun(ctx context.Context, gh *github.Client, cfg *Config, r *Results) error {
	conclusion := checkRunConclusion(r, cfg.FailOn)
	title := checkRunTitle(r)
	summary := resultsSummaryMarkdown(r)
	annotations := buildAnnotations(r)

	now := github.Timestamp{Time: time.Now()}
	run, _, err := gh.Checks.CreateCheckRun(ctx, cfg.Owner, cfg.Repo, github.CreateCheckRunOptions{
		Name:        cfg.CheckName,
		HeadSHA:     cfg.SHA,
		Status:      github.String("completed"),
		Conclusion:  github.String(conclusion),
		CompletedAt: &now,
		Output: &github.CheckRunOutput{
			Title:       github.String(title),
			Summary:     github.String(summary),
			Annotations: batch(annotations, 0, 50),
		},
	})
	if err != nil {
		return err
	}

	for i := 50; i < len(annotations); i += 50 {
		_, _, err = gh.Checks.UpdateCheckRun(ctx, cfg.Owner, cfg.Repo, run.GetID(), github.UpdateCheckRunOptions{
			Name: cfg.CheckName,
			Output: &github.CheckRunOutput{
				Title:       github.String(title),
				Summary:     github.String(summary),
				Annotations: batch(annotations, i, i+50),
			},
		})
		if err != nil {
			return fmt.Errorf("uploading annotations batch %d: %w", i/50, err)
		}
	}
	return nil
}

func buildAnnotations(r *Results) []*github.CheckRunAnnotation {
	var out []*github.CheckRunAnnotation
	for _, c := range r.Cases {
		if (c.Status != StatusFailed && c.Status != StatusError) || c.File == "" {
			continue
		}
		line := c.Line
		if line == 0 {
			line = 1
		}
		name := c.Name
		if c.ClassName != "" {
			name = c.ClassName + "." + c.Name
		}
		msg := c.Message
		if msg == "" {
			msg = name + " failed"
		}
		out = append(out, &github.CheckRunAnnotation{
			Path:            github.String(c.File),
			StartLine:       github.Int(line),
			EndLine:         github.Int(line),
			AnnotationLevel: github.String("failure"),
			Title:           github.String(name),
			Message:         github.String(msg),
		})
	}
	return out
}

func batch(s []*github.CheckRunAnnotation, lo, hi int) []*github.CheckRunAnnotation {
	if lo >= len(s) {
		return nil
	}
	if hi > len(s) {
		hi = len(s)
	}
	return s[lo:hi]
}

func updatePRComment(ctx context.Context, gh *github.Client, cfg *Config, r *Results) error {
	prNum, err := findPR(ctx, gh, cfg)
	if err != nil {
		return err
	}
	if prNum == 0 {
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
	prs, _, err := gh.PullRequests.ListPullRequestsWithCommit(ctx, cfg.Owner, cfg.Repo, cfg.SHA, nil)
	if err != nil {
		return 0, err
	}
	for _, pr := range prs {
		if pr.GetState() == "open" {
			return pr.GetNumber(), nil
		}
	}
	return 0, nil
}
