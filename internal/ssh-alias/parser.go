package sshalias

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	aliasStartPrefix = "# >>> ssh-alias "
	aliasStartSuffix = " >>>"
	aliasEndPrefix   = "# <<< ssh-alias "
	aliasEndSuffix   = " <<<"
)

func parseAliases(content string) (map[string]Alias, error) {
	aliases := make(map[string]Alias)
	lines := strings.Split(content, "\n")

	inBlock := false
	var startLine int
	var markerAlias string
	var hostAlias string
	var hostname string
	var user string
	port := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if alias, ok := parseMarkerStart(trimmed); ok {
			if inBlock {
				return nil, fmt.Errorf("nested ssh-alias block at line %d", i+1)
			}

			inBlock = true
			startLine = i + 1
			markerAlias = alias
			hostAlias = ""
			hostname = ""
			user = ""
			port = 0
			continue
		}

		if alias, ok := parseMarkerEnd(trimmed); ok {
			if !inBlock {
				return nil, fmt.Errorf("ssh-alias end marker without start at line %d", i+1)
			}
			if alias != markerAlias {
				return nil, fmt.Errorf("mismatched ssh-alias marker alias at line %d", i+1)
			}

			if hostAlias == "" {
				hostAlias = markerAlias
			}

			sshAlias, err := NewAlias(hostAlias, hostname, user, port)
			if err != nil {
				return nil, fmt.Errorf("invalid ssh-alias block starting at line %d: %w", startLine, err)
			}

			if sshAlias.alias != markerAlias {
				return nil, fmt.Errorf("host alias differs from marker alias at line %d", startLine)
			}

			if _, exists := aliases[sshAlias.alias]; exists {
				return nil, fmt.Errorf("duplicate ssh-alias %q at line %d", sshAlias.alias, startLine)
			}

			aliases[sshAlias.alias] = sshAlias
			inBlock = false
			continue
		}

		if !inBlock {
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "Host":
			hostAlias = fields[1]
		case "HostName":
			hostname = fields[1]
		case "User":
			user = fields[1]
		case "Port":
			parsedPort, err := strconv.Atoi(fields[1])
			if err != nil {
				return nil, fmt.Errorf("invalid port at line %d: %w", i+1, err)
			}
			port = parsedPort
		}
	}

	if inBlock {
		return nil, fmt.Errorf("unterminated ssh-alias block starting at line %d", startLine)
	}

	return aliases, nil
}

func parseMarkerStart(line string) (string, bool) {
	alias, ok := parseMarker(line, aliasStartPrefix, aliasStartSuffix)
	return alias, ok
}

func parseMarkerEnd(line string) (string, bool) {
	alias, ok := parseMarker(line, aliasEndPrefix, aliasEndSuffix)
	return alias, ok
}

func parseMarker(line, prefix, suffix string) (string, bool) {
	if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, suffix) {
		return "", false
	}

	alias := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, prefix), suffix))
	if alias == "" {
		return "", false
	}

	return alias, true
}
