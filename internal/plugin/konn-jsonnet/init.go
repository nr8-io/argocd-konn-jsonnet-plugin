package konn_jsonnet

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nr8-io/argocd-konn-jsonnet-plugin/internal/argocd"
	"github.com/nr8-io/argocd-konn-jsonnet-plugin/internal/gitsync"
)

func (p *KonnJsonnetPlugin) Init() error {
	gitRepos := []string{} // git repos
	for _, lib := range p.Libs {
		// if matches https:// or git@ then add to gitRepos
		if len(lib) > 0 && (lib[:8] == "https://" || lib[:4] == "git@") {
			gitRepos = append(gitRepos, lib)
		}
	}

	p.log.Debug().Interface("repos", gitRepos).Msg("Matched git repos from libs")

	// create git sync manager
	sync, err := gitsync.NewGitSync(gitRepos, gitsync.WithLogger(p.log))
	if err != nil {
		p.log.Error().Err(err).Msg("Failed to create git sync manager")
		return err
	}

	// sync all repos
	paths, err := sync.SyncRepos()
	if err != nil {
		p.log.Error().Err(err).Msg("Failed to sync repos")
		return err
	}

	rev := argocd.AppRevisionShort()
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
