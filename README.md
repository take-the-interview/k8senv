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

# Dockerfile example

For Rails app
```
FROM ruby:2.5.1-slim

ENV PATH="/usr/local/bundle/bin:$PATH" \
    GEM_HOME="/usr/local/bundle" \
    EPOINT_URL=https://github.com/take-the-interview/k8senv/releases/download/v0.4/entrypoint.sh \
    K8SENV_URL=https://github.com/take-the-interview/k8senv/releases/download/v0.4/k8senv.linux.amd64

ADD Gemfile Gemfile.lock /app/

WORKDIR /app

RUN apt-get update -qq && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y build-essential libpq-dev libxml2-dev nodejs git postgresql-client curl netcat && \
  curl -sL ${EPOINT_URL} > /ep && chmod 755 /ep && \
  curl -sL ${K8SENV_URL} > /usr/local/bin/k8senv && chmod 755 /usr/local/bin/k8senv && \
  useradd -u 1000 -U -d /app -s /bin/bash app && \
  gem install bundler && \
  cd /app && bundle install  && \
  chown -R app. /app /usr/local/bundle && \
  apt-get -y autoremove && \
  apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

USER app

ENTRYPOINT ["/ep"]
CMD [""]
```
