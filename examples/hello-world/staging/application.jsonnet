local app = import '../application/main.libsonnet';
local k = import 'konn/main.libsonnet';

function(namespace='hello-world') (
  app.init(
    {
      namespace: namespace,
    },
    profile='staging'
  )
)
