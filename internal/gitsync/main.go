package gitsync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
			repoPath = "/tmp/_konn_jsonnet_repos"
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

func (g *GitSync) SyncRepos() ([]string, error) {
	g.log.Debug().Msg("Syncing repos")

	paths := make([]string, len(g.Repos))
	for i, repo := range g.Repos {
		path, err := g.SyncRepo(repo)
		if err != nil {
			return nil, err
		}
		paths[i] = path
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
			return "", err
		}
	}

	repoName := strings.TrimSuffix(filepath.Base(repo), ".git")
	repoDir := filepath.Join(g.RepoPath, repoName)

	_, err := os.Stat(repoDir)
	if os.IsNotExist(err) {
		g.log.Debug().Msgf("Repo does not exist, cloning: %s", repo)

		// clone repo
		cmd := exec.Command("git", "clone", repo, repoDir)
		err := cmd.Run()
		if err != nil {
			g.log.Error().Err(err).Msg("Failed to clone repo")
			return "", err
		}
	} else {
		g.log.Debug().Msgf("Repo exists, pulling: %s", repo)

		cmd := exec.Command("git", "-C", repoDir, "pull")
		err := cmd.Run()
		if err != nil {
			g.log.Error().Err(err).Msg("Failed to pull repo")
			return "", err
		}
	}

	g.log.Debug().Msgf("Repo exists, checking out latest commit: %s", repo)

	// check out the latest commit
	cmd := exec.Command("git", "-C", repoDir, "checkout", "HEAD")
	err = cmd.Run()
	if err != nil {
		g.log.Error().Err(err).Msg("Failed to checkout repo")
		return "", err
	}

	return repoDir, nil
}
