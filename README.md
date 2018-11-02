K8S env variables from configmaps
=================================

Runs inside container to prepopulate environment for application.

Looks for the following ENV variables when running inside of container.
K8S_APP_NAME
K8S_POD_NAMESPACE

Performs ConfigMaps search in K8S_POD_NAMESPACE with "app" label set to "K8S_APP_NAME" or "all".
It's expecting ConfigMap to have label "app-conf-weight" set (defaults to 1000 if the label is not present) to
specify the order of ENV variables loading (for interpolation purposes). Lower weight number get's loaded first.

Entrypoint.sh executes/evals k8senv on container startup setting ordered env variables in current process and than runs the application process.

# Per-Instance ENV support

For K8S statefulset it's possible to specify env variable to load only for
specified instance. This depends on downward-api and presence of
K8S_POD_NAME.

E.g.

instance: media-receiver-0
envVariable: PERINSTANCE_0_DATA=something
exposing: DATA=something
