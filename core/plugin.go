package core

import "context"

type Plugin interface {
	Init(args ...any) error
}

type PluginFactory func(ctx context.Context, args ...any) Plugin

var gPluginFactoryMap = make(map[string]PluginFactory)

func RegisterPlugin(name string, f PluginFactory) {
	if _, exists := gPluginFactoryMap[name]; exists {
		panic("plugin already registered: " + name)
	}
	gPluginFactoryMap[name] = f
}
