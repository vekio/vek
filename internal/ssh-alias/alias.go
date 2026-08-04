package sshalias

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const defaultSSHPort = 22
const maxSSHPort = 65535

type Alias struct {
	alias    string
	hostname string
	user     string
	port     int
}

func (a Alias) Name() string {
	return a.alias
}

func (a Alias) Hostname() string {
	return a.hostname
}

func (a Alias) User() string {
	return a.user
}

func (a Alias) Port() int {
	return a.port
}

func NewAlias(alias, hostname, user string, port int) (Alias, error) {
	alias = strings.TrimSpace(alias)
	hostname = strings.TrimSpace(hostname)
	user = strings.TrimSpace(user)

	if err := validateSSHToken("alias", alias); err != nil {
		return Alias{}, err
	}

	if err := validateSSHToken("hostname", hostname); err != nil {
		return Alias{}, err
	}

	if err := validateSSHToken("user", user); err != nil {
		return Alias{}, err
	}

	if port == 0 {
		port = defaultSSHPort
	}

	if port < 1 || port > maxSSHPort {
		return Alias{}, fmt.Errorf("port must be between 1 and %d", maxSSHPort)
	}

	return Alias{
		alias:    alias,
		hostname: hostname,
		user:     user,
		port:     port,
	}, nil
}

func (a Alias) Block() string {
	var builder strings.Builder
	builder.Grow(len(a.alias) + len(a.hostname) + len(a.user) + 96)

	builder.WriteString(a.markerStart())
	builder.WriteString("\n")
	builder.WriteString("Host ")
	builder.WriteString(a.alias)
	builder.WriteString("\n")
	builder.WriteString("\tHostName ")
	builder.WriteString(a.hostname)
	builder.WriteString("\n")
	builder.WriteString("\tUser ")
	builder.WriteString(a.user)
	builder.WriteString("\n")
	builder.WriteString("\tPort ")
	builder.WriteString(strconv.Itoa(a.port))
	builder.WriteString("\n")
	builder.WriteString(a.markerEnd())
	builder.WriteString("\n")

	return builder.String()
}

func (a Alias) markerStart() string {
	return fmt.Sprintf("# >>> ssh-alias %s >>>", a.alias)
}

func (a Alias) markerEnd() string {
	return fmt.Sprintf("# <<< ssh-alias %s <<<", a.alias)
}

func validateSSHToken(name, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", name)
	}

	for _, r := range value {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return fmt.Errorf("%s must not contain whitespace or control characters", name)
		}
	}

	return nil
}
