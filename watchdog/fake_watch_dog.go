package watchdog

func NewFakeWatchDog() WatchDog {
	return &fakeWatchDogImpl{}
}

type fakeWatchDogImpl struct {
}

func (impl *fakeWatchDogImpl) Touch() {

}

func (impl *fakeWatchDogImpl) Start() {

}
func (impl *fakeWatchDogImpl) Stop() {

}

func (impl *fakeWatchDogImpl) Started() bool {
	return false
}
