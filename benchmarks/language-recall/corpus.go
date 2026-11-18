package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// corpusRepository is a public repository pinned to a commit. The corpus is
// public on purpose: a measurement nobody else can repeat is a claim.
type corpusRepository struct {
	Language string `json:"language"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Commit   string `json:"commit"`
	// Project is what the oracle type-checks: the repository's own tsconfig for
	// TypeScript, empty for Go (every package, tests included). Kivgraph reads
	// the same project, so the two are asked about the same files.
	Project string `json:"project,omitempty"`
	// Setup installs dependencies. The type checkers need them to resolve
	// imports, and so does Kivgraph.
	Setup []string `json:"-"`

	Path string `json:"-"`
}

func corpus() []corpusRepository {
	return []corpusRepository{
		{
			Language: "go", Name: "cobra", URL: "https://github.com/spf13/cobra.git",
			Commit: "adbc8813901bba65827259daa8e22ff94ec1f30e",
			Setup:  []string{"go mod download"},
		},
		{
			Language: "typescript", Name: "ky", URL: "https://github.com/sindresorhus/ky.git",
			Commit:  "3541888878c4d303a8bc2ad61815f4e048f55a5d",
			Project: "tsconfig.json",
			Setup:   []string{"npm install --no-audit --no-fund --ignore-scripts --fetch-retries=5 --fetch-timeout=120000"},
		},
	}
}

func describe(repositories []corpusRepository) []corpusRepository {
	return append([]corpusRepository(nil), repositories...)
}

// prepare clones the repository at its commit, or reuses a clone already there,
// and runs its setup. A clone that moved off the commit is an error: a result
// must name the source it was measured on.
func prepare(ctx context.Context, work string, repository *corpusRepository) error {
	repository.Path = filepath.Join(work, repository.Name)
	if _, err := os.Stat(filepath.Join(repository.Path, ".git")); err != nil {
		if err := os.MkdirAll(work, 0o755); err != nil {
			return err
		}
		if output, err := exec.CommandContext(ctx, "git", "clone", "-q", repository.URL, repository.Path).CombinedOutput(); err != nil {
			return fmt.Errorf("clone: %w: %s", err, strings.TrimSpace(string(output)))
		}
		if output, err := exec.CommandContext(ctx, "git", "-C", repository.Path, "checkout", "-q", repository.Commit).CombinedOutput(); err != nil {
			return fmt.Errorf("checkout %s: %w: %s", repository.Commit, err, strings.TrimSpace(string(output)))
		}
	}
	head, err := exec.CommandContext(ctx, "git", "-C", repository.Path, "rev-parse", "HEAD").Output()
	if err != nil {
		return err
	}
	if got := strings.TrimSpace(string(head)); got != repository.Commit {
		return fmt.Errorf("the clone is at %s, want %s", got, repository.Commit)
	}
	if status, err := exec.CommandContext(ctx, "git", "-C", repository.Path, "status", "--porcelain", "--untracked-files=no").Output(); err != nil || len(strings.TrimSpace(string(status))) != 0 {
		return fmt.Errorf("the clone has uncommitted changes: a previous run did not restore it")
	}
	for _, line := range repository.Setup {
		fields := strings.Fields(line)
		process := exec.CommandContext(ctx, fields[0], fields[1:]...)
		process.Dir = repository.Path
		process.Env = toolEnvironment()
		if output, err := process.CombinedOutput(); err != nil {
			return fmt.Errorf("%s: %w: %s", line, err, strings.TrimSpace(string(output)))
		}
	}
	return nil
}

// toolEnvironment keeps the toolchains this machine has, and stops pnpm from
// fetching the version a repository's `packageManager` field names.
func toolEnvironment() []string {
	return append(
		os.Environ(),
		"npm_config_pm_on_fail=ignore", "pnpm_config_pm_on_fail=ignore", "COREPACK_ENABLE_STRICT=0",
		"KIVGRAPH_TELEMETRY=0",
	)
}

func toolVersion(ctx context.Context, name string, arguments ...string) string {
	output, err := exec.CommandContext(ctx, name, arguments...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func kivgraphVersion(ctx context.Context, binary string) string {
	return toolVersion(ctx, binary, "version")
}

// typescriptVersion is the compiler the repository itself resolves, which is
// the one the oracle runs.
func typescriptVersion(ctx context.Context, work string, repositories []corpusRepository) string {
	for _, repository := range repositories {
		if repository.Language != "typescript" {
			continue
		}
		process := exec.CommandContext(ctx, "npx", "--no-install", "tsc", "--version")
		process.Dir = repository.Path
		process.Env = toolEnvironment()
		if output, err := process.Output(); err == nil {
			return strings.TrimSpace(string(output))
		}
	}
	return ""
}
