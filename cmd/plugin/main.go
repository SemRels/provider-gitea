// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package main

import (
	"log"

	plugin "github.com/SemRels/provider-gitea/internal/plugin"
)

func main() {
	client := plugin.NewClient(plugin.Config{})
	log.Printf("provider-gitea plugin ready: creates Gitea releases and Go registry metadata (%T)", client)
}
