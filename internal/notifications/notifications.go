package notifications

type Notification struct {
	Title   string
	Content string
}

var NotificationChannel chan Notification

func Init() {
	NotificationChannel = make(chan Notification)
}
