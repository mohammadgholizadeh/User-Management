package domain

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Email string

func (e Email) Validate() error {
	if len(e) < 3 || len(e) > 254 {
		zap.L().Warn("Email length out of bounds",
			zap.String("entity", "email"),
			zap.String("value", string(e)),
			zap.Int("length", len(e)),
		)

		return fmt.Errorf("email length must be between 3 and 254 characters")
	}

	if !emailRegex.MatchString(string(e)) {
		zap.L().Warn("Email regex mismatch",
			zap.String("entity", "email"),
			zap.String("value", string(e)),
		)

		return fmt.Errorf("email format is invalid")
	}

	parts := strings.Split(string(e), "@")
	mx, err := net.LookupMX(parts[1])
	if err != nil || len(mx) == 0 {
		zap.L().Warn("Email domain has no MX records",
			zap.String("entity", "email"),
			zap.String("value", string(e)),
			zap.Error(err),
		)

		return fmt.Errorf("email domain has no valid mail servers")
	}

	return nil
}
