package config

import (
	"fmt"
	"strings"
)

func (c Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port outside range")
	}
	if strings.TrimSpace(c.DatabasePath) == "" {
		return fmt.Errorf("database path is required")
	}
	if c.BusinessZone == nil {
		return fmt.Errorf("business timezone is required")
	}
	if c.SessionTTL <= 0 {
		return fmt.Errorf("session ttl must be positive")
	}
	if c.ShutdownGrace <= 0 {
		return fmt.Errorf("shutdown grace must be positive")
	}
	return nil
}
func (c Config) Address() string { return fmt.Sprintf(":%d", c.Port) }
