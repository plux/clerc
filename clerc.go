package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/docopt/docopt-go"
	"io/ioutil"
	"net/http"
	"os/user"
)

type Buckets struct {
	Buckets []string `json:"buckets"`
}

type Keys struct {
	Keys []string `json:"keys"`
}

type Command int

const (
	buckets = iota
	keys
	obj
)

type Config struct {
	Url     string `json:"url"`
	Command Command
	Verbose bool `json:"verbose"`
	Bucket  string
	Key     string
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
	}
}

func parse_options() Args {
	usage := `clerc - Command LinE Riak Client

Usage:
  clerc BUCKET [KEY] [--url=URL] [--verbose] [--show]
  clerc -h | --help
  clerc --version

Options:
  --url=URL  Set the URL of the riak web API. [default: http://127.0.0.1:8098]
  --verbose  Show additional information, useful for debugging.
  --show     List objects instead of keys when listing a bucket.
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
		config.Url = args["--url"].(string)
	}
	if args["BUCKET"] == "/" {
		config.Command = buckets
	} else if args["BUCKET"] != nil {
		config.Command = keys
		config.Bucket = args["BUCKET"].(string)
	}
	if args["KEY"] != nil {
		config.Command = obj
		config.Key = args["KEY"].(string)
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
		"/buckets/"+config.Bucket+"/keys?keys=true")
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

func show_obj(config Config, bucket string, key string) {
	obj := get_obj(config, bucket, key)
	log(config, "Showing object: "+bucket+"/"+key)
	fmt.Println(obj)
}

func get_obj(config Config, bucket string, key string) string {
	resource := "/buckets/" + bucket + "/keys/" + key
	log(config, "Making request: "+config.Url+resource)
	resp, err := http.Get(config.Url + resource)
	perror(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	perror(err)
	log(config, "Got response: "+string(body))
	return prettify(body)
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
	log(config, "Making request: "+config.Url+resource)
	resp, err := http.Get(config.Url + resource)
	perror(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	perror(err)
	log(config, "Got response: "+prettify(body))
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