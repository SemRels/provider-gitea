// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The provider-gitea Authors

package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigFromEnvSuccess(t *testing.T) {
	t.Parallel()

	cfg, err := ConfigFromEnv(testEnv(map[string]string{
		"SEMREL_PLUGIN_BASE_URL":   " https://gitea.example.com ",
		"SEMREL_PLUGIN_TOKEN":      " token ",
		"SEMREL_PLUGIN_OWNER":      " owner ",
		"SEMREL_PLUGIN_REPO":       " repo ",
		"SEMREL_PLUGIN_TAG_NAME":   "v1.2.3",
		"SEMREL_PLUGIN_BODY":       "notes",
		"SEMREL_PLUGIN_DRAFT":      "true",
		"SEMREL_PLUGIN_PRERELEASE": "false",
	}))

	require.NoError(t, err)
	require.Equal(t, "https://gitea.example.com", cfg.BaseURL)
	require.Equal(t, "token", cfg.Token)
	require.Equal(t, "owner", cfg.Owner)
	require.Equal(t, "repo", cfg.Repo)
	require.Equal(t, "v1.2.3", cfg.TagName)
	require.Equal(t, "v1.2.3", cfg.Name)
	require.Equal(t, "notes", cfg.Body)
	require.NotNil(t, cfg.Draft)
	require.True(t, *cfg.Draft)
	require.NotNil(t, cfg.Prerelease)
	require.False(t, *cfg.Prerelease)
}

func TestConfigFromEnvUsesFallbacks(t *testing.T) {
	t.Parallel()

	cfg, err := ConfigFromEnv(testEnv(map[string]string{
		"SEMREL_PLUGIN_BASE_URL": "https://gitea.example.com/base",
		"SEMREL_PLUGIN_TOKEN":    "token",
		"SEMREL_PLUGIN_OWNER":    "owner",
		"SEMREL_PLUGIN_REPO":     "repo",
		"SEMREL_NEXT_VERSION":    "1.2.3",
		"SEMREL_CHANGELOG":       "generated notes",
		"SEMREL_IS_PRERELEASE":   "true",
	}))

	require.NoError(t, err)
	require.Equal(t, "1.2.3", cfg.TagName)
	require.Equal(t, "1.2.3", cfg.Name)
	require.Equal(t, "generated notes", cfg.Body)
	require.NotNil(t, cfg.Prerelease)
	require.True(t, *cfg.Prerelease)
	require.Nil(t, cfg.Draft)
}

func TestConfigFromEnvMissingRequired(t *testing.T) {
	t.Parallel()

	cfg, err := ConfigFromEnv(testEnv(map[string]string{}))

	require.Error(t, err)
	require.Equal(t, Config{}, cfg)
	require.Contains(t, err.Error(), "SEMREL_PLUGIN_BASE_URL")
	require.Contains(t, err.Error(), "SEMREL_PLUGIN_TOKEN")
	require.Contains(t, err.Error(), "SEMREL_PLUGIN_OWNER")
	require.Contains(t, err.Error(), "SEMREL_PLUGIN_REPO")
}

func TestConfigFromEnvInvalidBoolean(t *testing.T) {
	t.Parallel()

	cfg, err := ConfigFromEnv(testEnv(map[string]string{
		"SEMREL_PLUGIN_BASE_URL": "https://gitea.example.com",
		"SEMREL_PLUGIN_TOKEN":    "token",
		"SEMREL_PLUGIN_OWNER":    "owner",
		"SEMREL_PLUGIN_REPO":     "repo",
		"SEMREL_TAG_NAME":        "v1.2.3",
		"SEMREL_PLUGIN_DRAFT":    "maybe",
	}))

	require.Error(t, err)
	require.Equal(t, Config{}, cfg)
	require.Contains(t, err.Error(), "SEMREL_PLUGIN_DRAFT")
}

func TestCreateReleaseSuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/v1/repos/owner/repo/releases", r.URL.Path)
		require.Equal(t, "token secret", r.Header.Get("Authorization"))
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "v1.2.3", payload["tag_name"])
		require.Equal(t, "Release 1.2.3", payload["name"])
		require.Equal(t, "notes", payload["body"])
		require.Equal(t, true, payload["draft"])
		require.Equal(t, false, payload["prerelease"])

		w.WriteHeader(http.StatusCreated)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"id":       42,
			"html_url": "https://gitea.example.com/owner/repo/releases/tag/v1.2.3",
		}))
	}))
	defer server.Close()

	creator := New(server.Client())
	release, err := creator.CreateRelease(context.Background(), Config{
		BaseURL:    server.URL + "/",
		Token:      "secret",
		Owner:      "owner",
		Repo:       "repo",
		TagName:    "v1.2.3",
		Name:       "Release 1.2.3",
		Body:       "notes",
		Draft:      boolPtr(true),
		Prerelease: boolPtr(false),
	})

	require.NoError(t, err)
	require.Equal(t, int64(42), release.ID)
	require.Equal(t, "https://gitea.example.com/owner/repo/releases/tag/v1.2.3", release.URL)
}

func TestCreateReleaseOmitsOptionalBooleans(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, hasDraft := payload["draft"]
		_, hasPrerelease := payload["prerelease"]
		require.False(t, hasDraft)
		require.False(t, hasPrerelease)

		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"id":  7,
			"url": "https://gitea.example.com/api/v1/repos/owner/repo/releases/7",
		}))
	}))
	defer server.Close()

	creator := New(server.Client())
	release, err := creator.CreateRelease(context.Background(), Config{
		BaseURL: server.URL,
		Token:   "secret",
		Owner:   "owner",
		Repo:    "repo",
		TagName: "v1.2.3",
		Name:    "v1.2.3",
	})

	require.NoError(t, err)
	require.Equal(t, int64(7), release.ID)
	require.Equal(t, "https://gitea.example.com/api/v1/repos/owner/repo/releases/7", release.URL)
}

func TestCreateReleaseFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "denied", http.StatusUnauthorized)
	}))
	defer server.Close()

	creator := New(server.Client())
	release, err := creator.CreateRelease(context.Background(), Config{
		BaseURL: server.URL,
		Token:   "secret",
		Owner:   "owner",
		Repo:    "repo",
		TagName: "v1.2.3",
		Name:    "v1.2.3",
	})

	require.Nil(t, release)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected status 401 Unauthorized")
}

func TestCreateReleaseValidationFailure(t *testing.T) {
	t.Parallel()

	creator := New(nil)
	release, err := creator.CreateRelease(context.Background(), Config{})

	require.Nil(t, release)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required configuration")
}

func TestCreateReleaseInvalidBaseURL(t *testing.T) {
	t.Parallel()

	creator := New(nil)
	release, err := creator.CreateRelease(context.Background(), Config{
		BaseURL: "not-a-url",
		Token:   "secret",
		Owner:   "owner",
		Repo:    "repo",
		TagName: "v1.2.3",
		Name:    "v1.2.3",
	})

	require.Nil(t, release)
	require.Error(t, err)
	require.Contains(t, err.Error(), "absolute URL")
}

func testEnv(values map[string]string) func(string) string {
	return func(key string) string {
		return values[key]
	}
}

func boolPtr(value bool) *bool {
	return &value
}
