package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)


func httpRequestBody(cl *http.Client, url, method, token string) ([]byte, error) {
	switch method {
	case "GET","POST":
	default: return nil, errors.New("incorrect http method")
	}
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := cl.Do(req)
	if err != nil {
		fmt.Printf("Response error: %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Response status: %d: %s\n", resp.StatusCode, resp.Status)
		return nil, errors.New(resp.Status)
	}
	//outFile not set, returning []byte
	body, _ := io.ReadAll(resp.Body)
	return body, nil

}

type cluster struct {
	Name string
	Cluster struct {
		Server string
		Cert string `yaml:"certificate-authority-data"`
	}
}

type usertoken struct{
	Token string
}

type user struct {
	Name string
	User usertoken 
}

type clustercontext struct {
	User, Cluster string
	}

type context struct{	
	Name string
	Context clustercontext
}

type config struct {
		ApiVersion string `yaml:"apiVersion"`
		Kind string `yaml:"kind"`
		Clusters []cluster
		Users []user
		Contexts []context
		Currentcontext string `yaml:"current-context"`
}


func main(){
	cl:=&http.Client{}
	defaultusername:="myRancher2User"
	filename:="fullkubeconfig"

	token, ok:=os.LookupEnv("RANCHER2_API_TOKEN")
	if !ok {
		fmt.Printf("Error loading token, put it into RANCHER2_API_TOKEN envvar")
		return
	}

	apiurl, ok:=os.LookupEnv("RANCHER2_API_URL")
	if !ok {
		fmt.Printf("Error loading apiurl, put it into RANCHER2_API_URL envvar")
		return
	}

	body, err := httpRequestBody(cl, fmt.Sprintf("%s/v3/clusters",apiurl), "GET", token)
	if err!=nil {
		fmt.Printf("Error getting cluster list from API")
		return
	}
	
	var clusters struct {
		Data []struct{
			Name string
			Actions map[string]string
		}
	}

	if err := json.Unmarshal(body, &clusters); err != nil {
		fmt.Printf("JSON unmarshall error: %s\n", err)
		return
	}

	var configs []config

	for _, cluster := range (clusters.Data) {
		var c config
		body, err = httpRequestBody(cl, cluster.Actions["generateKubeconfig"], "POST", token)
		if err!=nil {
			fmt.Printf("Error getting cluster %s config from API",cluster.Name)
			return
		}
		
		var configjson struct {Config string}
		if err := json.Unmarshal(body, &configjson); err != nil {
			fmt.Printf("JSON unmarshall error: %s\n", err)
			return
		}

		err := yaml.Unmarshal([]byte(configjson.Config),&c)
		if err!=nil {
			fmt.Printf("YAML unmarshall error: %s\n", err)
			return	
		}
		configs=append(configs, c)
	}

	fullconfig := config {
		ApiVersion: "v1",
		Kind: "Config",
		Users:[]user{
			{
				Name: defaultusername,
				User: usertoken{
					Token: configs[0].Users[0].User.Token,
				},
			},
		},
	Currentcontext: configs[0].Clusters[0].Name,
	}

	for _, clconf := range(configs) {
		fullconfig.Contexts=append(fullconfig.Contexts,
				context{
					Name: clconf.Clusters[0].Name,
					Context: clustercontext{
						User: defaultusername,
						Cluster: clconf.Clusters[0].Name},
				})
		fullconfig.Clusters=append(fullconfig.Clusters, 
			clconf.Clusters...
			)
	}
	result,err := yaml.Marshal(fullconfig)
	if err!=nil {
		fmt.Printf("marshall to YAML error: %s\n", err)
		return	
	}
	err = os.WriteFile(filename,result,0644)
	if err!=nil {
		fmt.Printf("Error writing config to file: %s\n", err)
		return	
	}
}