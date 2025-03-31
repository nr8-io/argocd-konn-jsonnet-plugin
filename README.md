# argocd-konn-jsonnet-plugin

ArgoCD [config management
plugin](https://argo-cd.readthedocs.io/en/stable/operator-manual/config-management-plugins/)
to extend the basic functionality of the jsonnet plugin intended for use with
[konn](https://github.com/nr8-io/konn) to support external libraries by using
git repos.

The plugin provides the same basic functionality as the standard [jsonnet
plugin](https://argo-cd.readthedocs.io/en/stable/user-guide/jsonnet/) but allows
supplied libs to be git repos which the plugin will clone and add to the jsonnet
library search path automatically before generating the configs.

## Installation
```bash
kubectl apply -n argocd -f config/konn-jsonnet-plugin-cm.yaml
kubectl patch -n argocd cm/argocd-cmd-params-cm --patch-file=config/argocd-cmd-params-cm-patch.yaml
kubectl patch -n argocd deployment/argocd-repo-server --patch-file=config/argocd-repo-server-deploy-patch.yaml
```

## Usage
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: konn-jsonnet
  annotations:
    argocd.argoproj.io/manifest-generate-paths: "."
spec:
  project: default
  source:
    repoURL: https://github.com/nr8-io/argocd-konn-jsonnet-plugin.git
    path: examples/hello-world
    targetRevision: HEAD
    plugin:
      name: konn-jsonnet
      parameters:
        - name: path
          string: ./staging
        - name: entrypoint
          string: ./application.jsonnet
        - name: extVars
          array:
            - app=hello-world-stg
        - name: tlas
          array:
            - namespace=hello-world-stg
        - name: libs
          array:
            - https://github.com/nr8-io/konn.git
            - https://github.com/nr8-io/k8s-libsonnet.git
  destination:
    server: https://kubernetes.default.svc
    namespace: hello-world-stg
```

### Parameters

#### path - jsonnet root path (optional)

- This is the path to the jsonnet directory relative to the source path that
  will act as the jsonnet root.
- Typically this is set to the name of the environment. eg. ./staging
- Defaults to './' unless otherwise specified.

#### entrypoint - jsonnet entrypoint (optional)

- The main jsonnet file to evaluate, only one entrypoint is allowed, and it must
  be a jsonnet file. 
- Defaults to './application.jsonnet' unless otherwise specified.

#### extVars - jsonnet external variables (optional)

extVars are supported but recommended, consider using tlas instead.

- External variables to pass to jsonnet. This is a list of key=value pairs
  passed as --ext-str.
- The key must be a valid jsonnet variable name. eg. app=konn-jsonnet
- Defaults to an empty list unless otherwise specified.

#### tlas - jsonnet top level arguments (optional)

- Top level arguments to pass to jsonnet. This is a list of key=value pairs
  passed as --tla-str.
- The key must be a valid jsonnet variable name. eg. ns=konn-jsonnet
- Defaults to an empty list unless otherwise specified.

#### libs - jsonnet libraries (optional)

- Specify a list of jsonnet libraries to include in the jsonnet library search
  dir passed as --jpath.
- If a git repo is specified, it will be cloned and added to the jsonnet library
  search dir.
- Git repos are matched by the value of the library. eg. starts with git@ or
  https://
- Git repos library names match the name of the repo excluding the .git suffix.
- Defaults to an empty list unless otherwise specified.