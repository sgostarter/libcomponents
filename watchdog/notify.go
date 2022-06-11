package watchdog

type INotify interface {
	NotifyTimeout(msg string)
}
