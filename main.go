package main

import (
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"log"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"os"
)

type DockerConfig struct {
	Auths map[string]*DockerConfigAuth `json:"auths"`
}

type DockerConfigAuth struct {
	Auth  *string `json:"auth"`
	Email *string `json:"email"`
}

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: ecr-get-credentials -config DOCKER_CONFIG_LOCATION\n")
		flag.PrintDefaults()
	}

	region := flag.String("region", "", "Optional AWS region, otherwise read from instance metadata")
	replace := flag.Bool("replace", false, "Replace the docker config file")
	config := flag.String("config", "", "Docker Config File")
	flag.Parse()

	if len(*config) == 0 {
		flag.Usage()
		return
	}

	if len(*region) == 0 {
		metadata := ec2metadata.New(nil)
		metadataRegion, err := metadata.Region()
		if err !=nil {
			fmt.Println("Error: ", err)
		}
		region = &metadataRegion
	}

	awsConfig := aws.NewConfig().WithRegion(*region)
	svc := ecr.New(session.New(), awsConfig)
	authorizationTokenOutput, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		log.Fatalln(err)
	}
	dockerConfig, err := getDockerConfig(*config)
	if err != nil {
		log.Fatalln(err)
	}
	none := "none"
	for _, authorizationData := range authorizationTokenOutput.AuthorizationData {
		auth, found := dockerConfig.Auths[*authorizationData.ProxyEndpoint]
		if found {
      auth.Auth = authorizationData.AuthorizationToken
		} else {
			dockerConfig.Auths[*authorizationData.ProxyEndpoint] = &DockerConfigAuth{
				Auth: authorizationData.AuthorizationToken,
				Email: &none,
			}
		}
	}
	result, err := json.MarshalIndent(dockerConfig, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}
	if *replace {
		ioutil.WriteFile(*config, result, 0644)
	} else {
		println(string(result))
	}
}

func getDockerConfig(location string) (*DockerConfig, error) {
	if _, err := os.Stat("/path/to/whatever"); err == nil {
		file, err := ioutil.ReadFile(location)
		if err != nil {
			return nil, err
		}
		var dockerConfig DockerConfig
		json.Unmarshal(file, &dockerConfig)
		if dockerConfig.Auths == nil {
			dockerConfig.Auths = make(map[string]*DockerConfigAuth)
		}
		return &dockerConfig, nil
	} else if os.IsNotExist(err) {
		return &DockerConfig{
			Auths: make(map[string]*DockerConfigAuth),
		}, nil
	} else {
		return nil, err
	}
}
