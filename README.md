# ECR Credentials

The docker credentials that `aws ecr get-login` provides only last 12 hours.
 This docker provides a way to keep docker auth config up to date. It calls
 the AWS ECR get-credentials endpoint and replaces a configured docker config
 file with an updated one.

## Usage

### Usage in a docker in EC2 coreos instance

    docker run -v /home/core/:/home/core/ \
        jamiemccrindle/ecr-get-credentials \
        -config /home/core/.dockercfg \
        -replace

### Usage outside of a docker

    Usage: ecr-get-credentials -config DOCKER_CONFIG_LOCATION
      -config string
            Docker Config File
      -region string
            Optional AWS region, otherwise read from instance metadata
      -replace
            Replace the docker config file

### Usage in cloud-init

    - name: ecr-get-credentials.service
      enable: true
      command: start
      content: |
        [Unit]
        Description=ECR login
        Requires=docker.service
        After=docker.service

        [Service]
        Type=oneshot
        ExecStartPre=-/usr/bin/docker kill ecr-get-credentials
        ExecStartPre=-/usr/bin/docker rm -f ecr-get-credentials
        ExecStartPre=/usr/bin/docker pull jamiemccrindle/ecr-get-credentials
        ExecStart=/usr/bin/docker run -v /home/core/:/home/core/ jamiemccrindle/ecr-get-credentials -config /home/core/.dockercfg -replace

        [Install]
        WantedBy=multi-user.target
    - name: ecr-get-credentials.timer
      enable: true
      command: start
      content: |
        [Unit]
        Description=Runs ECR login every hour
        Requires=docker.service
        After=docker.service

        [Timer]
        OnUnitActiveSec=10h
        Unit=ecr-get-credentials.service

        [Install]
        WantedBy=multi-user.target

## Building

    ./build.bash