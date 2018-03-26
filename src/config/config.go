package config

import (
	"flag"

	"github.com/spf13/pflag"
)

var (
	address            = pflag.String("address", "0.0.0.0", "bind http address")
	port               = pflag.String("port", "8181", "http listen port")
	namespace          = pflag.String("namespace", "kube-system", "tiller namespace")
	tillerHost         = pflag.String("TillerHost", "", "tiller host")
	tillerPortForward  = pflag.Bool("TillerPortForward", false, "TillerPortForward")
)

var conf *Config

type Config struct {
	Address              string `json:"address"`
	Port                 string `json:"port"`
	Namespace            string `json:"namespace"`
	TillerHost           string `json:"tillerHost"`
	TillerPortForward    bool   `json:"tillerPortForward"`
}

func init() {
	conf = loadConfig()
}

func loadConfig() *Config {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	return &Config{
		Address:              *address,
		Port:                 *port,
		Namespace:            *namespace,
		TillerHost:           *tillerHost,
		TillerPortForward:    *tillerPortForward,
	}
}

func GetConfig() *Config {
	return conf
}
