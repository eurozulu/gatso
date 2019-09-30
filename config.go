package main

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path"
)

const configEnvKey = "TODOHOME"
const configName = "todo-properties.json"

// Config is a container for configuration properties, read from a json file located
// in a directory specified by the TODOHOME environment variable
type Config map[string]interface{}

func Newconfig() (Config, error) {
	// First check if ENV TODOHOME is set
	home, ok := os.LookupEnv(configEnvKey)
	if !ok || len(home) == 0 {
		var err error
		home, err = os.Getwd()
		if nil != err {
			return nil, err
		}
	}
	f, err := os.Open(path.Join(home, configName))
	if nil != err {
		return nil, err
	}
	by, err := ioutil.ReadAll(f)
	if nil != err {
		return nil, err
	}

	var cf Config
	if err = json.Unmarshal(by, &cf); nil != err {
		return nil, err
	}
	return cf, nil
}

func (cf Config) ReadString(key string, value string) string {
	v, ok := cf[key]
	if !ok {
		return value
	}
	s, ok := v.(string)
	if !ok {
		return value
	}
	return s
}

func (cf Config) ReadUrl(key string) *url.URL {
	s := cf.ReadString(key, "")
	if s == "" {
		return nil
	}
	u, err := url.Parse(s)
	if nil != err {
		return nil
	}
	return u
}

func (cf Config) ReadInt(key string, value int) int {
	v, ok := cf[key]
	if !ok {
		return value
	}

	i, ok := v.(int)
	if !ok {
		return value
	}

	return i
}

func (cf Config) ReadBool(key string, value bool) bool {
	v, ok := cf[key]
	if !ok {
		return value
	}
	b, ok := v.(bool)
	if !ok {
		return value
	}
	return b
}
