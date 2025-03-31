local k = import 'konn/main.libsonnet';

//
k.app(
  [
    k.fromYaml(importstr './templates/hello-world-deploy.yaml'),
    k.fromYaml(importstr './templates/hello-world-svc.yaml'),
  ],
  defaults={
    name: 'hello-world',
    namespace: 'hello-world',
  },
  profiles={
    staging: {
      name: 'hello-world-stg',
    },
  }
)
