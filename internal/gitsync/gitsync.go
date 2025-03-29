package gitsync

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nr8-io/argocd-konn-jsonnet-plugin/pkg/zino"
	"github.com/rs/zerolog"
)

type GitSync struct {
	log      *zerolog.Logger
	repos    []string
	repoPath string
}

type GitSyncOption func(*GitSync)

func WithLogger(logger *zerolog.Logger) GitSyncOption {
	return func(g *GitSync) {
		g.log = logger
	}
}

func NewGitSync(repos []string, options ...GitSyncOption) *GitSync {
	repoPath := os.Getenv("REPO_PATH")

	if repoPath == "" {
		repoPath = "/tmp/_konn_jsonnet_repos"
	}

	gitSync := &GitSync{
		repoPath: repoPath,
		repos:    repos,
	}

	// needs logger
	if gitSync.log == nil {
		log, err := zino.Init(zino.InitOptions{Level: "debug"})

		if err != nil {
			panic(err)
		}

		gitSync.log = log
	}

	for _, opt := range options {
		opt(gitSync)
	}

	return gitSync
}

func (gs *GitSync) SyncRepos() ([]string, error) {
	gs.log.Debug().Msg("Syncing repos")

	paths := make([]string, len(gs.repos))
	for i, repo := range gs.repos {
		path, err := gs.SyncRepo(repo)
		if err != nil {
			return nil, err
		}
		paths[i] = path
	}

	return paths, nil
}

func (gs *GitSync) SyncRepo(repo string) (string, error) {
	gs.log.Debug().Msgf("Syncing repo: %s", repo)

	// check if repoPath exists
	if _, err := os.Stat(gs.repoPath); os.IsNotExist(err) {
		// create repoPath
		err := os.MkdirAll(gs.repoPath, os.ModePerm)

		if err != nil {
			gs.log.Error().Err(err).Msg("Failed to create repo path")
			return "", err
		}
	}

	repoName := strings.TrimSuffix(filepath.Base(repo), ".git")
	repoDir := filepath.Join(gs.repoPath, repoName)

	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		gs.log.Debug().Msgf("Repo does not exist, cloning: %s", repo)

		// clone repo
		cmd := exec.Command("git", "clone", repo, repoDir)
		err := cmd.Run()
		if err != nil {
			gs.log.Error().Err(err).Msg("Failed to clone repo")
			return "", err
		}
	} else {
		gs.log.Debug().Msgf("Repo exists, pulling: %s", repo)

		cmd := exec.Command("git", "-C", repoDir, "pull")
		err := cmd.Run()
		if err != nil {
			gs.log.Error().Err(err).Msg("Failed to pull repo")
			return "", err
		}
	}

	gs.log.Debug().Msgf("Repo exists, checking out latest commit: %s", repo)

	// check out the latest commit
	cmd := exec.Command("git", "-C", repoDir, "checkout", "HEAD")
	err := cmd.Run()
	if err != nil {
		gs.log.Error().Err(err).Msg("Failed to checkout repo")
		return "", err
	}

	return repoDir, nil
}
