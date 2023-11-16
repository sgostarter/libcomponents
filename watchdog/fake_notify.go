package watchdog

func NewFakeNotify() INotify {
	return &fakeNotifyImpl{}
}

type fakeNotifyImpl struct {
}

func (impl *fakeNotifyImpl) NotifyTimeout(_ string) {

}
