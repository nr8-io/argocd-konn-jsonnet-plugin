package konn_jsonnet

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nr8-io/argocd-konn-jsonnet-plugin/internal/argocd"
	"github.com/nr8-io/argocd-konn-jsonnet-plugin/internal/gitsync"
)

func (p *KonnJsonnetPlugin) Generate() error {
	libs := []string{}     // regular libs
	gitRepos := []string{} // git repos
	for _, lib := range p.Libs {
		// if matches https:// or git@ then add to gitRepos
		if (len(lib) >= 8 && lib[:8] == "https://") || (len(lib) >= 4 && lib[:4] == "git@") {
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
		sync, err := gitsync.NewGitSync(gitRepos, gitsync.WithLogger(p.log))
		if err != nil {
			p.log.Error().Err(err).Msg("Failed to create git sync manager")
			return err
		}

		_, errs := sync.SyncRepos()
		if len(errs) > 0 {
			p.log.Error().Err(err).Msg("Failed to sync repos")
			return fmt.Errorf("failed to sync repos: %v", errs)
		}

		rev := argocd.AppRevisionShort()
		if rev == "" {
			p.log.Error().Msg("ARGOCD_APP_REVISION_SHORT is not set")
			return fmt.Errorf("ARGOCD_APP_REVISION_SHORT is not set")
		}

		tmpDir := filepath.Join(os.TempDir(), "konn-"+rev)

		jsonnetArgs = append(jsonnetArgs, "--jpath", tmpDir)
	}

	// external vars
	for _, extStr := range p.ExtVars {
		jsonnetArgs = append(jsonnetArgs, "--ext-str", os.ExpandEnv(extStr))
	}

	for _, extCode := range p.ExtVarsCode {
		jsonnetArgs = append(jsonnetArgs, "--ext-code", os.ExpandEnv(extCode))
	}

	// top-level arguments
	for _, tlasStr := range p.Tlas {
		jsonnetArgs = append(jsonnetArgs, "--tla-str", os.ExpandEnv(tlasStr))
	}

	for _, tlasCode := range p.TlasCode {
		jsonnetArgs = append(jsonnetArgs, "--tla-code", os.ExpandEnv(tlasCode))
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
		return fmt.Errorf("failed to run jsonnet with %s: %w", jsonnetArgs, err)
	}

	return nil
}
