## local testing

```bash
export ARGOCD_APP_REVISION_SHORT=short-rev
export ARGOCD_APP_PARAMETERS='[{"name":"path","string":"./examples/hello-world/staging"},{"array":["app=test"],"name":"extVars"},{"array":["namespace=\"$ARGOCD_APP_REVISION_SHORT\""],"name":"tlasCode"},{"array":["shared", "https://github.com/nr8-io/konn.git","https://github.com/nr8-io/k8s-libsonnet.git", "https://github.com/nr8-io/konn-contrib.git"],"name":"libs"}]'
```

With branch/tag
```bash
export ARGOCD_APP_REVISION_SHORT=short-rev
export ARGOCD_APP_PARAMETERS='[{"name":"path","string":"./examples/hello-world/staging"},{"array":["app=test"],"name":"extVars"},{"array":["namespace=\"$ARGOCD_APP_REVISION_SHORT\""],"name":"tlasCode"},{"array":["shared", "https://github.com/nr8-io/konn.git#tags/0.1.0","https://github.com/nr8-io/k8s-libsonnet.git#gh-pages", "https://github.com/nr8-io/konn-contrib.git"],"name":"libs"}]'
```

## build image
docker buildx build -f Dockerfile -t eu.gcr.io/topvine-co/argocd-konn-jsonnet-plugin ./
docker push eu.gcr.io/topvine-co/argocd-konn-jsonnet-plugin