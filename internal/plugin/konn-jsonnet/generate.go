package konn_jsonnet

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nr8-io/argocd-konn-jsonnet-plugin/internal/argocd"
)

func (p *KonnJsonnetPlugin) Generate() error {
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

	// top-level arguments
	for _, tlasStr := range p.Tlas {
		jsonnetArgs = append(jsonnetArgs, "--tla-str", os.ExpandEnv(tlasStr))
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
