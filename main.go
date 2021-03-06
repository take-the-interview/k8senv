package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"

	//cnf "shep/config"

	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	cnf "github.com/take-the-interview/shep/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	appname     string
	namespace   string
	podname     = ""
	podnum      = ""
	secretspath = ""
	clientset   *kubernetes.Clientset
	data        = map[int]map[string]string{}
	secrets     = map[string]map[string]interface{}{}
	keys        = []int{}
	exitCode    = 0
)

func getClientSet(configFile string) {
	if configFile == "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
	} else {
		config, err := clientcmd.BuildConfigFromFlags("", configFile)
		if err != nil {
			panic(err.Error())
		}

		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
	}

	return
}

func calculateWeight(wStr string) (w int) {
	w = 1000
	if weight, err := strconv.Atoi(wStr); err == nil {
		w = weight
	}

	for _, k := range keys {
		if w == k {
			return
		}
	}
	if _, ok := data[w]; !ok {
		data[w] = make(map[string]string)
	}
	keys = append(keys, w)

	return
}

func getSecrets(secretPath string) (secretsMap map[string]interface{}) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(cnf.GetAWSRegion()),
	}))
	svc := secretsmanager.New(sess)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretPath),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "**** Problem Getting AWS Secrets %s, %v\n", secretPath, err)
		if !strings.HasPrefix(err.Error(), "ResourceNotFoundException") {
			exitCode = 1
		}
	} else {
		err = json.Unmarshal([]byte(*result.SecretString), &secretsMap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "**** Problem Parsing AWS Secrets %s, %v\n", secretPath, err)
			exitCode = 1
		}
	}
	return secretsMap
}

func injectSecrets() {
	for weight, node := range data {
		for envKey, envVal := range node {
			if strings.Contains(envVal, "{secret:") {
				re := regexp.MustCompile(`{secret:.*?:.*?}`)
				matches := re.FindAllString(envVal, -1)
				if matches != nil {
					for _, match := range matches {
						// fmt.Printf("-> %d - %s\n", idx, match[8:len(match)-1])
						chunks := strings.Split(match[8:len(match)-1], ":")
						secretPath := chunks[0]
						if secretPath == "" {
							secretPath = fmt.Sprintf("%s/env", secretspath)
						}
						secretKey := chunks[1]

						if _, ok := secrets[secretPath]; !ok {
							secrets[secretPath] = getSecrets(secretPath)
							fmt.Fprintf(os.Stderr, "**** Loading AWS Secrets %s [weight: 0] : %d items\n", secretPath, len(secrets[secretPath]))
						}

						if secretVal, ok := secrets[secretPath][secretKey]; ok {
							re := regexp.MustCompile(match)
							data[weight][envKey] = string(re.ReplaceAll([]byte(envVal), []byte(secretVal.(string))))
						} else {
							fmt.Fprintf(os.Stderr, "**** Uknown secret for %s=%s\n", envKey, envVal)
						}
					}
				}
			}
		}
	}
}

func getCM(cmName string) {
	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(cmName, metav1.GetOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "**** Problem Loading ConfigMap %s : %s\n", cmName, err.Error())
		exitCode = 1
		return
	}

	isShep, _ := cm.ObjectMeta.Annotations["conveyiq.com/shep"]
	if isShep == "" {
		return
	}

	appConfWeightStr, _ := cm.ObjectMeta.Labels["app-conf-weight"]
	appConfWeight := calculateWeight(appConfWeightStr)

	fmt.Fprintf(os.Stderr, "**** Loading ConfigMap %s [weight: %d] : %d items\n", cmName, appConfWeight, len(cm.Data))

	for cmKey, cmVal := range cm.Data {
		data[appConfWeight][cmKey] = cmVal
	}
}

func getPODnum() {
	if podname == "" {
		return
	}
	chunks := strings.Split(podname, "-")
	last := chunks[len(chunks)-1]
	if _, err := strconv.Atoi(last); err == nil {
		podnum = last
	}
}

func getPODInfo() {
	var ok bool
	namespace, ok = os.LookupEnv("K8S_POD_NAMESPACE")
	if !ok || namespace == "" {
		fmt.Fprintf(os.Stderr, "Unable to get namespace from env variable K8S_POD_NAMESPACE. Look for Downward API.\n")
		os.Exit(1)
	}
	appname, ok = os.LookupEnv("K8S_APP_NAME")
	if !ok || appname == "" {
		fmt.Fprintf(os.Stderr, "Unable to get namespace from env variable K8S_APP_NAME.\n")
		os.Exit(1)
	}

	podname, ok = os.LookupEnv("K8S_POD_NAME")
	secretspath, ok = os.LookupEnv("SECRETS_PATH")
	if !ok {
		fmt.Fprint(os.Stderr, "No SECRETS_PATH set, I won't try to get secrets from secretsmanager\n")
	}
	getPODnum()
}

func filterKeypair(k string, v string) (string, string, bool) {
	if strings.HasPrefix(k, "PERINSTANCE_") {
		if podnum == "" {
			return k, v, false
		}

		chunks := strings.SplitN(k, "_", 3)
		if len(chunks) != 3 {
			return k, v, false
		}

		envID := chunks[1]
		if podnum != envID {
			return k, v, false
		}

		keyName := chunks[2]
		return keyName, v, true
	} else {
		return k, v, true
	}
}

func main() {
	configFile := flag.String("c", "", "Config file and outside-the-cluster auth.")
	export := flag.Bool("e", false, "Export env vars")
	verbose := flag.Bool("verbose", false, "Verbose")

	flag.Parse()

	getPODInfo()

	getClientSet(*configFile)

	var err error
	var cmlist *v1.ConfigMapList
	var cms []v1.ConfigMap

	// app label configmaps
	labelSelector := fmt.Sprintf("app in (%s, all)", appname)
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
		Limit:         100,
	}

	cmlist, err = clientset.CoreV1().ConfigMaps(namespace).List(listOptions)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get ConfigMap with labelSelector %s. %v\n", labelSelector, err)
		os.Exit(1)
	}

	cms = append(cms, cmlist.Items...)

	// app-<app_name> label configmaps
	labelSelector = fmt.Sprintf("app-%s in (true, True, yes, Yes)", appname)
	listOptions = metav1.ListOptions{
		LabelSelector: labelSelector,
		Limit:         100,
	}

	cmlist, err = clientset.CoreV1().ConfigMaps(namespace).List(listOptions)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get ConfigMap with labelSelector %s. %v\n", labelSelector, err)
		os.Exit(1)
	}

	cms = append(cms, cmlist.Items...)

	if secretspath != "" {
		secretsEnvPath := fmt.Sprintf("%s/env", secretspath)
		secrets[secretsEnvPath] = getSecrets(secretsEnvPath)

		fmt.Fprintf(os.Stderr, "**** Loading AWS Secrets %s [weight: 0] : %d items\n", secretsEnvPath, len(secrets[secretsEnvPath]))

		calculateWeight("0")

		if _, ok := secrets[secretsEnvPath]; ok {
			// data[weight][envKey] = string(re.ReplaceAll([]byte(envVal), []byte(secretVal.(string))))
			for secretEnvKey, secretEnvVal := range secrets[secretsEnvPath] {
				data[0][secretEnvKey] = secretEnvVal.(string)
			}
		}

	}

	for _, item := range cms {
		isShep, _ := item.ObjectMeta.Annotations["conveyiq.com/shep"]
		if isShep == "" {
			continue
		}
		getCM(item.ObjectMeta.Name)
	}

	injectSecrets()

	fi, _ := os.Stdout.Stat()

	sort.Ints(keys)

	for _, k := range keys {
		for envKey, envVal := range data[k] {
			var ok bool
			if envKey, envVal, ok = filterKeypair(envKey, envVal); !ok {
				continue
			}
			var envLine string
			if *export {
				envLine = fmt.Sprintf("export %s=%s", envKey, envVal)
			} else {
				envLine = fmt.Sprintf("%s=%s", envKey, envVal)
			}
			fmt.Printf("%s\n", envLine)
			if *verbose && (fi.Mode()&os.ModeCharDevice) == 0 {
				fmt.Fprintf(os.Stderr, "%s\n", envLine)
			}
		}
	}
	if exitCode > 0 {
		os.Exit(exitCode)
	}
}
