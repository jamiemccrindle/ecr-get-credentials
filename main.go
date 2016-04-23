package main

import (
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"fmt"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"os"
)

type InstanceDocument struct {
	Region *string
}

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

	region := flag.String("region", "", "Region")
	replace := flag.Bool("replace", false, "Replace the docker config file")
	config := flag.String("config", "", "Docker Config File")
	metadata := flag.String("metadata", "http://169.254.169.254", "Meta data endpoint")
	flag.Parse()

	if len(*config) == 0 {
		flag.Usage()
		return
	}

	if len(*region) == 0 {
		instanceDocument, err := getInstanceDocument(*metadata);
		if err != nil {
			log.Fatalln(err)
		}
		region = instanceDocument.Region
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


// request the content of a http endpoint as a string
func getInstanceDocument(metadata string) (*InstanceDocument, error) {
	client := http.DefaultClient
	documentMetadataEndpoint := metadata + "/latest/dynamic/instance-identity/document"
	resp, err := client.Get(documentMetadataEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var instanceDocument InstanceDocument
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(bytes, &instanceDocument)
	return &instanceDocument, nil
}

