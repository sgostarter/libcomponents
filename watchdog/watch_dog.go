package watchdog

import (
	"sync"
	"time"
)

type WatchDog interface {
	Touch()

	Start()
	Stop()
	Started() bool
}

type Config struct {
	CheckInterval time.Duration

	CheckMaxDuration time.Duration
	CheckFailCount   int
}

func NewWatchDog(cfg Config, notify INotify) WatchDog {
	if notify == nil {
		return nil
	}

	impl := &watchDogImpl{
		cfg:         cfg,
		notify:      notify,
		lastTouchAt: time.Now(),
	}

	impl.init()

	return impl
}

type watchDogImpl struct {
	cfg    Config
	notify INotify

	lock        sync.Mutex
	started     bool
	failCount   int
	lastTouchAt time.Time
}

func (impl *watchDogImpl) Touch() {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	impl.lastTouchAt = time.Now()
}

func (impl *watchDogImpl) Start() {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	impl.started = true
	impl.lastTouchAt = time.Now()
	impl.failCount = 0
}

func (impl *watchDogImpl) Stop() {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	impl.started = false
}

func (impl *watchDogImpl) Started() bool {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	return impl.started
}

func (impl *watchDogImpl) init() {
	if impl.cfg.CheckInterval <= 0 {
		impl.cfg.CheckInterval = time.Second * 20
	}

	if impl.cfg.CheckMaxDuration <= 0 {
		impl.cfg.CheckMaxDuration = time.Minute
	}

	if impl.cfg.CheckFailCount <= 0 {
		impl.cfg.CheckFailCount = 1
	}

	go impl.mainRoutine()
}

func (impl *watchDogImpl) check() bool {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	return time.Since(impl.lastTouchAt) < impl.cfg.CheckMaxDuration
}

func (impl *watchDogImpl) mainRoutine() {
	for {
		time.Sleep(impl.cfg.CheckInterval)

		if !impl.started {
			continue
		}

		if impl.check() {
			impl.failCount = 0
		} else {
			impl.failCount++
		}

		if impl.failCount >= impl.cfg.CheckFailCount {
			impl.notify.NotifyTimeout("")

			impl.failCount = 0
			impl.lastTouchAt = time.Now()
		}
	}
}
