// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The provider-gitea Authors

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	plugin "github.com/SemRels/provider-gitea/internal/plugin"
)

const pluginSchemaVersion = 1

type releaseCreator interface {
	CreateRelease(context.Context, plugin.Config) (*plugin.Release, error)
}

type creatorFactory func(*http.Client) releaseCreator

var newCreator = func(client *http.Client) releaseCreator {
	return plugin.New(client)
}

func run(ctx context.Context, getenv func(string) string, stdout, stderr io.Writer) int {
	_, _ = fmt.Fprintf(stderr, "plugin_schema_version=%d\n", pluginSchemaVersion)
	return runWithCreator(ctx, getenv, stdout, stderr, newCreator)
}

func runWithCreator(ctx context.Context, getenv func(string) string, stdout, stderr io.Writer, factory creatorFactory) int {
	cfg, err := plugin.ConfigFromEnv(getenv)
	if err != nil {
		fmt.Fprintln(stderr, "provider-gitea:", err)
		return 1
	}

	if isDryRun(getenv("SEMREL_DRY_RUN")) {
		fmt.Fprintf(stdout, "[dry-run] would create Gitea release %s for %s/%s\n", cfg.TagName, cfg.Owner, cfg.Repo)
		return 0
	}

	release, err := factory(nil).CreateRelease(ctx, cfg)
	if err != nil {
		fmt.Fprintln(stderr, "provider-gitea:", err)
		return 1
	}

	switch {
	case release == nil:
		return 0
	case release.URL != "":
		fmt.Fprintln(stdout, release.URL)
	default:
		fmt.Fprintf(stdout, "created release %d\n", release.ID)
	}

	return 0
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	os.Exit(run(ctx, os.Getenv, os.Stdout, os.Stderr))
}

func isDryRun(raw string) bool {
	return strings.EqualFold(strings.TrimSpace(raw), "true")
}
