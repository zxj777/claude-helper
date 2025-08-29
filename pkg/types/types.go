package types

// TextExpanderConfig represents the configuration for text expander
type TextExpanderConfig struct {
	Mappings    map[string]string `json:"mappings"`
	EscapeChar  string           `json:"escape_char,omitempty"`  // Default: "\"
}

// NotificationConfig represents the configuration for task completion notifications
type NotificationConfig struct {
	NotificationTypes []string       `json:"notification_types"` // Types of notifications: "audio", "desktop"
	CooldownSecs      int            `json:"cooldown_seconds"`   // Cooldown period to prevent frequent notifications
	Desktop           DesktopConfig  `json:"desktop"`            // Desktop notification settings
	Audio             AudioConfig    `json:"audio"`              // Audio notification settings
}

// DesktopConfig represents desktop notification settings
type DesktopConfig struct {
	Enabled     bool `json:"enabled"`      // Whether desktop notifications are enabled
	ShowDetails bool `json:"show_details"` // Show detailed information in notifications
}

// AudioConfig represents audio notification settings
type AudioConfig struct {
	Enabled      bool   `json:"enabled"`       // Whether audio notifications are enabled
	SuccessSound string `json:"success_sound"` // Sound file for successful operations
	ErrorSound   string `json:"error_sound"`   // Sound file for failed operations
	DefaultSound string `json:"default_sound"` // Default sound file for general completions
	Volume       int    `json:"volume"`        // Volume level (0-100)
}

// TODO: Define other core data structures for agents, hooks, and templates as needed