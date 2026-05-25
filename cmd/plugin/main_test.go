// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The provider-gitea Authors

package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	plugin "github.com/SemRels/provider-gitea/internal/plugin"
	"github.com/stretchr/testify/require"
)

type fakeCreator struct {
	release *plugin.Release
	err     error
	cfg     plugin.Config
	called  bool
}

func (f *fakeCreator) CreateRelease(_ context.Context, cfg plugin.Config) (*plugin.Release, error) {
	f.called = true
	f.cfg = cfg
	return f.release, f.err
}

func TestRunSuccess(t *testing.T) {
	t.Parallel()

	fake := &fakeCreator{release: &plugin.Release{ID: 42, URL: "https://gitea.example.com/owner/repo/releases/tag/v1.2.3"}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runWithCreator(context.Background(), testEnv(map[string]string{
		"SEMREL_PLUGIN_BASE_URL": "https://gitea.example.com",
		"SEMREL_PLUGIN_TOKEN":    "token",
		"SEMREL_PLUGIN_OWNER":    "owner",
		"SEMREL_PLUGIN_REPO":     "repo",
		"SEMREL_TAG_NAME":        "v1.2.3",
		"SEMREL_CHANGELOG":       "notes",
	}), &stdout, &stderr, func(*http.Client) releaseCreator {
		return fake
	})

	require.Equal(t, 0, code)
	require.True(t, fake.called)
	require.Equal(t, "v1.2.3", fake.cfg.TagName)
	require.Equal(t, "notes", fake.cfg.Body)
	require.Equal(t, "https://gitea.example.com/owner/repo/releases/tag/v1.2.3\n", stdout.String())
	require.Empty(t, stderr.String())
}

func TestRunDryRun(t *testing.T) {
	t.Parallel()

	fake := &fakeCreator{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runWithCreator(context.Background(), testEnv(map[string]string{
		"SEMREL_PLUGIN_BASE_URL": "https://gitea.example.com",
		"SEMREL_PLUGIN_TOKEN":    "token",
		"SEMREL_PLUGIN_OWNER":    "owner",
		"SEMREL_PLUGIN_REPO":     "repo",
		"SEMREL_TAG_NAME":        "v1.2.3",
		"SEMREL_DRY_RUN":         "true",
	}), &stdout, &stderr, func(*http.Client) releaseCreator {
		return fake
	})

	require.Equal(t, 0, code)
	require.False(t, fake.called)
	require.Contains(t, stdout.String(), "[dry-run] would create Gitea release v1.2.3 for owner/repo")
	require.Empty(t, stderr.String())
}

func TestRunValidationError(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runWithCreator(context.Background(), testEnv(map[string]string{}), &stdout, &stderr, func(*http.Client) releaseCreator {
		t.Fatal("creator should not be constructed when config is invalid")
		return nil
	})

	require.Equal(t, 1, code)
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), "SEMREL_PLUGIN_BASE_URL")
}

func TestRunCreateError(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runWithCreator(context.Background(), testEnv(map[string]string{
		"SEMREL_PLUGIN_BASE_URL": "https://gitea.example.com",
		"SEMREL_PLUGIN_TOKEN":    "token",
		"SEMREL_PLUGIN_OWNER":    "owner",
		"SEMREL_PLUGIN_REPO":     "repo",
		"SEMREL_TAG_NAME":        "v1.2.3",
	}), &stdout, &stderr, func(*http.Client) releaseCreator {
		return &fakeCreator{err: errors.New("boom")}
	})

	require.Equal(t, 1, code)
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), "provider-gitea: boom")
}

func TestRunSuccessWithoutURL(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runWithCreator(context.Background(), testEnv(map[string]string{
		"SEMREL_PLUGIN_BASE_URL": "https://gitea.example.com",
		"SEMREL_PLUGIN_TOKEN":    "token",
		"SEMREL_PLUGIN_OWNER":    "owner",
		"SEMREL_PLUGIN_REPO":     "repo",
		"SEMREL_TAG_NAME":        "v1.2.3",
	}), &stdout, &stderr, func(*http.Client) releaseCreator {
		return &fakeCreator{release: &plugin.Release{ID: 7}}
	})

	require.Equal(t, 0, code)
	require.Equal(t, "created release 7\n", stdout.String())
	require.Empty(t, stderr.String())
}

func TestRunSuccessWithNilRelease(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runWithCreator(context.Background(), testEnv(map[string]string{
		"SEMREL_PLUGIN_BASE_URL": "https://gitea.example.com",
		"SEMREL_PLUGIN_TOKEN":    "token",
		"SEMREL_PLUGIN_OWNER":    "owner",
		"SEMREL_PLUGIN_REPO":     "repo",
		"SEMREL_TAG_NAME":        "v1.2.3",
	}), &stdout, &stderr, func(*http.Client) releaseCreator {
		return &fakeCreator{}
	})

	require.Equal(t, 0, code)
	require.Empty(t, stdout.String())
	require.Empty(t, stderr.String())
}

func testEnv(values map[string]string) func(string) string {
	return func(key string) string {
		return values[key]
	}
}
