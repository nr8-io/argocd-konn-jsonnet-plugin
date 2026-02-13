package konn_jsonnet

import (
	"fmt"

	"github.com/nr8-io/argocd-konn-jsonnet-plugin/internal/argocd"
	"github.com/nr8-io/argocd-konn-jsonnet-plugin/pkg/zino"
	"github.com/rs/zerolog"
)

// Structured parameters from ArgoCD Config Management KonnJsonnetPlugin
type KonnJsonnetPlugin struct {
	// logger
	log *zerolog.Logger

	// jsonnet parameters
	Entrypoint  string   `json:"entrypoint"`
	Path        string   `json:"path"`
	ExtVars     []string `json:"extVars"`
	ExtVarsCode []string `json:"extVarsCode"`
	Tlas        []string `json:"tlas"`
	TlasCode    []string `json:"tlasCode"`
	Libs        []string `json:"libs"`
}

func NewKonnJsonnetPlugin(options ...KonnJsonnetPluginOption) (*KonnJsonnetPlugin, error) {
	plugin := &KonnJsonnetPlugin{
		Entrypoint: "./application.jsonnet",
		Path:       "./",
	}

	err := plugin.Configure(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to configure plugin: %w", err)
	}

	return plugin, nil
}

// configure the plugin with functional options and set defaults
func (p *KonnJsonnetPlugin) Configure(options ...KonnJsonnetPluginOption) error {
	// apply options
	for _, option := range options {
		option(p)
	}

	// set default logger if not set
	if p.log == nil {
		log, err := zino.NewLogger("debug")
		if err != nil {
			return err
		}
		p.log = log
	}

	// get argo app params from env
	params, err := argocd.AppParameters()
	if err != nil {
		return err
	}

	// set plugin options from argo app parameters
	for _, param := range params {
		switch param.Name {
		case "entrypoint":
			p.Entrypoint = param.String
		case "path":
			p.Path = param.String
		case "extVars":
			p.ExtVars = param.Array
		case "extVarsCode":
			p.ExtVarsCode = param.Array
		case "tlas":
			p.Tlas = param.Array
		case "tlasCode":
			p.TlasCode = param.Array
		case "libs":
			p.Libs = param.Array
		default:
			// ignore unknown parameters
			continue
		}
	}

	return nil
}
