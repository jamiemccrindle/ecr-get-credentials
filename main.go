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
	"strings"
)

type DockerConfigAuth struct {
	Auth  *string `json:"auth"`
	Email *string `json:"email"`
}

type DockerConfig struct {
	Auths map[string]*DockerConfigAuth `json:"auths"`
}

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: ecr-get-credentials -config DOCKER_CONFIG_LOCATION\n")
		flag.PrintDefaults()
	}

	region := flag.String("region", "", "Optional AWS region, otherwise read from instance metadata")
	replace := flag.Bool("replace", false, "Replace the docker config file")
	config := flag.String("config", "", "Docker Config File")
	configType := flag.String("type", "", "If empty, it will autodetect, otherwise the options are .dockercfg or config.json")

	flag.Parse()

	if len(*config) == 0 {
		flag.Usage()
		return
	}

	if len(*configType) == 0 {
		if strings.HasSuffix(*config, "dockercfg") {
			*configType = ".dockercfg"
		} else if strings.HasSuffix(*config, "config.json") {
			*configType = "config.json"
		} else {
			log.Fatal("Could not determine config type automatically, specify it using the -type argument")
		}
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

	var result []byte

	if *configType == ".dockercfg" {
		result, err = updateDockerConfigVersion1(config, authorizationTokenOutput)
	} else {
		result, err = updateDockerConfigVersion2(config, authorizationTokenOutput)
	}

	if err != nil {
		log.Fatalln(err)
	}
	if *replace {
		ioutil.WriteFile(*config, result, 0644)
	} else {
		println(string(result))
	}
}

func updateDockerConfigVersion1(config *string, authorizationTokenOutput *ecr.GetAuthorizationTokenOutput) ([]byte, error) {
	dockerConfig, err := getDockerConfigVersion1(*config)
	if err != nil {
		return nil, err
	}
	none := "none"
	for _, authorizationData := range authorizationTokenOutput.AuthorizationData {
		_, found := dockerConfig[*authorizationData.ProxyEndpoint]
		if found {
			dockerConfig[*authorizationData.ProxyEndpoint].Auth = authorizationData.AuthorizationToken
		} else {
			dockerConfig[*authorizationData.ProxyEndpoint] = &DockerConfigAuth{
				Auth: authorizationData.AuthorizationToken,
				Email: &none,
			}
		}
	}
	result, err := json.MarshalIndent(dockerConfig, "", "  ")
	return result, err
}

func updateDockerConfigVersion2(config *string, authorizationTokenOutput *ecr.GetAuthorizationTokenOutput) ([]byte, error) {
	dockerConfig, err := getDockerConfigVersion2(*config)
	if err != nil {
		return nil, err
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
	return result, err
}

func getDockerConfigVersion2(location string) (*DockerConfig, error) {
	if _, err := os.Stat(location); err == nil {
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

func getDockerConfigVersion1(location string) (map[string]*DockerConfigAuth, error) {
	if _, err := os.Stat(location); err == nil {
		file, err := ioutil.ReadFile(location)
		if err != nil {
			return nil, err
		}
		var dockerConfig map[string]*DockerConfigAuth
		json.Unmarshal(file, &dockerConfig)
		return dockerConfig, nil
	} else if os.IsNotExist(err) {
		dockerConfig := make(map[string]*DockerConfigAuth)
		return dockerConfig, nil
	} else {
		return nil, err
	}
}