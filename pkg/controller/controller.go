package controller

import (
	"context"
	"fmt"
	"log"
	//"math/rand"
	//"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/Senjuti256/crdextended/pkg/apis/sde.dev/v1alpha1"
	pclientset "github.com/Senjuti256/crdextended/pkg/client/clientset/versioned"
	samplescheme "github.com/Senjuti256/crdextended/pkg/client/clientset/versioned/scheme"
	pinformers "github.com/Senjuti256/crdextended/pkg/client/informers/externalversions/sde.dev/v1alpha1"
	plisters "github.com/Senjuti256/crdextended/pkg/client/listers/sde.dev/v1alpha1"
	//"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const controllerAgentName = "pipelinerun"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by PR"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "PR synced successfully"
)

// Controller is the controller implementation for Foo resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset pclientset.Interface

	prLister plisters.PipelineRunLister
	prSynced cache.InformerSynced

	trLister plisters.TaskRunLister
	trSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	wq workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset pclientset.Interface,
	prInformer pinformers.PipelineRunInformer,
	trInfomer pinformers.TaskRunInformer) *Controller {

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:   kubeclientset,
		sampleclientset: sampleclientset,
		prLister:        prInformer.Lister(),
		prSynced:        prInformer.Informer().HasSynced,
		trLister:        trInfomer.Lister(),
		trSynced:        trInfomer.Informer().HasSynced,
		wq:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "PRs"),
		recorder:        recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change
	prInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueuePR,
		UpdateFunc: func(old, new interface{}) {
			oldpod := old.(*v1alpha1.PipelineRun)
				newpod := new.(*v1alpha1.PipelineRun)
				if newpod == oldpod {
					return
				}
			controller.enqueuePR(new)
		},
		DeleteFunc: func(obj interface{}) {
			controller.deletePR(obj)
		},
	})
	// Set up an event handler for when TaskRun resources change
	// trInfomer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
	// 	FilterFunc: contra.FilterController(&v1alpha1.PipelineRun{}),
	// 	Handler:    contra.HandleAll(controller.enqueueFoo),
	// })
	// trInfomer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	// 	DeleteFunc: func(obj interface{}) {
	// 		controller.deleteTR(obj)
	// 	},
	// })
	return controller
}

// func FilterControllerGK(gk schema.GroupKind) func(obj interface{}) bool {
// 	return func(obj interface{}) bool {
// 		object, ok := obj.(metav1.Object)
// 		if !ok {
// 			return false
// 		}

// 		owner := metav1.GetControllerOf(object)
// 		if owner == nil {
// 			return false
// 		}

// 		ownerGV, err := schema.ParseGroupVersion(owner.APIVersion)
// 		return err == nil &&
// 			ownerGV.Group == gk.Group &&
// 			owner.Kind == gk.Kind
// 	}
// }

// // FilterController makes it simple to create FilterFunc's for use with
// // cache.FilteringResourceEventHandler that filter based on the
// // controlling resource.
// func FilterController(r kmeta.OwnerRefable) func(obj interface{}) bool {
// 	return FilterControllerGK(r.GetGroupVersionKind().GroupKind())
// }

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash() // Error handling
	defer c.wq.ShutDown()    // Till work queue get empty

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting PR controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.prSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch k workers to process Foo resources
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh // waiting for channel to return all queries
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.wq.Get() // Get the item from workqueue
	if shutdown {
		klog.Info("Shutting down")
		return false
	}

	defer c.wq.Forget(obj) // prevent item to re-enter queue at the end of the function

	// Method 1 : Getting PipelineRun Resource
	// item := obj.(*v1alpha1.PipelineRun)
	// name := item.GetName()
	// namespace := item.GetNamespace()
	var key string
	var ok bool

	if key, ok = obj.(string); !ok {
		// As the item in the workqueue is actually invalid, we call
		// Forget here else we'd go into a loop of attempting to
		// process a work item that is invalid.
		// c.workqueue.Forget(obj)
		fmt.Println("-------------------------------ERROR GETTING KEY --------------------------------------")
		utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
	}
	fmt.Println(" -------------------------------Trying to get Namespace and name -----------------------------")
	fmt.Printf("\nKey: %v \n", key)
	namespace, name, _ := cache.SplitMetaNamespaceKey(key)
	fmt.Printf("\nName: %v \nNamespace: %v", name, namespace)
	fmt.Println(" -------------------------------GETTING PR CR-----------------------")
	// Getting pr resource of PipelineRun Type
	pr, err := c.prLister.PipelineRuns(namespace).Get(name) // Get PR Resource
	
	if errors.IsNotFound(err) {
		fmt.Println("PipelineRun Resouce Not Found")
	}
	if err != nil {
		klog.Errorf("error %s, Getting the PR resource from lister.", err.Error())
		return false
	}
	fmt.Printf("\nPR Resource Spec: %+v \n", pr.Spec)
	fmt.Printf("\nPR Resource Status: %+v \n", pr.Status)

	if pr.Spec.Message != pr.Status.Message || pr.Spec.Count != pr.Status.Count {
		fmt.Printf("\n-------------------------Checked conditions ------------------------\n")
		// Call SyncHandler
		err := c.syncHandlerTR(pr)
		if err != nil {
			klog.Errorf("Error while syncing the current vs desired state for PipelineRun %v: %v\n", pr.Name, err.Error())
			return false
		}

		fmt.Printf("\nPR Resource Status: %+v \n", pr.Status)
		c.updateFooStatus(pr)
		// fmt.Printf("\nFoo Resource Status Error: %+v \n", errS)
		fmt.Printf("\nPR Resource Spec: %+v \n", pr.Spec)
		fmt.Printf("\nPR Resource Status: %+v \n", pr.Status)

	}

	return true
}

// Set Current State =  Desired State
func (c *Controller) syncHandlerTR(foo *v1alpha1.PipelineRun) error {
	// Creating Task Run Object from foo resource
	//tr := createTaskRun(foo)
	// fmt.Printf("\nTask Run Resource: %+v \n", tr)
/*	count := 0
	// Creating TaskRun Instance
	for count < 2 {
		taskRun, tErr := c.sampleclientset.SamplecontrollerV1alpha1().TaskRuns(tr.Namespace).Create(context.TODO(), tr, metav1.CreateOptions{})
		if tErr != nil {
			if errors.IsAlreadyExists(tErr) {
				fmt.Printf("\nCreating Another Task Run: %v", tErr)
				count += 1
			} else if tErr != nil {
				fmt.Printf("\nError Creating taskrun: %v\n", tErr)
			}
		} else {
			// Creating Pods for the TaskRun CR
			c.createTaskRunPods(taskRun)
			break
		}
	}
*/




// Creating Task Run Object from foo resource
		tr, err := c.sampleclientset.SamplecontrollerV1alpha1().TaskRuns(foo.Namespace).Create(context.TODO(), newTaskRun(foo), metav1.CreateOptions{})
		if err != nil {
			klog.Errorf("TaskRun creation failed for Pipeline %s", foo.Name)
			return err
		}

		if tr != nil {
			klog.Infof("Taskrun %s has been created for PipelineRun %s", tr.Name, foo.Name)
			if err := c.createPods(foo, tr); err != nil {
				klog.Errorf("Taskrun %s failed to create pods", tr.Name)
			}
		}

	return nil
}

func newTaskRun(prun *v1alpha1.PipelineRun) *v1alpha1.TaskRun {
	return &v1alpha1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v-tr-%v", prun.Name, prun.ObjectMeta.Generation),
			Namespace: prun.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(prun, v1alpha1.SchemeGroupVersion.WithKind("PipelineRun")),
			},
		},
		Spec: v1alpha1.TaskRunSpec{
			Message: prun.Spec.Message,
			Count:   prun.Spec.Count,
		},
	}
}

func (c *Controller) createPods(prun *v1alpha1.PipelineRun, trun *v1alpha1.TaskRun) error {
	// var podCreate, podDelete bool
	itr := trun.Spec.Count
	// Creates pod
	for i := 0; i < itr; i++ {
		nPod, err := c.kubeclientset.CoreV1().Pods(trun.Namespace).Create(context.TODO(), newPod(trun), metav1.CreateOptions{})
		if err != nil {
			klog.Errorf("Pod creation failed for  TR CR %v\n", trun.Name)
			return err
		}
		if nPod.Name != "" {
			klog.Infof("Pod %v created successfully!\n", nPod.Name)
		}
	}

	return nil
}

// Creates the new pod with the specified template
func newPod(trun *v1alpha1.TaskRun) *corev1.Pod {
	labels := map[string]string{
		"controller": trun.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       labels,
			GenerateName: fmt.Sprintf("%s-", trun.Name),
			Namespace:    trun.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(trun, v1alpha1.SchemeGroupVersion.WithKind("TaskRun")),
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: "Never",
			Containers: []corev1.Container{
				{
					Name:  "static-nginx",
					Image: "nginx:latest",
					Env: []corev1.EnvVar{
						{
							Name:  "MESSAGE",
							Value: trun.Spec.Message,
						},
					},
					Command: []string{
						"/bin/sh",
					},
					Args: []string{
						"-c",
						"echo '$(MESSAGE)'",
					},
				},
			},
		},
	}
}
















// Function for Creating Pods
/*func (c *Controller) createTaskRunPods(tr *v1alpha1.TaskRun) error {

	// desired number of pods for tr
	desiredPods := tr.Spec.Count

	// If number of upPods lower than desired Pods

	noPodsCreate := desiredPods
	log.Printf("\nCreating desired Pods for CR %v; Expected: %v\n\n", tr.Name, desiredPods)

	// Creating desired number of pods
	for i := 0; i < noPodsCreate; i++ {
		log.Println("---------Started Creating Pod --------------")
		podNew, err := c.kubeclientset.CoreV1().Pods(tr.Namespace).Create(context.TODO(), createPod(tr), metav1.CreateOptions{})
		log.Println("---------Successfully Created Pods --------------")
		if err != nil {
			if errors.IsAlreadyExists(err) {
				// So we try to create another pod with different name
				noPodsCreate++
			} else {
				log.Println("-------------------------Returning ERROR------------------")
				return err
			}
		} else {
			log.Printf("Successfully created %v Pod for CR %v \n", podNew.Name, tr.Name)
		}

	}

	log.Printf("\nSuccessfully created %v Pods for CR %v \n", desiredPods, tr.Name)

	return nil
}
*/
// When Task Run is Deleted
func (c *Controller) deleteTR(obj interface{}) error {
	tr := obj.(*v1alpha1.TaskRun)

	prName := tr.GetName()
	prName = strings.Split(prName, "-")[0]
	prNamespace := tr.GetNamespace()

	fmt.Printf("\n\nDeleting TR - %v\n\n", tr.GetName())

	foo, err := c.prLister.PipelineRuns(prNamespace).Get(prName) // Get Foo Resource
	if err != nil {
		fmt.Printf("\nError getting foo resource from TaskRun Instance: %v", foo)
		return nil
	}
	c.syncHandlerTR(foo)

	// // Creating TaskRun Instance
	// taskRun, tErr := c.sampleclientset.SamplecontrollerV1alpha1().TaskRuns(tr.Namespace).Create(context.TODO(), tr, metav1.CreateOptions{})
	// if tErr != nil {
	// 	fmt.Printf("Error Creating taskrun: %v", tErr)
	// }
	// fmt.Println("\n\nCreated TR - %v\n\n", taskRun.GetName())
	// // Creating Pods for the TaskRun CR
	// c.createTaskRunPods(taskRun)
	return nil
}

func (c *Controller) updateFooStatus(foo *v1alpha1.PipelineRun) error {

	fooCopy := foo.DeepCopy()
	foo.Status.Message = fooCopy.Spec.Message
	foo.Status.Count = fooCopy.Spec.Count
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.sampleclientset.SamplecontrollerV1alpha1().PipelineRuns(foo.Namespace).UpdateStatus(context.Background(), foo, metav1.UpdateOptions{})
	return err
}

// enqueueFoo takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *Controller) enqueuePR(obj interface{}) {
	log.Println("\nCR added in the Workqueue")
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	fmt.Printf("Key: %v", key)
	c.wq.Add(key)
	log.Println("\n Successfully CR added in the Workqueue")
}

// Create a Task Run Object

/*func createTaskRun(foo *v1alpha1.PipelineRun) *v1alpha1.TaskRun {
	// Parse the PipelineRun Object
	prName := foo.GetName()
	prNamespace := foo.GetNamespace()
	// fmt.Printf("\nName : %+v \n", prName)

	spec := foo.Spec
	// fmt.Printf("\nSpec : %+v \n", spec)
	// Parsing Spec value
	prMessage := spec.Message // Storing the message
	prCount := spec.Count     // Storing the count

	labels := map[string]string{
		"controller": foo.Name,
	}

	// creating TaskRun Custom Resource
	tr := &v1alpha1.TaskRun{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sde.dev/v1alpha1",
			Kind:       "TaskRun",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:       labels,
			GenerateName: fmt.Sprintf("%v-tr-%v", prName, foo.ObjectMeta.Generation),
			Namespace:    prNamespace,
			//UID:          types.UID("my-uid"),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(foo, v1alpha1.SchemeGroupVersion.WithKind("PipelineRun")),
			},
		},
		Spec: v1alpha1.TaskRunSpec{
			// Set fields in the spec of the custom resource object.
			Message: prMessage,
			Count:   prCount,
		},
	}
	return tr
}
*/
func (c *Controller) deletePR(obj interface{}) {
	// Finishing Obj
	c.wq.Done(obj)
}

// Creates the new pod with the specified template
/*func createPod(foo *v1alpha1.TaskRun) *corev1.Pod {
	labels := map[string]string{
		"controller": foo.Name,
	}
	fmt.Println("Message: ", foo.Spec.Message)
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    labels,
			Name:      fmt.Sprintf(foo.Name + "-" + strconv.Itoa(rand.Intn(10000000))),
			Namespace: foo.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(foo, v1alpha1.SchemeGroupVersion.WithKind("TaskRun")),
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyOnFailure,
			Containers: []corev1.Container{
				{
					Name:  "static-nginx",
					Image: "nginx:latest",
					Env: []corev1.EnvVar{
						{
							Name:  "MESSAGE",
							Value: foo.Spec.Message,
						},
					},
					Command: []string{
						"bin/sh", "-c", "echo 'Message: $(MESSAGE) !' && sleep 5",
					},
					Args: []string{
						"-c",
						"while true; do echo '$(MESSAGE)'; sleep 10; done",
					},
				},
			},
		},
	}
}
*/