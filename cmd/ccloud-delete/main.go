package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func init() {
	log.SetOutput(os.Stdout)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// if the force flag is true, log the action
// if not append the 'skipping' prefix
func forcedLog(msg string, dti DeleteTopicsInput) string {
	if !dti.Force {
		log.Printf("DRY RUN (SKIPPING): %s", msg)
	}
	return msg

}
func supportedEnvironments() []string {
	return []string{
		"dev",
		"sandbox",
		"integration",
		"staging",
		"qeint",
		"production",
	}
}

// DeleteTopicsInput input required to run
type DeleteTopicsInput struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Environment  string `json:"environment"`
	RESTEndpoint string `json:"RESTEndpoint"`
	ClusterID    string `json:"clusterID"`
	Force        bool   `json:"force"`
}

// Validate the DeleteTopicsInput data
func (i *DeleteTopicsInput) Validate() (err error) {
	if !contains(supportedEnvironments(), i.Environment) {
		return fmt.Errorf("invalid environment: %s", i.Environment)
	}
	return err
}

// InputFromFile get input data from a json file
func InputFromFile(filepath string) DeleteTopicsInput {
	jsonFile, _ := os.ReadFile(filepath)
	var dti DeleteTopicsInput
	err := json.Unmarshal(jsonFile, &dti)
	if err != nil {
		panic(err)
	}
	return dti
}

// InputFromArgs get input from flags
func InputFromArgs() DeleteTopicsInput {
	force := flag.Bool("force", false, "default: dry-run. With force, execute the deletions")
	username := flag.String("username", "", "API Token Username")
	password := flag.String("password", "", "API Token Password")
	environment := flag.String(
		"environment", "", "sandbox|dev|integration|qeint|staging|production")
	RESTEndpoint := flag.String("RESTEndpoint", "", "confluent cloud REST endpoint")
	clusterID := flag.String("clusterID", "", "confluent cloud REST endpoint")
	flag.Parse()
	return DeleteTopicsInput{
		Username:     *username,
		Password:     *password,
		Environment:  *environment,
		RESTEndpoint: *RESTEndpoint,
		ClusterID:    *clusterID,
		Force:        *force,
	}
}

// Credentials combine and encode confluent cloud credentials for basic authentication
// https://docs.confluent.io/cloud/current/api.html#section/Authentication
func Credentials(dti DeleteTopicsInput) string {
	result := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", dti.Username, dti.Password)))
	return string(result)
}

// ParseTopicList get a list of topic name strings from the REST response
func ParseTopicList(response []byte) []string {
	result := []string{}
	var topicData map[string]interface{}

	_ = json.Unmarshal(response, &topicData)
	for _, topic := range topicData["data"].([]interface{}) {
		m := topic.(map[string]interface{})
		result = append(result, fmt.Sprint(m["topic_name"]))
	}
	return result
}

// QueryListTopics query confluent clouds for teh list of topics on a cluster
func QueryListTopics(dti DeleteTopicsInput) []byte {
	url := dti.RESTEndpoint + "/kafka/v3/clusters/" + dti.ClusterID + "/topics"
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", Credentials(dti)))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		panic(fmt.Errorf("failed to get topic list: %v", res.Status))
	}
	return body
}

// DeleteTopic delete a single topic by name
func DeleteTopic(dti DeleteTopicsInput, topicName string) (err error) {
	logMsg := fmt.Sprintf("deleting topic: %s", topicName)
	forcedLog(logMsg, dti)
	if !dti.Force {
		return err
	}

	url := dti.RESTEndpoint + "/kafka/v3/clusters/" + dti.ClusterID + "/topics/" + topicName

	req, _ := http.NewRequest("DELETE", url, nil)

	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", Credentials(dti)))

	res, _ := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	if res.StatusCode != 204 {
		err = errors.New(res.Status)
	}
	return err
}

// FilterTopicsByEnvironment Use the environment to filter a list of topic names
func FilterTopicsByEnvironment(environment string, topics []string) (result []string) {
	for _, v := range topics {
		if strings.HasPrefix(v, environment+".") {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		panic(errors.New("no topics to delete"))
	}
	return result
}

// DeleteTopics loop through topics and delete
// if the first deletion fails, wait a second and try once more
// confluent  does some rate limiting and this is a cheap hack
func DeleteTopics(dti DeleteTopicsInput, topics []string) {
	for _, v := range topics {
		err := DeleteTopic(dti, v)
		if err != nil {
			time.Sleep(1 * time.Second)
			DeleteTopic(dti, v)
		}
	}
}

func main() {
	dti := InputFromArgs()
	err := dti.Validate()
	if err != nil {
		panic(err)
	}
	response := QueryListTopics(dti)
	allTopics := ParseTopicList(response)
	topics := FilterTopicsByEnvironment(dti.Environment, allTopics)
	DeleteTopics(dti, topics)
}
