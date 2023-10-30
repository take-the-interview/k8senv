FROM ruby:2.7.8-slim

ENV PATH="/usr/local/bundle/bin:$PATH" \
    GEM_HOME="/usr/local/bundle" \
    EPOINT_URL=https://github.com/take-the-interview/k8senv/releases/download/v0.4/entrypoint.sh \
    K8SENV_URL=https://github.com/take-the-interview/k8senv/releases/download/v0.4/k8senv.linux.amd64

ADD . /app/

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
