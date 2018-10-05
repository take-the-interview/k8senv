package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	appname   string
	namespace string
	clientset *kubernetes.Clientset
	data      = map[int]map[string]string{}
	keys      = []int{}
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

func getCM(cmName string) {
	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(cmName, metav1.GetOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "**** Problem Loading ConfigMap %s : %s\n", cmName, err.Error())
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
}

func main() {
	configFile := flag.String("c", "", "Config file and outside-the-cluster auth.")
	export := flag.Bool("e", false, "Export env vars")
	verbose := flag.Bool("verbose", false, "Verbose")

	flag.Parse()

	getPODInfo()

	getClientSet(*configFile)

	labelSelector := fmt.Sprintf("app in (%s, all)", appname)
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
		Limit:         100,
	}
	cms, err := clientset.CoreV1().ConfigMaps(namespace).List(listOptions)

	if err != nil {
		panic(err.Error())
	}

	for _, item := range cms.Items {
		getCM(item.ObjectMeta.Name)
	}

	fi, _ := os.Stdout.Stat()

	sort.Ints(keys)

	for _, k := range keys {
		for envKey, envVal := range data[k] {
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
}
