package gitsync

import "github.com/rs/zerolog"

type GitSyncOption func(*GitSync)

func WithLogger(logger *zerolog.Logger) GitSyncOption {
	return func(g *GitSync) {
		g.log = logger
	}
}

func (g *GitSync) SetLogger(logger *zerolog.Logger) {
	g.log = logger
}

func WithRepoPath(repoPath string) GitSyncOption {
	return func(g *GitSync) {
		g.RepoPath = repoPath
	}
}
