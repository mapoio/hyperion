package hyperion

// noopConfig is a no-op implementation of Config interface.
type noopConfig struct{}

// NewNoOpConfig creates a new no-op Config implementation.
func NewNoOpConfig() Config {
	return &noopConfig{}
}

func (c *noopConfig) Unmarshal(key string, rawVal any) error { return nil }
func (c *noopConfig) Get(key string) any                     { return nil }
func (c *noopConfig) GetString(key string) string            { return "" }
func (c *noopConfig) GetInt(key string) int                  { return 0 }
func (c *noopConfig) GetInt64(key string) int64              { return 0 }
func (c *noopConfig) GetBool(key string) bool                { return false }
func (c *noopConfig) GetFloat64(key string) float64          { return 0.0 }
func (c *noopConfig) GetStringSlice(key string) []string     { return nil }
func (c *noopConfig) IsSet(key string) bool                  { return false }
func (c *noopConfig) AllKeys() []string                      { return nil }
