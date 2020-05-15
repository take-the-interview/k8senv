package config

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	Clientset    *kubernetes.Clientset
	Namespace    string
	Params       map[string]string
	ConsulClient *consulapi.Client
}

func auth() (clientset *kubernetes.Clientset, namespace string) {
	namespace = viper.GetString("namespace")
	kubeconfig := viper.GetString("kubeconfig")

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	loadedConfig, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	currentNamespace := loadedConfig.Contexts[loadedConfig.CurrentContext].Namespace
	if namespace == "" {
		if currentNamespace == "" {
			fmt.Fprintln(os.Stderr, "You need to specify namespace.")
			os.Exit(1)
		}
		namespace = currentNamespace
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return
}

func (c *Config) loadShepConfig() {
	verbose := viper.GetBool("verbose")

	var err error

	cm, err := c.Clientset.CoreV1().ConfigMaps("default").Get("shep-config", metav1.GetOptions{})

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			if verbose {
				fmt.Println("CM shep-config not found")
			}
		} else {
			panic(err.Error())
		}
	}

	c.Params = cm.Data

}

func (c *Config) initConsul() {
	config := consulapi.DefaultConfig()

	config.Address = c.Params["consul_http_addr"]
	if ssl, ok := c.Params["consul_http_ssl"]; ok && ssl == "true" {
		config.Scheme = "https"
	} else {
		config.Scheme = "http"
	}

	if _, ok := c.Params["consul_http_user"]; ok {
		if _, ok := c.Params["consul_http_pass"]; ok {
			config.HttpAuth = &consulapi.HttpBasicAuth{
				Username: c.Params["consul_http_user"],
				Password: c.Params["consul_http_pass"],
			}
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	config.HttpClient = &http.Client{Transport: tr}

	if token, ok := c.Params["consul_http_token"]; ok {
		config.Token = token
	}

	c.ConsulClient, _ = consulapi.NewClient(config)
}

func NewConfig() (c *Config) {
	clientset, namespace := auth()

	c = &Config{
		Clientset: clientset,
		Namespace: namespace,
	}
	c.loadShepConfig()
	return
}

func NewConsulConfig() (c *Config) {
	clientset, namespace := auth()

	c = &Config{
		Clientset: clientset,
		Namespace: namespace,
	}
	c.loadShepConfig()
	c.initConsul()
	return
}
