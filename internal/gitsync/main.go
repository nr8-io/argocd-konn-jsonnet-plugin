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

type GitSyncPath struct {
	Name string
	Path string
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

func (g *GitSync) SyncRepos() ([]GitSyncPath, []error) {
	g.log.Debug().Msg("Syncing repos")

	wg := sync.WaitGroup{}
	paths := make([]GitSyncPath, len(g.Repos))

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

func (g *GitSync) SyncRepo(repo string) (GitSyncPath, error) {
	g.log.Debug().Msgf("Syncing repo: %s", repo)

	// allows time for the lock file to be created by another process before we check for it
	time.Sleep(1000 * time.Millisecond)

	// check if repoPath exists
	if _, err := os.Stat(g.RepoPath); os.IsNotExist(err) {
		// create repoPath
		err := os.MkdirAll(g.RepoPath, os.ModePerm)

		if err != nil {
			g.log.Error().Err(err).Msg("Failed to create repo path")

			// Hard fail since we can't create the repo path, so there's no existing repo to fall back to.
			return GitSyncPath{}, fmt.Errorf("failed to create repo path: %w", err)
		}
	}

	// trim any trailing .git from the repo name and use that as the directory name
	repoName := repo
	repoBranchTagOrCommit := "HEAD"

	if strings.Contains(repo, "#") {
		parts := strings.Split(repo, "#")
		repo = parts[0]
		repoBranchTagOrCommit = parts[1]
	}

	if strings.Contains(repo, ".git") {
		repoName = repo[:strings.LastIndex(repo, ".git")]
	}

	repoName = filepath.Base(repoName)
	repoDir := filepath.Join(g.RepoPath, repoName+"_"+strings.ReplaceAll(repoBranchTagOrCommit, "/", "_"))

	// create a lock file to prevent concurrent access
	lockFile := repoDir + ".lock"
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		err = os.WriteFile(lockFile, []byte{}, os.ModePerm)
		if err != nil {
			g.log.Error().Err(err).Msg("Failed to create lock file")
			return GitSyncPath{}, fmt.Errorf("failed to create lock file: %w", err)
		}
		defer os.Remove(lockFile)
	} else {
		g.log.Debug().Msgf("Lock file exists, repo already being synced: %s", repo)

		// wait for lock file to be removed
		for {
			_, err := os.Stat(lockFile)
			if os.IsNotExist(err) {
				break
			}

			// check back every second to see if the lock file has been removed
			time.Sleep(1000 * time.Millisecond)
		}

		g.log.Debug().Msgf("Lock file was removed: %s", repo)

		// return the existing repo path even if we didn't sync it, since another process
		// is already syncing it and we don't want to cause a conflict by trying to sync it again.
		return GitSyncPath{Name: repoName, Path: repoDir}, nil
	}

	_, dirErr := os.Stat(repoDir)
	if os.IsNotExist(dirErr) {
		g.log.Debug().Msgf("Repo does not exist, cloning: %s", repo)

		// clone repo
		cmd := exec.Command("git", "clone", "--depth=1", repo, repoDir)
		err := cmd.Run()
		if err != nil {
			g.log.Error().Err(err).Msgf("Failed to clone repo: %s", repo)

			// Hard fail since the repo doesn't exist and we can't clone it, so there's no existing repo to fall back to.
			return GitSyncPath{}, fmt.Errorf("failed to clone repo %s: %w", repo, err)
		}
	}

	g.log.Debug().Msgf("Repo exists, fetching updates: %s %s", repo, repoBranchTagOrCommit)

	cmd := exec.Command("git", "-C", repoDir, "fetch", "--depth=1", "origin", repoBranchTagOrCommit)
	err := cmd.Run()
	if err != nil {
		g.log.Error().Err(err).Msgf("Failed to fetch repo: %s", repo)

		// return anyway so the plugin can use the existing repo, even if it's not up to date.
		// This is a best effort to avoid complete failure of the plugin if there are issues with the git repo.
		return GitSyncPath{Name: repoName, Path: repoDir}, nil
	}

	cmd = exec.Command("git", "-C", repoDir, "reset", "--hard", "FETCH_HEAD")
	err = cmd.Run()
	if err != nil {
		g.log.Error().Err(err).Msgf("Failed to reset repo: %s", repo)

		// Same as above, return anyway so the plugin can use the existing repo, even if it's not up to date.
		// Prevents an outage of the git repo from causing a complete failure of the plugin.
		return GitSyncPath{Name: repoName, Path: repoDir}, nil
	}

	return GitSyncPath{Name: repoName, Path: repoDir}, nil
}
