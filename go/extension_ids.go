package main

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
)

type manifestOrigins struct {
	AllowedOrigins []string `json:"allowed_origins"`
}

func buildAllowedOrigins(manifestPath string, extensionID string) []string {
	ids := map[string]struct{}{}

	addID := func(id string) {
		clean := strings.TrimSpace(id)
		if clean == "" {
			return
		}
		ids[clean] = struct{}{}
	}

	addID(extensionID)

	payload, err := os.ReadFile(manifestPath)
	if err == nil && len(payload) > 0 {
		var old manifestOrigins
		if json.Unmarshal(payload, &old) == nil {
			for _, origin := range old.AllowedOrigins {
				addID(extensionIDFromOrigin(origin))
			}
		}
	}

	idList := make([]string, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	sort.Strings(idList)

	origins := make([]string, 0, len(idList))
	for _, id := range idList {
		origins = append(origins, "chrome-extension://"+id+"/")
	}
	return origins
}

func extensionIDFromOrigin(origin string) string {
	const prefix = "chrome-extension://"
	if !strings.HasPrefix(origin, prefix) {
		return ""
	}
	id := strings.TrimSuffix(strings.TrimPrefix(origin, prefix), "/")
	if id == "" {
		return ""
	}
	return id
}
