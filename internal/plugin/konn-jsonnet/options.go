package konn_jsonnet

import "github.com/rs/zerolog"

type KonnJsonnetPluginOption func(*KonnJsonnetPlugin)

// with external logger
func WithLogger(logger *zerolog.Logger) KonnJsonnetPluginOption {
	return func(p *KonnJsonnetPlugin) {
		p.log = logger
	}
}

func (p *KonnJsonnetPlugin) SetLogger(logger *zerolog.Logger) {
	p.log = logger
}

func WithEntrypoint(entrypoint string) KonnJsonnetPluginOption {
	return func(p *KonnJsonnetPlugin) {
		p.Entrypoint = entrypoint
	}
}

func WithPath(path string) KonnJsonnetPluginOption {
	return func(p *KonnJsonnetPlugin) {
		p.Path = path
	}
}
