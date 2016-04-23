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
      -metadata string
            Meta data endpoint (default "http://169.254.169.254")
      -region string
            Region
      -replace
            Replace the docker config file

## Building

    ./build.bash