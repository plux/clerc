package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docopt/docopt-go"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strings"
)

type Buckets struct {
	Buckets []Bucket `json:"buckets"`
}

type Keys struct {
	Keys []Key `json:"keys"`
}

type Command int
type Url string
type Bucket string
type Key string

const (
	buckets = iota
	keys
	obj
	put
	delete
)

type Config struct {
	Url     Url `json:"url"`
	Command Command
	Verbose bool `json:"verbose"`
	Bucket  Bucket
	Key     Key
	Show    bool `json:"show"`
}

type Args map[string]interface{}

func main() {
	args := parse_options()
	config := init_config(args)
	switch config.Command {
	case buckets:
		show_buckets(config)
	case obj:
		show_obj(config, config.Bucket, config.Key)
	case keys:
		if config.Show {
			show_objs(config)
		} else {
			show_keys(config)
		}
	case put:
		obj := read_stdin()
		put_obj(config, config.Bucket, config.Key, obj)
	case delete:
		delete_obj(config, config.Bucket, config.Key)
	}
}

func parse_options() Args {
	usage := `clerc - Command LinE Riak Client

Usage:
  clerc BUCKET KEY [--url=URL] [--put | --delete] [--verbose]
  clerc BUCKET [--url=URL] [--verbose] [--show]
  clerc -h | --help
  clerc --version

Options:
  --url=URL  Set the URL of the riak web API.
  --verbose  Show additional information, useful for debugging.
  --show     List objects instead of keys when listing a bucket.
  --put      Put object which is read from stdin.
  -h --help  Show this screen.
  --version  Show version.
`
	args, err := docopt.Parse(usage, nil, true, "clerc (Command LinE Riak Client) 0.1", false)
	perror(err)
	return args
}

func init_config(args Args) Config {
	config := read_config_file()
	if args["--verbose"] == true {
		config.Verbose = true
	}
	if args["--show"] == true {
		config.Show = true
	}
	if args["--url"] != nil {
		config.Url = Url(args["--url"].(string))
	}
	if args["BUCKET"] == "/" {
		config.Command = buckets
	} else if args["BUCKET"] != nil {
		config.Command = keys
		config.Bucket = Bucket(args["BUCKET"].(string))
	}
	if args["KEY"] != nil {
		config.Command = obj
		config.Key = Key(args["KEY"].(string))
	}
	if args["--put"] == true {
		config.Command = put
	}
	if args["--delete"] == true {
		config.Command = delete
	}
	return config
}

func read_config_file() Config {
	config := new_config()
	usr, err := user.Current()
	perror(err)
	bytes, err := ioutil.ReadFile(usr.HomeDir + "/.clerc")
	if err != nil {
		log(config, "unable to read file :(")
	} else {
		err = json.Unmarshal(bytes, &config)
		log(config, "config: "+string(bytes))
		perror(err)
	}
	return config
}

func new_config() Config {
	return Config{
		Url:     "http://127.0.0.1:8098",
		Verbose: false,
		Show:    false,
	}
}

func read_stdin() []byte {
	bytes, err := ioutil.ReadAll(os.Stdin)
	perror(err)
	return bytes
}

func show_keys(config Config) {
	keys := get_keys(config)
	log(config, "Listing keys:")
	for _, key := range keys.Keys {
		fmt.Println(key)
	}
}

func get_keys(config Config) Keys {
	var keys Keys
	make_request(&keys, config,
		"/buckets/"+string(config.Bucket)+"/keys?keys=true")
	return keys
}

func show_objs(config Config) {
	keys := get_keys(config)
	for _, key := range keys.Keys {
		fmt.Println("Key: " + key)
		show_obj(config, config.Bucket, key)
		fmt.Println("")
	}
	return
}

func show_obj(config Config, bucket Bucket, key Key) {
	obj := get_obj(config, bucket, key)
	log(config, "Showing object: "+string(bucket)+"/"+string(key))
	fmt.Println(obj)
}

func delete_obj(config Config, bucket Bucket, key Key) {
	resource := "/riak/" + string(bucket) + "/" + string(key)
	log(config, "Deleting: "+string(config.Url)+resource)
	req, err := http.NewRequest("DELETE", string(config.Url)+resource, nil)
	perror(err)
	response, err := http.DefaultClient.Do(req)
	perror(err)
	log(config, "Deleted object: " + string(key))
	read_body(config, response)
}

func put_obj(config Config, bucket Bucket, key Key, obj []byte) {
	resource := "/riak/" + string(bucket) + "/" + string(key)
	log(config, "Making request: "+string(config.Url)+resource)
	reader := strings.NewReader(string(obj))
	response, err := http.Post(string(config.Url)+resource, "application/json", reader)
	perror(err)
	assert_status(response, 204)
	read_body(config, response)
}

func read_body(config Config, response *http.Response) []byte {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	perror(err)
	log(config, "Got response: "+string(body))
	return body
}

func get_obj(config Config, bucket Bucket, key Key) string {
	resource := "/buckets/" + string(bucket) + "/keys/" + string(key)
	log(config, "Making request: "+string(config.Url)+resource)
	response, err := http.Get(string(config.Url) + resource)
	perror(err)
	body := read_body(config, response)
	return prettify(body)
}

func assert_status(response *http.Response, status int) {
	if response.StatusCode != status {
		perror(errors.New("Unexpected status: " + response.Status))
	}
}

func prettify(data []byte) string {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, data, "", "    ")
	if err != nil {
		return string(data)
	} else {
		return string(prettyJson.Bytes())
	}
}

func show_buckets(config Config) {
	buckets := get_buckets(config)
	log(config, "Listing buckets:")
	for _, bucket := range buckets.Buckets {
		fmt.Println(bucket)
	}
}

func get_buckets(config Config) Buckets {
	var buckets Buckets
	make_request(&buckets, config, "/buckets?buckets=true")
	return buckets
}

func make_request(data interface{}, config Config, resource string) {
	log(config, "Making request: "+string(config.Url)+resource)
	response, err := http.Get(string(config.Url) + resource)
	perror(err)
	assert_status(response, 200)
	body := read_body(config, response)
	json.Unmarshal(body, &data)
}

func log(config Config, str string) {
	if config.Verbose {
		fmt.Println("*** " + str)
	}
}

func perror(err error) {
	if err != nil {
		panic(err)
	}
}
