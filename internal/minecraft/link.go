package minecraft

import "strings"

type Link struct {
	UUID     string
	Username string
}

func normalizeUUID(uuid string) string {
	return strings.ToLower(
		strings.TrimSpace(uuid),
	)
}
