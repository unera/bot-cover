package dialog

import "time"

// Option is a function that configures a dialog.
type Option func(d *Dialog)

// WithInactiveTimeout set timeout for stopping the dialog
func WithInactiveTimeout(timeout int) Option {
	return func(d *Dialog) {
		d.inactiveTimeout = timeout
	}
}

// WithHandler set dialog handler
func WithHandler(handler Handler) Option {
	return func(d *Dialog) {
		d.handler = handler
	}
}

// ProfileLoader is a loader for user profile
type ProfileLoader func(botID int64, chatID int64, userID int64, opts ...any) (any, error)

// ProfileStorer is a storer for user profile
type ProfileStorer func(botID int64, chatID int64, userID int64, profile any, opts ...any) error

// WithProfileLoader defines profile loader
func WithProfileLoader(loader ProfileLoader, opts ...any) Option {
	return func(d *Dialog) {
		d.profileLoader = loader
		d.profileLoaderOpts = opts
	}
}

// WithProfileStorer defines profile storer
func WithProfileStorer(storer ProfileStorer, opts ...any) Option {
	return func(d *Dialog) {
		d.profileStorer = storer
		d.profileStorerOpts = opts
	}
}

// WithRateLimit defines how often the bot can send messages
func WithRateLimit(interval time.Duration) Option {
	return func(d *Dialog) {
		d.sendInterval = interval
	}
}
