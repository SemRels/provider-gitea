// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The provider-gitea Authors

package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Config contains the release parameters passed to the Gitea API.
type Config struct {
	BaseURL    string
	Token      string
	Owner      string
	Repo       string
	TagName    string
	Name       string
	Body       string
	Draft      *bool
	Prerelease *bool
}

// Release contains the fields returned after creating a Gitea release.
type Release struct {
	ID  int64
	URL string
}

// Creator creates releases in Gitea.
type Creator interface {
	CreateRelease(context.Context, Config) (*Release, error)
}

type creator struct {
	client *http.Client
}

// New creates a Gitea release creator.
func New(client *http.Client) Creator {
	if client == nil {
		client = http.DefaultClient
	}

	return &creator{client: client}
}

// ConfigFromEnv loads release configuration from the SemRel subprocess plugin contract.
func ConfigFromEnv(getenv func(string) string) (Config, error) {
	if getenv == nil {
		getenv = func(string) string { return "" }
	}

	draft, err := parseOptionalBool(getenv("SEMREL_PLUGIN_DRAFT"))
	if err != nil {
		return Config{}, fmt.Errorf("SEMREL_PLUGIN_DRAFT: %w", err)
	}

	prerelease, err := parseOptionalBool(firstNonEmpty(getenv("SEMREL_PLUGIN_PRERELEASE"), getenv("SEMREL_IS_PRERELEASE")))
	if err != nil {
		return Config{}, fmt.Errorf("SEMREL_PLUGIN_PRERELEASE: %w", err)
	}

	cfg := Config{
		BaseURL:    strings.TrimSpace(getenv("SEMREL_PLUGIN_BASE_URL")),
		Token:      strings.TrimSpace(getenv("SEMREL_PLUGIN_TOKEN")),
		Owner:      strings.TrimSpace(getenv("SEMREL_PLUGIN_OWNER")),
		Repo:       strings.TrimSpace(getenv("SEMREL_PLUGIN_REPO")),
		TagName:    strings.TrimSpace(firstNonEmpty(getenv("SEMREL_PLUGIN_TAG_NAME"), getenv("SEMREL_TAG_NAME"), getenv("SEMREL_VERSION"), getenv("SEMREL_NEXT_VERSION"))),
		Name:       strings.TrimSpace(firstNonEmpty(getenv("SEMREL_PLUGIN_NAME"), getenv("SEMREL_PLUGIN_TAG_NAME"), getenv("SEMREL_TAG_NAME"), getenv("SEMREL_VERSION"), getenv("SEMREL_NEXT_VERSION"))),
		Body:       getenv("SEMREL_PLUGIN_BODY"),
		Draft:      draft,
		Prerelease: prerelease,
	}

	if cfg.Name == "" {
		cfg.Name = cfg.TagName
	}
	if cfg.Body == "" {
		cfg.Body = getenv("SEMREL_CHANGELOG")
	}

	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// CreateRelease creates a Gitea release via the REST API.
func (c *creator) CreateRelease(ctx context.Context, cfg Config) (*Release, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	endpoint, err := releaseEndpoint(cfg)
	if err != nil {
		return nil, err
	}

	payload := createReleaseRequest{
		TagName:    cfg.TagName,
		Name:       cfg.Name,
		Body:       cfg.Body,
		Draft:      cfg.Draft,
		Prerelease: cfg.Prerelease,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal release request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build release request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "token "+cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create Gitea release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("create Gitea release: unexpected status %s (reading body: %v)", resp.Status, readErr)
		}
		message = bytes.TrimSpace(message)
		if len(message) == 0 {
			return nil, fmt.Errorf("create Gitea release: unexpected status %s", resp.Status)
		}
		return nil, fmt.Errorf("create Gitea release: unexpected status %s: %s", resp.Status, message)
	}

	var created createReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, fmt.Errorf("decode release response: %w", err)
	}

	return &Release{ID: created.ID, URL: firstNonEmpty(created.HTMLURL, created.URL)}, nil
}

type createReleaseRequest struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	Draft      *bool  `json:"draft,omitempty"`
	Prerelease *bool  `json:"prerelease,omitempty"`
}

type createReleaseResponse struct {
	ID      int64  `json:"id"`
	HTMLURL string `json:"html_url"`
	URL     string `json:"url"`
}

func validateConfig(cfg Config) error {
	missing := make([]string, 0, 5)
	if strings.TrimSpace(cfg.BaseURL) == "" {
		missing = append(missing, "SEMREL_PLUGIN_BASE_URL")
	}
	if strings.TrimSpace(cfg.Token) == "" {
		missing = append(missing, "SEMREL_PLUGIN_TOKEN")
	}
	if strings.TrimSpace(cfg.Owner) == "" {
		missing = append(missing, "SEMREL_PLUGIN_OWNER")
	}
	if strings.TrimSpace(cfg.Repo) == "" {
		missing = append(missing, "SEMREL_PLUGIN_REPO")
	}
	if strings.TrimSpace(cfg.TagName) == "" {
		missing = append(missing, "SEMREL_PLUGIN_TAG_NAME or SEMREL_TAG_NAME or SEMREL_VERSION or SEMREL_NEXT_VERSION")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}

	if strings.TrimSpace(cfg.Name) == "" {
		return fmt.Errorf("release name is required")
	}

	return nil
}

func releaseEndpoint(cfg Config) (string, error) {
	baseURL, err := url.Parse(strings.TrimSpace(cfg.BaseURL))
	if err != nil {
		return "", fmt.Errorf("parse SEMREL_PLUGIN_BASE_URL: %w", err)
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return "", fmt.Errorf("SEMREL_PLUGIN_BASE_URL must be an absolute URL")
	}

	baseURL.RawQuery = ""
	baseURL.Fragment = ""
	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + "/api/v1/repos/" + url.PathEscape(cfg.Owner) + "/" + url.PathEscape(cfg.Repo) + "/releases"
	return baseURL.String(), nil
}

func parseOptionalBool(raw string) (*bool, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return nil, nil
	}

	switch value {
	case "1", "true", "yes", "on":
		parsed := true
		return &parsed, nil
	case "0", "false", "no", "off":
		parsed := false
		return &parsed, nil
	default:
		return nil, fmt.Errorf("must be a boolean")
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}

	return ""
}

var _ Creator = (*creator)(nil)
