## local testing

```bash
export ARGOCD_APP_REVISION_SHORT=short-rev
export ARGOCD_APP_PARAMETERS='[{"name":"path","string":"./test"},{"array":["app=test"],"name":"extVars"},{"array":["ns=$ARGOCD_APP_REVISION_SHORT"],"name":"tlas"},{"array":["shared", "https://github.com/nr8-io/konn.git","https://github.com/nr8-io/k8s-libsonnet.git"],"name":"libs"}]'
```

## build image
docker buildx build -f Dockerfile -t eu.gcr.io/topvine-co/argocd-konn-jsonnet-plugin ./
docker push eu.gcr.io/topvine-co/argocd-konn-jsonnet-plugin