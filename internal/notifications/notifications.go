// Package notifications provides primitives to initialise the notification channel.
package notifications

// Notification struct
type Notification struct {
	Title   string
	Content string
}

var NotificationChannel chan Notification

// Init initialises the notification channel which can be used to push notifications
// to the user.
func Init() {
	NotificationChannel = make(chan Notification)
}
