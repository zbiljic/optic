package optic

// Plugin is a base interface all Optic plugins must satisfy.
type Plugin interface {
	// Unique name of the plugin.
	Kind() string

	// Description returns a one-sentence description on the plugin.
	Description() string
}
