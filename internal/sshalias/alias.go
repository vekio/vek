package sshalias

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const defaultSSHPort = 22
const maxSSHPort = 65535

// Alias is a validated, vek-managed SSH host entry.
type Alias struct {
	alias    string
	hostname string
	user     string
	port     int
}

// Name returns the SSH host alias.
func (a Alias) Name() string {
	return a.alias
}

// Hostname returns the remote hostname or IP address.
func (a Alias) Hostname() string {
	return a.hostname
}

// User returns the SSH username.
func (a Alias) User() string {
	return a.user
}

// Port returns the SSH port.
func (a Alias) Port() int {
	return a.port
}

// NewAlias validates and constructs an Alias. A zero port defaults to 22.
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

// Block renders the managed OpenSSH configuration block for the alias.
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
	return fmt.Sprintf("# >>> vek-ssh-alias %s >>>", a.alias)
}

func (a Alias) markerEnd() string {
	return fmt.Sprintf("# <<< vek-ssh-alias %s <<<", a.alias)
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
