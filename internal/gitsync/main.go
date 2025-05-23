package gitsync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nr8-io/argocd-konn-jsonnet-plugin/pkg/zino"
	"github.com/rs/zerolog"
)

type GitSync struct {
	log *zerolog.Logger

	// git parameters
	Repos    []string
	RepoPath string
}

// Manage git repositories used in libs provided by the plugin
func NewGitSync(repos []string, options ...GitSyncOption) (*GitSync, error) {
	g := &GitSync{
		Repos: repos,
	}

	err := g.Configure(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to configure plugin: %w", err)
	}

	return g, nil
}

func (g *GitSync) Configure(options ...GitSyncOption) error {
	for _, option := range options {
		option(g)
	}

	// set default repo path if not set
	if g.RepoPath == "" {
		repoPath := os.Getenv("REPO_PATH")
		if repoPath == "" {
			repoPath = "/tmp/konn-jsonnet-plugin-repos"
		}
		g.RepoPath = repoPath
	}

	// needs logger
	if g.log == nil {
		log, err := zino.NewLogger("debug")
		if err != nil {
			return err
		}
		g.log = log
	}

	return nil
}

func (g *GitSync) SyncRepos() ([]string, []error) {
	g.log.Debug().Msg("Syncing repos")

	wg := sync.WaitGroup{}
	paths := make([]string, len(g.Repos))

	var errs []error
	for i, repo := range g.Repos {
		wg.Add(1)
		go func() {
			defer wg.Done()
			path, err := g.SyncRepo(repo)

			if err != nil {
				errs = append(errs, err)
			} else {
				paths[i] = path
			}
		}()
	}

	// wait for all goroutines to finish
	wg.Wait()
	if len(errs) > 0 {
		g.log.Error().Errs("errs", errs).Msg("Failed to sync repos")
		return nil, errs
	}

	return paths, nil
}

func (g *GitSync) SyncRepo(repo string) (string, error) {
	g.log.Debug().Msgf("Syncing repo: %s", repo)

	// check if repoPath exists
	if _, err := os.Stat(g.RepoPath); os.IsNotExist(err) {
		// create repoPath
		err := os.MkdirAll(g.RepoPath, os.ModePerm)

		if err != nil {
			g.log.Error().Err(err).Msg("Failed to create repo path")
			return "", fmt.Errorf("failed to create repo path: %w", err)
		}
	}

	repoName := strings.TrimSuffix(filepath.Base(repo), ".git")
	repoDir := filepath.Join(g.RepoPath, repoName)

	// create a lock file to prevent concurrent access
	lockFile := repoDir + ".lock"
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		err = os.WriteFile(lockFile, []byte{}, os.ModePerm)
		if err != nil {
			g.log.Error().Err(err).Msg("Failed to create lock file")
			return "", fmt.Errorf("failed to create lock file: %w", err)
		}
		defer os.Remove(lockFile)
	} else {
		g.log.Debug().Msgf("Lock file exists: %s", repo)

		// wait for lock file to be removed
		for {
			_, err := os.Stat(lockFile)
			if os.IsNotExist(err) {
				break
			}

			time.Sleep(100 * time.Millisecond)
		}

		g.log.Debug().Msgf("Lock file was removed: %s", repo)

		return repoDir, nil
	}

	_, dirErr := os.Stat(repoDir)
	if os.IsNotExist(dirErr) {
		g.log.Debug().Msgf("Repo does not exist, cloning: %s", repo)

		// clone repo
		cmd := exec.Command("git", "clone", repo, repoDir)
		err := cmd.Run()
		if err != nil {
			g.log.Error().Err(err).Msg("Failed to clone repo")
			return "", fmt.Errorf("failed to clone repo %s: %w", repo, err)
		}
	} else {
		g.log.Debug().Msgf("Repo exists, pulling: %s", repo)

		cmd := exec.Command("git", "-C", repoDir, "pull")
		err := cmd.Run()
		if err != nil {
			g.log.Error().Err(err).Msg("Failed to pull repo")
			return "", fmt.Errorf("failed to pull repo %s: %w", repo, err)
		}
	}

	g.log.Debug().Msgf("Repo exists, checking out latest commit: %s", repo)

	cmd := exec.Command("git", "-C", repoDir, "checkout", "HEAD")
	err := cmd.Run()
	if err != nil {
		g.log.Error().Err(err).Msg("Failed to checkout repo")
		return "", fmt.Errorf("failed to checkout repo %s: %w", repo, err)
	}

	return repoDir, nil
}
