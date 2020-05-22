K8S env variables from configmaps
=================================

Runs inside container to prepopulate environment for application.

Looks for the following ENV variables when running inside of container.
K8S_APP_NAME
K8S_POD_NAMESPACE

Performs ConfigMaps search in K8S_POD_NAMESPACE with "app" label set to "K8S_APP_NAME" or "all" and/or with "app-<app_name>" label set to "true".
With "app-<app_name>" label we can have one ConfigMap loaded by several different applications.
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

# AWS Secrets Manager support

By default k8senv is looking for env variable SECRETS_PATH
which indicates the path in AWS Secrets Manager from which to pull Secret Env Variables.
Example: runtime/e1/dora

If SECRETS_PATH is not set it won't try to read secret from secretsmanager

It will pull with weight:0 secret from AWS Secrets Manager from the
${SECRETS_PATH}/env which containts json with key/value pairs of each secret.
These key/value pairs are then exported as ENV variables.

To reference secret stored in AWS Secrets Manager in env variable use the
following format (interpolation works):

{secret:secret_path:secret_key}

Where:
- secret_path - is name of the secret in Secrets Manager. If it is omitted (e.g. {secret::secret_key}), it will default to runtime/env.K8S_POD_NAMESPACE/env.K8S_APP_NAME/env.
- secret_key - is name of the Key in JSON data stored in Secrets Manager secret

k8senv will automatically figure out which secret from AWS Secrets Manager
to fetch and interpolate env variables.
