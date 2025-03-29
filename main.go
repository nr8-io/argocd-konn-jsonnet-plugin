package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nr8-io/argocd-konn-jsonnet-plugin/internal/gitsync"
	"github.com/nr8-io/argocd-konn-jsonnet-plugin/pkg/zino"
	"github.com/rs/zerolog"
)

// Parameter from ArgoCD Config Management Plugin
type Parameter struct {
	Name   string   `json:"name"`
	String string   `json:"string,omitempty"`
	Array  []string `json:"array,omitempty"`
}

// All parameters
type ParameterList []Parameter

// Structured parameters from ArgoCD Config Management Plugin
type Plugin struct {
	log        *zerolog.Logger
	Entrypoint string   `json:"entrypoint"`
	Path       string   `json:"path"`
	ExtVars    []string `json:"extVars"`
	Tlas       []string `json:"tlas"`
	Libs       []string `json:"libs"`
}

func parseParamList(jsonStr string) (ParameterList, error) {
	var result ParameterList

	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func getParamList() (ParameterList, error) {
	params := os.Getenv("ARGOCD_APP_PARAMETERS")
	if params == "" {
		return nil, fmt.Errorf("ARGOCD_APP_PARAMETERS is not set")
	}
	parsed, err := parseParamList(params)
	if err != nil {
		return nil, err
	}

	return parsed, nil
}

func pluginInit(p *Plugin) error {
	gitRepos := []string{} // git repos
	for _, lib := range p.Libs {
		// if matches https:// or git@ then add to gitRepos
		if len(lib) > 0 && (lib[:8] == "https://" || lib[:4] == "git@") {
			gitRepos = append(gitRepos, lib)
		}
	}

	p.log.Debug().Interface("repos", gitRepos).Msg("Matched git repos from libs")

	// create git sync manager
	sync := gitsync.NewGitSync(gitRepos, gitsync.WithLogger(p.log))

	// sync all repos
	paths, err := sync.SyncRepos()
	if err != nil {
		p.log.Error().Err(err).Msg("Failed to sync repos")
		return err
	}

	rev := os.Getenv("ARGOCD_APP_REVISION_SHORT")
	if rev == "" {
		p.log.Error().Msg("ARGOCD_APP_REVISION_SHORT is not set")
		return fmt.Errorf("ARGOCD_APP_REVISION_SHORT is not set")
	}

	// if there are git paths, create a temp dir and symlink them
	if len(paths) > 0 {
		tmpDir := filepath.Join(os.TempDir(), "konn-"+rev)

		p.log.Debug().Msgf("Creating temp repo dir %s", tmpDir)

		// warn if tmpDir already exists
		if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
			p.log.Warn().Msgf("Temp dir %s already exists, removing it...", tmpDir)

			err := os.RemoveAll(tmpDir)
			if err != nil {
				p.log.Error().Err(err).Msgf("Failed to remove temp dir %s", tmpDir)
				return err
			}
		}

		err := os.MkdirAll(tmpDir, os.ModePerm)
		if err != nil {
			p.log.Error().Err(err).Msg("Failed to create temp dir")
			return err
		}

		// create symlinks for each path
		for _, path := range paths {
			p.log.Debug().Msgf("Creating symlink for %s", path)

			err := os.Symlink(path, filepath.Join(tmpDir, filepath.Base(path)))
			if err != nil {
				p.log.Error().Err(err).Msgf("Failed to create symlink for %s", path)
				return err
			}
		}
	}

	return nil
}

func pluginGenerate(p *Plugin) error {
	libs := []string{}     // regular libs
	gitRepos := []string{} // git repos
	for _, lib := range p.Libs {
		// if matches https:// or git@ then add to gitRepos
		if len(lib) > 0 && (lib[:8] == "https://" || lib[:4] == "git@") {
			gitRepos = append(gitRepos, lib)
		} else {
			// else add to libs
			libs = append(libs, lib)
		}
	}

	jsonnetArgs := []string{}

	// default path
	jsonnetArgs = append(jsonnetArgs, "-y")
	jsonnetArgs = append(jsonnetArgs, "--jpath", p.Path)

	// libs
	for _, lib := range libs {
		jsonnetArgs = append(jsonnetArgs, "--jpath", lib)
	}

	if len(gitRepos) > 0 {
		rev := os.Getenv("ARGOCD_APP_REVISION_SHORT")
		if rev == "" {
			p.log.Error().Msg("ARGOCD_APP_REVISION_SHORT is not set")
			return fmt.Errorf("ARGOCD_APP_REVISION_SHORT is not set")
		}

		tmpDir := filepath.Join(os.TempDir(), "konn-"+rev)

		jsonnetArgs = append(jsonnetArgs, "--jpath", tmpDir)
	}

	// external vars
	for _, extStr := range p.ExtVars {
		jsonnetArgs = append(jsonnetArgs, "--ext-str", extStr)
	}

	// top-level arguments
	for _, tlasStr := range p.Tlas {
		jsonnetArgs = append(jsonnetArgs, "--tla-str", tlasStr)
	}

	// file name is path + entrypoint
	fileName := filepath.Join(p.Path, p.Entrypoint)
	jsonnetArgs = append(jsonnetArgs, "./"+fileName)

	// generate the jsonnet file
	cmd := exec.Command("jsonnet", jsonnetArgs...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: init | generate")
		os.Exit(1)
	}

	// get log level from env
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "debug"
	}

	log, err := zino.Init(zino.InitOptions{Level: level})
	if err != nil {
		panic(err)
	}

	list, err := getParamList()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get parameters")
		os.Exit(1)
		return
	}

	// defaults
	params := &Plugin{
		log:        log,
		Entrypoint: "./application.jsonnet",
	}

	for _, param := range list {
		if param.Name == "entrypoint" {
			params.Entrypoint = param.String
		} else if param.Name == "path" {
			params.Path = param.String
		}
		if param.Name == "extVars" {
			params.ExtVars = param.Array
		}
		if param.Name == "tlas" {
			params.Tlas = param.Array
		}
		if param.Name == "libs" {
			params.Libs = param.Array
		}
	}

	if os.Args[1] == "init" {
		log.Debug().Interface("params", params).Msg("Init")

		err := pluginInit(params)
		if err != nil {
			log.Error().Err(err).Msg("Failed to init plugin")
			os.Exit(1)
		}

		return
	}

	if os.Args[1] == "generate" {
		err := pluginGenerate(params)
		if err != nil {
			os.Exit(1)
		}
		return
	}
}
