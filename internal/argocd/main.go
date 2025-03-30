package argocd

import (
	"encoding/json"
	"fmt"
	"os"
)

// All parameters
type ParametersList []Parameter

// Parameter from ArgoCD Config Management Plugin
type Parameter struct {
	Name           string            `json:"name"`
	String         string            `json:"string,omitempty"`
	Array          []string          `json:"array,omitempty"`
	Map            map[string]string `json:"map,omitempty"`
	CollectionType string            `json:"collectionType,omitempty"`
}

// get argo app params from env
func AppParameters() (ParametersList, error) {
	params := os.Getenv("ARGOCD_APP_PARAMETERS")
	if params == "" {
		return nil, fmt.Errorf("couldn't not find argo app parameters in ARGOCD_APP_PARAMETERS")
	}

	var result ParametersList

	err := json.Unmarshal([]byte(params), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal argo app parameters: %w", err)
	}
	return result, nil
}

func AppRevisionShort() string {
	return os.Getenv("ARGOCD_APP_REVISION_SHORT")
}
