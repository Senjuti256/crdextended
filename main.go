/*package main

import (
	//"context"
	"flag"
	"fmt"
	//"fmt"
	"path/filepath"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"

	klient "github.com/Senjuti256/crdextended/pkg/client/clientset/versioned"
	kInfFac "github.com/Senjuti256/crdextended/pkg/client/informers/externalversions"
	"github.com/Senjuti256/crdextended/pkg/controller"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func main() {
	// find the kubeconfig file
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String(
			"kubeconfig",
			filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// Building config from flags might fail inside the pod,
	// hence adding the code for usage of in-clusterconfig.
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Errorf("Building config from flags failed, %s, trying to build inclusterconfig", err.Error())
		// uses serviceAccount mounted inside the pod.
		config, err = rest.InClusterConfig()
		if err != nil {
			klog.Errorf("error %s building inclusterconfig", err.Error())
		}
	}

	// creating the clientset
	klientset, err := klient.NewForConfig(config)
	if err != nil {
		klog.Errorf("getting klient set %s\n", err.Error())
	}
	 fmt.Println(klientset)

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorf("getting std client %s\n", err.Error())
	}

	infoFact := kInfFac.NewSharedInformerFactory(klientset, 10*time.Minute)
	// infoFact :=
	ch := make(chan struct{})
	pc := controller.NewController(client, klientset, infoFact.Samplecontroller().V1alpha1().PipelineRuns(), infoFact.Samplecontroller().V1alpha1().TaskRuns())

	infoFact.Start(ch)
	if err := pc.Run(ch); err != nil {
		klog.Errorf("error running controller %s\n", err)
	}
}

*/

package main

import (
	"flag"
	"path/filepath"
	"time"

	clientset "github.com/Senjuti256/crdextended/pkg/client/clientset/versioned"
	informers "github.com/Senjuti256/crdextended/pkg/client/informers/externalversions"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"github.com/Senjuti256/crdextended/pkg/controller"

	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	// stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building example clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	exampleInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)

	trInformer := exampleInformerFactory.Samplecontroller().V1alpha1().TaskRuns()
	prInformer := exampleInformerFactory.Samplecontroller().V1alpha1().PipelineRuns()

	controller := controller.NewController(kubeClient, exampleClient,
		prInformer,
		trInformer) // Ak().V1alpha1().Klusters()

	stopCh := make(chan struct{})
	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	exampleInformerFactory.Start(stopCh)

	if err = controller.Run(1, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}














































/*
import (
	"flag"
	"log"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"

	klient "github.com/Senjuti256/crdextended/pkg/client/clientset/versioned"
	kInfFac "github.com/Senjuti256/crdextended/pkg/client/informers/externalversions"
	//"github.com/Senjuti256/crdextended/pkg/controller"
)

func main() {
	klog.InitFlags(nil)
	var kubeconfig *string

	klog.Info("Searching for kubeConfig")

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	klog.Info("Building config from the kubeConfig")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("Building config from flags failed, %s, trying to build inclusterconfig", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Printf("error %s building inclusterconfig", err.Error())
		}
	}
     
	klog.Info("getting the custom clientset")
	klientset, err := klient.NewForConfig(config)
	if err != nil {
		log.Printf("getting klient set %s\n", err.Error())
	}
    
	klog.Info("getting the k8s client")
	kubeclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("getting std client %s\n", err.Error())
	}
    
	infoFactory:= kInfFac.NewSharedInformerFactory(klientset, 20*time.Second)
	ch := make(chan struct{})
	c := controller.NewController(kubeclient, klientset)     
	//*controller := NewController(kubeClient, exampleClient,
		//kubeInformerFactory.Apps().V1().Deployments(),
		//exampleInformerFactory.Samplecontroller().V1alpha1().Foos())
	//c,err:= controller.NewController(klientset,10*time.Minute)
	
    klog.Info("Starting channel and Run mthod of controller")
	infoFactory.Start(ch)
	if err := c.Run(ch); err != nil {
		log.Printf("error running controller %s\n", err.Error())
	}
}
*/