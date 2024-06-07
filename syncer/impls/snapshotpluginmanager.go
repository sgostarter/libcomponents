package impls

import (
	"sync"

	"github.com/sgostarter/libcomponents/syncer"
)

type SnapshotPluginStorageGenerator func(lastData *syncer.SnapshotData) (cache syncer.SnapshotPluginCache, hasData bool)

func NewSnapshotPluginStorageManager(generators map[string]SnapshotPluginStorageGenerator, lastSnapshotData *syncer.SnapshotData) syncer.SnapshotPluginCacheManager {
	if generators == nil {
		return nil
	}

	impl := &snapshotPluginStorageManagerImpl{
		plugins:    make(map[string]syncer.SnapshotPluginCache),
		generators: generators,
	}

	impl.init(lastSnapshotData)

	return impl
}

type snapshotPluginStorageManagerImpl struct {
	pluginsLock sync.Mutex
	plugins     map[string]syncer.SnapshotPluginCache
	generators  map[string]SnapshotPluginStorageGenerator
}

func (impl *snapshotPluginStorageManagerImpl) init(lastSnapshotData *syncer.SnapshotData) {
	for pluginID, generator := range impl.generators {
		plugin, hasData := generator(lastSnapshotData)
		if plugin == nil {
			continue
		}

		if !hasData {
			continue
		}

		impl.plugins[pluginID] = plugin
	}
}

func (impl *snapshotPluginStorageManagerImpl) GetCache(id string) (plugin syncer.SnapshotPluginCache, err error) {
	impl.pluginsLock.Lock()
	defer impl.pluginsLock.Unlock()

	plugin, ok := impl.plugins[id]
	if ok {
		return
	}

	generator, ok := impl.generators[id]
	if !ok || generator == nil {
		return
	}

	plugin, _ = generator(nil)
	if plugin == nil {
		return
	}

	impl.plugins[id] = plugin

	return
}

func (impl *snapshotPluginStorageManagerImpl) GetCaches4Save() (plugins []syncer.SnapshotPluginCache, _ error) {
	impl.pluginsLock.Lock()
	defer impl.pluginsLock.Unlock()

	for _, storage := range impl.plugins {
		plugins = append(plugins, storage)
	}

	return
}
