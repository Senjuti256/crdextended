/*
package controller

import (
	"context"
	"fmt"
	"log"

	//"os"
	//"os/signal"
	//"syscall"
	"time"

	v1alpha1 "github.com/Senjuti256/crdextended/pkg/apis/sde.dev/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	//"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	//"k8s.io/client-go/kubernetes/scheme"
	//"k8s.io/client-go/rest"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
    cLister "github.com/Senjuti256/crdextended/pkg/client/listers/sde.dev/v1alpha1"
)

const (
    prKind       = "PipelineRun"
    trKind       = "TaskRun"
    prAPIVersion = "sde.dev/v1alpha1"
    trAPIVersion = "sde.dev/v1alpha1"
)

var (
    prGroupVersion = schema.GroupVersion{Group: "sde.dev", Version: "v1alpha1"}
    trGroupVersion = schema.GroupVersion{Group: "sde.dev", Version: "v1alpha1"}
)

type controller struct {
    kubeClient     kubernetes.Interface
    dynamicClient  dynamic.Interface
    prInformer     cache.SharedIndexInformer
    trInformer     cache.SharedIndexInformer
    prQueue        workqueue.RateLimitingInterface
    trQueue        workqueue.RateLimitingInterface
    prLister       cLister.PipelineRunLister
    trLister       cLister.TaskRunLister
}
/*
func NewController(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) *controller {
    // Create informer for PR custom resource
    prInformer := cache.NewSharedIndexInformer(
        &cache.ListWatch{
            ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
                return dynamicClient.Resource(schema.GroupVersionResource{
                    Group:    prGroupVersion.Group,
                    Version:  prGroupVersion.Version,
                    Resource: prKind,
                }).List(context.Background(), options)
            },
            WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
                return dynamicClient.Resource(schema.GroupVersionResource{
                    Group:    prGroupVersion.Group,
                    Version:  prGroupVersion.Version,
                    Resource: prKind,
                }).Watch(context.Background(), options)
            },
        },
        &v1alpha1.PipelineRun{},
        time.Second*30,
        cache.Indexers{},
    )

    // Create informer for TR custom resource
    trInformer := cache.NewSharedIndexInformer(
        &cache.ListWatch{
            ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
                return dynamicClient.Resource(schema.GroupVersionResource{
                    Group:    trGroupVersion.Group,
                    Version:  trGroupVersion.Version,
                    Resource: trKind,
}).List(context.Background(), options)
},
WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
return dynamicClient.Resource(schema.GroupVersionResource{
Group: trGroupVersion.Group,
Version: trGroupVersion.Version,
Resource: trKind,
}).Watch(context.Background(), options)
},
},
&v1alpha1.TaskRun{},
time.Second*30,
cache.Indexers{},
)
// Create workqueue for PR custom resource
prQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

// Create workqueue for TR custom resource
trQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

// Create controller object
c := &controller{
    kubeClient:     kubeClient,
    dynamicClient:  dynamicClient,
    prInformer:     prInformer,
    trInformer:     trInformer,
    prQueue:        prQueue,
    trQueue:        trQueue,
}

// Set event handlers for PR informer
prInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
    AddFunc:    c.handlePRAdd,
    UpdateFunc: c.handlePRUpdate,
    DeleteFunc: c.handlePRDelete,
})

// Set event handlers for TR informer
trInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
    AddFunc:    c.handleTRAdd,
    UpdateFunc: c.handleTRUpdate,
    DeleteFunc: c.handleTRDelete,
})

return c
}
*/
// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until ch
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
/*func (c *controller) Run(ch chan struct{}) error {
	defer c.prQueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting the Customcluster controller\n")
	klog.Info("Waiting for informer caches to sync\n")

	// Wait for the caches to be synced before starting workers
	if ok := cache.WaitForCacheSync(ch, c.prInformer.HasSynced); !ok {
		log.Println("failed to wait for cache to sync")
	}
	// Launch the goroutine for workers to process the CR
	klog.Info("Starting workers\n")
	go wait.Until(c.worker, time.Second, ch)
	klog.Info("Started workers\n")
	<-ch
	klog.Info("Shutting down the worker")

	return nil
}

// worker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue
func (c *controller) worker() {
	for c.processNextItem() {
	}
}

func (c *controller) processNextItem() bool {
    // Get the next item from the PR queue
    key, quit := c.prQueue.Get()
    if quit {
        return false
    }

    // Ensure that the key is valid
    if key == "" {
        c.prQueue.Forget(key)
        return true
    }
    // Get the corresponding PR custom resource
    pr, err := c.prLister.Get(key)
    if err != nil {
        // Handle the error and requeue the item
        if errors.IsNotFound(err) {
            // PR resource no longer exists, so forget the item
            c.prQueue.Forget(key)
            return true
        } else {
            // Requeue the item
            c.prQueue.AddRateLimited(key)
            return true
        }
    }
// Create or update the TR custom resource based on the existence of an existing TR
tr, err := c.trLister.TaskRuns(pr.Namespace).Get(pr.Name)
if errors.IsNotFound(err) {
    // TR resource doesn't exist, so create it
    tr = &v1alpha1.TaskRun{
        ObjectMeta: metav1.ObjectMeta{
            Name:      pr.Name,
            Namespace: pr.Namespace,
        },
        Spec: v1alpha1.TaskRunSpec{
            Message: pr.Spec.Message,
        },
    }

    _, err = c.dynamicClient.Samplecontrollerv1alpha1().TRs(pr.Namespace).Create(context.Background(), tr, metav1.CreateOptions{})
    if err != nil {
        // Requeue the item
        c.prQueue.AddRateLimited(key)
        return true
    }
} else {
    // TR resource already exists, so update it
    tr.Spec.Message = pr.Spec.Message

    _, err = c.dynamicClient.SamplecontrollerV1alpha1().TRs(pr.Namespace).Update(context.Background(), tr, metav1.UpdateOptions{})
    if err != nil {
        // Requeue the item
        c.prQueue.AddRateLimited(key)
        return true
    }
}

// Forget the item from the PR queue
c.prQueue.Forget(key)
return true
}

// handlePRAdd handles the "add" event for the PR custom resource
func (c *controller) handlePRAdd(obj interface{}) {
	pr := obj.(*v1alpha1.PipelineRun)
	fmt.Printf("Received PR add event: %v\n", pr)
	// Check if a corresponding TR custom resource exists
tr, err := c.dynamicClient.Resource(schema.GroupVersionResource{
    Group:    trGroupVersion.Group,
    Version:  trGroupVersion.Version,
    Resource: trKind,
}).Get(context.Background(), pr.Name, metav1.GetOptions{})
if err != nil {
    if errors.IsNotFound(err) {
        // Create new TR custom resource
        tr := &v1alpha1.TaskRun{
            ObjectMeta: metav1.ObjectMeta{
                Name: pr.Name,
            },
        }
        _, err := c.dynamicClient.Resource(schema.GroupVersionResource{
            Group:    trGroupVersion.Group,
            Version:  trGroupVersion.Version,
            Resource: trKind,
        }).Create(context.Background(), tr, metav1.CreateOptions{})
        if err != nil {
            fmt.Printf("Error creating TR custom resource: %v\n", err)
            c.prQueue.AddRateLimited(obj)
        } else {
            fmt.Printf("Created TR custom resource: %v\n", tr)
        }
    } else {
        fmt.Printf("Error getting TR custom resource: %v\n", err)
        c.prQueue.AddRateLimited(obj)
    }
} else {
    // Update existing TR custom resource
    tr.SetAnnotations(pr.Annotations)
    _, err := c.dynamicClient.Resource(schema.GroupVersionResource{
        Group:    trGroupVersion.Group,
        Version:  trGroupVersion.Version,
        Resource: trKind,
    }).Update(context.Background(), tr, metav1.UpdateOptions{})
    if err != nil {
        fmt.Printf("Error updating TR custom resource: %v\n", err)
        c.prQueue.AddRateLimited(obj)
    } else {
        fmt.Printf("Updated TR custom resource: %v\n", tr)
    }
}
}
// PR custom resource
func (c *controller) handlePRUpdate(oldObj, newObj interface{}) {
	oldPR := oldObj.(*v1alpha1.PipelineRun)
	newPR := newObj.(*v1alpha1.PipelineRun)
	fmt.Printf("Received PR update event: old=%v, new=%v\n", oldPR, newPR)
	// Check if a corresponding TR custom resource exists
tr, err := c.dynamicClient.Resource(schema.GroupVersionResource{
    Group:    trGroupVersion.Group,
    Version:  trGroupVersion.Version,
    Resource: trKind,
}).Get(context.Background(), newPR.Name, metav1.GetOptions{})
if err != nil {
    if errors.IsNotFound(err) {
        // Create new TR custom resource
        tr := &v1alpha1.TaskRun{
            ObjectMeta: metav1.ObjectMeta{
                Name: newPR.Name,
            },
        }
        _, err := c.dynamicClient.Resource(schema.GroupVersionResource{
            Group:    trGroupVersion.Group,
            Version:  trGroupVersion.Version,
            Resource: trKind,
        }).Create(context.Background(), tr, metav1.CreateOptions{})
        if err != nil {
            fmt.Printf("Error creating TR custom resource: %v\n", err)
            c.prQueue.AddRateLimited(newObj)
        } else {
            fmt.Printf("Created TR custom resource: %v\n", tr)
        }
    } else {
        fmt.Printf("Error getting TR custom resource: %v\n", err)
        c.prQueue.AddRateLimited(newObj)
    }
} else {
    // Update existing TR custom resource
    tr.SetAnnotations(newPR.Annotations)
    _, err := c.dynamicClient.Resource(schema.GroupVersionResource{
        Group:    trGroupVersion.Group,
        Version:  trGroupVersion.Version,
        Resource: trKind,
    }).Update(context.Background(), tr, metav1.UpdateOptions{})
    if err != nil {
        fmt.Printf("Error updating TR custom resource: %v\n", err)
        c.prQueue.AddRateLimited(newObj)
    } else {
        fmt.Printf("Updated TR custom resource: %v\n", tr)
    }
}
}
// handlePRDelete handles the "delete" event for the PR custom resource
func (c *controller) handlePRDelete(obj interface{}) {
pr := obj.(*v1alpha1.PipelineRun)
fmt.Printf("Received PR delete event: %v\n", pr)
// Delete corresponding TR custom resource
err := c.dynamicClient.Resource(schema.GroupVersionResource{
    Group:    trGroupVersion.Group,
    Version:  trGroupVersion.Version,
    Resource: trKind,
}).Delete(context.Background(), pr.Name, metav1.DeleteOptions{})
if err != nil {
    if errors.IsNotFound(err) {
        fmt.Printf("TR custom resource %s not found: %v\n", pr.Name, err)
    } else {
        fmt.Printf("Error deleting TR custom resource %s: %v\n", pr.Name, err)
        c.prQueue.AddRateLimited(obj)
    }
} else {
    fmt.Printf("Deleted TR custom resource %s\n", pr.Name)
}
}
// handleTRAdd handles the "add" event for the TR custom resource
func (c *controller) handleTRAdd(obj interface{}) {
    tr := obj.(*v1alpha1.TaskRun)
    fmt.Printf("Received TR add event: %v\n", tr)
    // Create pod
pod := &corev1.Pod{
    ObjectMeta: metav1.ObjectMeta{
        Name:      tr.Name,
        Namespace: tr.Namespace,
        Labels: map[string]string{
            "app": "echo-pod",
        },
    },
    Spec: corev1.PodSpec{
        Containers: []corev1.Container{
            {
                Name:  "echo-container",
                Image: "busybox",
                Command: []string{
                    "/bin/sh",
                    "-c",
                    fmt.Sprintf("echo %s && sleep 60", tr.Spec.Message),
                },
            },
        },
        RestartPolicy: corev1.RestartPolicyNever,
    },
}

pod, err := c.kubeClient.CoreV1().Pods(tr.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
if err != nil {
    fmt.Printf("Error creating pod for TR %s: %v\n", tr.Name, err)
} else {
    fmt.Printf("Created pod for TR %s: %s\n", tr.Name, pod.Name)
}

c.trQueue.Forget(tr)
}
// handleTRUpdate handles the "update" event for the TR custom resource
func (c *controller) handleTRUpdate(oldObj, newObj interface{}) {
    oldTR := oldObj.(*v1alpha1.TaskRun)
    newTR := newObj.(*v1alpha1.TaskRun)
    fmt.Printf("Received TR update event: old=%v, new=%v\n", oldTR, newTR)
    // Do nothing
    c.trQueue.Forget(newTR)
    }
    
    // handleTRDelete handles the "delete" event for the TR custom resource
    func (c *controller) handleTRDelete(obj interface{}) {
    tr := obj.(*v1alpha1.TaskRun)
    fmt.Printf("Received TR delete event: %v\n", tr)
    // Delete pod if it exists
    err := c.kubeClient.CoreV1().Pods(tr.Namespace).Delete(context.Background(), tr.Name, metav1.DeleteOptions{})
    if err != nil {
        fmt.Printf("Error deleting pod for TR %s: %v\n", tr.Name, err)
    } else {
        fmt.Printf("Deleted pod for TR %s\n", tr.Name)
    }
    
    c.trQueue.Forget(tr)
    }

    */

    package controller

    import "fmt"

    func main(){
        fmt.Println("Controller")
    }
