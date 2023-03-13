package controller

import "fmt"

func printcase(){
	fmt.Println("Inside PipelineRun controller")
}






/*package controller

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Senjuti256/crdextended/pkg/apis/sde.dev/v1alpha1"
	pClientSet "github.com/Senjuti256/crdextended/pkg/client/clientset/versioned"
	pInformer "github.com/Senjuti256/crdextended/pkg/client/informers/externalversions/sde.dev/v1alpha1"
	pLister "github.com/Senjuti256/crdextended/pkg/client/listers/sde.dev/v1alpha1"

	//"k8s.io/apimachinery/pkg/api/errors"
	//"k8s.io/apimachinery/pkg/api/errors"
	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//"k8s.io/apimachinery/pkg/fields"
	//"k8s.io/apimachinery/pkg/util/runtime"
	//"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by PipelineRun"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "PipelineRunResource synced successfully"
)

// Controller logic implementation
// TODO: Record events
type Controller struct {
	// K8s clientset
	kubeClient kubernetes.Interface
	// things required for controller:
	// - clientset for custom resource
	prClient pClientSet.Interface
	// - resource (informer) cache has synced
	prSync cache.InformerSynced
	// - interface provided by informer
	prLister pLister.PipelineRunLister

	prInformer pInformer.PipelineRunInformer
    trInformer pInformer.TaskRunInformer
	// taskrun specific lisers
	trLister pLister.TaskRunLister
	// - queue
	// stores the work that has to be processed, instead of performing
	// as soon as it's changed.
	// Helps to ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	wq workqueue.RateLimitingInterface
}

// returns a new TrackPod controller
func NewController(kubeClient kubernetes.Interface, prClient pClientSet.Interface, prInformer pInformer.PipelineRunInformer, trInformer pInformer.TaskRunInformer) *Controller {
	c := &Controller{
		kubeClient: kubeClient,
		prClient:   prClient,
		prSync:     prInformer.Informer().HasSynced,
		prLister:   prInformer.Lister(),
		trInformer: trInformer,
		trLister: trInformer.Lister(),
		wq:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "PipelineRun"),
	}

	// event handler when the pipelineRun resources are added/deleted/updated.
	prInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: c.handlePipelineAdd,
			UpdateFunc: func(old, obj interface{}) {
				oldpod := old.(*v1alpha1.PipelineRun)
				newpod := obj.(*v1alpha1.PipelineRun)
				if newpod == oldpod {
					return
				}
				c.handlePipelineAdd(obj)
			},
			DeleteFunc: c.handlePipelineDel,
		},
	)


	c.trInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
        
		/*DeleteFunc: func(obj interface{}) {
            // When a TaskRun is deleted, check if its owner PipelineRun still exists and recreate the TaskRun if it does
            taskRun := obj.(*v1alpha1.TaskRun)
            pipelineRun, err := c.getOwnerPipelineRun(taskRun)
            if err != nil {
                klog.Errorf("Failed to get owner PipelineRun for TaskRun %s: %v", taskRun.Name, err)
                return
            }
            if pipelineRun != nil {
                _, err := c.prClient.SamplecontrollerV1alpha1().TaskRuns(pipelineRun.Namespace).Create(context.TODO(), newTaskRun(pipelineRun), metav1.CreateOptions{})
				//c.prClient.SamplecontrollerV1alpha1().TaskRuns(taskRun.Namespace).Create(context.Background(), taskRun, metav1.CreateOptions{})
                if err != nil {
                    klog.Errorf("Failed to recreate TaskRun %s: %v", taskRun.Name, err)
                } else {
                    klog.Infof("Recreated TaskRun %s for PipelineRun %s", taskRun.Name, pipelineRun.Name)
                }
            }
        },*/
/*		DeleteFunc: c.handleTaskRunDel,
    })


		
	
	


	return c
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until ch
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(ch chan struct{}) error {
	defer c.wq.ShutDown()
   
   klog.Info("Starting PipelineRun Controller")

	// Wait for the caches to be synced before starting workers
	if ok := cache.WaitForCacheSync(ch, c.prSync); !ok {
		log.Println("failed to wait for cache to sync")
	}


   //addn 9/3
   if ok := cache.WaitForCacheSync(ch, c.trInformer.Informer().HasSynced); !ok {
	log.Println("failed to wait for tr cache to sync")
   }




	// Launch the goroutine for workers to process the CR
	klog.Info("Starting workers")
	go wait.Until(c.worker, time.Second, ch)
	klog.Info("Started workers")
	<-ch
	klog.Info("Shutting down the worker")

	return nil
}

// worker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue
func (c *Controller) worker() {
	for c.processNextItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextItem() bool {

	item, shutdown := c.wq.Get()
	if shutdown {
		klog.Info("Shutting down")
		return false
	}

	defer c.wq.Forget(item)
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		klog.Errorf("error while calling Namespace Key func on cache for item %s: %s", item, err.Error())
		return false
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Errorf("error while splitting key into namespace & name: %s", err.Error())
		return false
	}

	prun, err := c.prLister.PipelineRuns(ns).Get(name)
	if err != nil {
		klog.Errorf("%s, Getting the pipelinerun resource from lister.", err.Error())
		return false
	}


	trun, err := c.syncHandler(prun)
	if err != nil {
		klog.Errorf("Error creating the TaskRun for %s PipelineRun", prun.Name, err.Error())
		return false
	}

	if trun != nil {
		// wait for pods to be ready
		if err := c.waitForPods(trun); err != nil {
			klog.Errorf("error %s, waiting for pods to meet the expected state", err.Error())
		}

		// update taskrun status
		if err := c.updateTRStatus(trun); err != nil {
			klog.Errorf("error %s updating TaskRun status", err.Error())
		}

		// update pipelinerun status
		if err = c.updatePRStatus(prun, trun); err != nil {
			klog.Errorf("error %s updating PipelineRun status", err.Error())
		}

		if prun.Status.Count == c.totalCreatedPods(trun) {
			c.wq.Done(trun)
		}
	}


	return true
}

// syncHandler checks the status of pipeline, if current != desired, meets the requirement.
func (c *Controller) syncHandler(prun *v1alpha1.PipelineRun) (*v1alpha1.TaskRun, error) {
var createTR bool



fmt.Println("Checking if the desired state is met or not")
if prun.Spec.Message != prun.Status.Message || prun.Spec.Count != prun.Status.Count {
	fmt.Println("<<<<<<<<<<<<<< Mismatch found >>>>>>>>>>>> ")
	createTR = true
}

fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<>>>>>>>>>>>>>>>>>>>>>>> ", prun.Status, prun.Spec)



// create the taskrun CR.
if createTR {
		fmt.Println("Inside the condition to create TR custom resource")


		tr, err := c.prClient.SamplecontrollerV1alpha1().TaskRuns(prun.Namespace).Create(context.TODO(), newTaskRun(prun), metav1.CreateOptions{})
		if err != nil {
			klog.Errorf("TaskRun creation failed for Pipeline %s", prun.Name)
			return nil, err
		}

		if tr != nil {
			klog.Infof("Taskrun %s has been created for PipelineRun %s", tr.Name, prun.Name)
			if err := c.createPods(prun, tr); err != nil {
				klog.Errorf("Taskrun %s failed to create pods", tr.Name)
			}
		}

		return tr, nil
	}

	return nil, nil
}










// Creates the new TaskRun CR with the specified template
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

// Updates the status section of PipelineRun and TR CR
func (c *Controller) updatePRStatus(prun *v1alpha1.PipelineRun, trun *v1alpha1.TaskRun) error {

	p, err := c.prClient.SamplecontrollerV1alpha1().PipelineRuns(prun.Namespace).Get(context.Background(), prun.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	t, err := c.prClient.SamplecontrollerV1alpha1().TaskRuns(trun.Namespace).Get(context.Background(), trun.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	p.Status.Count = t.Status.Count
	p.Status.Message = t.Status.Message

	_, err = c.prClient.SamplecontrollerV1alpha1().PipelineRuns(prun.Namespace).UpdateStatus(context.Background(), p, metav1.UpdateOptions{})
	return err
}

func (c *Controller) handlePipelineAdd(obj interface{}) {
	klog.Info(" Inside handlePipelineAdd!!!")
	c.wq.Add(obj)
}

func (c *Controller) handlePipelineDel(obj interface{}) {
	klog.Info("Inside handlePipelineDel!!!")
	c.wq.Done(obj)
}

func (c *Controller) handleTaskRunDel(obj interface{}) {
	klog.Info("Inside handleTaskRunDel!!!")
	c.wq.Done(obj)
}

func (c *Controller) handleTaskRunAdd(obj interface{}) {
	klog.Info(" Inside handleTaskRunAdd!!!")
	c.wq.Add(obj)
}

*/








//Handling the case that when TR deleted intentionally it will be recreated as long as the PR CR exists

/*func (c *Controller) createTaskRun(prun *v1alpha1.PipelineRun) error {
	trun := &v1alpha1.TaskRun{
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

	_, err := c.prClient.SamplecontrollerV1alpha1().TaskRuns(trun.Namespace).Create(context.Background(), trun, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create TaskRun %s: %v", trun.Name, err)
	}

	klog.Infof("Created TaskRun %s for PipelineRun %s", trun.Name, prun.Name)
	return nil
}

func getPipelineRunNameFromTaskRun(trun *v1alpha1.TaskRun) string {

	klog.Info("Trying to get PipelineRun name from TaskRun CR")
	for _, ref := range trun.OwnerReferences {
		if ref.Kind == "PipelineRun" {
			return ref.Name
		}
	}
	return ""
}

func (c *Controller) createTaskRunOnDelete(obj interface{}) {
	klog.Info("Inside createTaskRunOnDelete fn")
	trun, ok := obj.(*v1alpha1.TaskRun)
	if !ok {
		return
	}

	klog.Info("Checking if the owner PipelineRun CR exists or not")
	prunName := getPipelineRunNameFromTaskRun(trun)
	if prunName == "" {
		return
	}

	// Check if PipelineRun exists
	prun, err := c.prLister.PipelineRuns(trun.Namespace).Get(prunName)
	if err != nil {
		klog.Errorf("Error getting PipelineRun %s for TaskRun %s deletion: %v", prunName, trun.Name, err)
		return
	}
	klog.Info("owner PipelineRun CR exists ")
	// Recreate TaskRun
	klog.Info("Recreating TaskRun CR")
	err = c.createTaskRun(prun)
	if err != nil {
		klog.Errorf("Failed to create TaskRun %s: %v", trun.Name, err)
	}
}

/*func (c *Controller) recreateTaskR(trun *v1alpha1.TaskRun) {
	ownerRef := metav1.GetControllerOf(trun)
	if ownerRef == nil {
		klog.Infof("TaskRun %s has no owner reference, skipping recreation", trun.Name)
		return
	}
	if ownerRef.Kind != "PipelineRun" {
		klog.Infof("TaskRun %s owner reference is not PipelineRun, skipping recreation", trun.Name)
		return
	}
	prun, err := c.prLister.PipelineRuns(trun.Namespace).Get(ownerRef.Name)
	if err != nil {
		klog.Errorf("Failed to get PipelineRun %s: %s", ownerRef.Name, err.Error())
		return
	}
	if prun == nil {
		klog.Infof("PipelineRun %s not found, skipping recreation of TaskRun %s", ownerRef.Name, trun.Name)
		return
	}
	if prun.DeletionTimestamp != nil {
		klog.Infof("PipelineRun %s is being deleted, skipping recreation of TaskRun %s", prun.Name, trun.Name)
		return
	}
	_, err = c.syncHandler(prun)
	if err != nil {
		klog.Errorf("Error recreating TaskRun for PipelineRun %s: %s", prun.Name, err.Error())
	}
}*/

/*func (c *Controller) recreateTaskRun(prun *v1alpha1.PipelineRun, trunName string) (*v1alpha1.TaskRun, error) {
	trun, err := c.prClient.SamplecontrollerV1alpha1().TaskRuns(prun.Namespace).Create(context.TODO(), &v1alpha1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
            Name:      trunName,
            Namespace: prun.Namespace,
            Labels: map[string]string{
                "pipelinerun": prun.Name,
            },
            OwnerReferences: []metav1.OwnerReference{
                *metav1.NewControllerRef(prun, v1alpha1.SchemeGroupVersion.WithKind("PipelineRun")),
            },
        },
    }, metav1.CreateOptions{})
*/

		/*ObjectMeta: metav1.ObjectMeta{
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
	},*/                   //
/*	if err != nil {
		klog.Errorf("Error creating TaskRun %s for PipelineRun %s: %s", trunName, prun.Name, err.Error())
		return nil, err
	}
	klog.Infof("TaskRun %s has been recreated for PipelineRun %s", trunName, prun.Name)

	if err := c.createPods(prun, trun); err != nil {
		klog.Errorf("Failed to create pods for TaskRun %s: %s", trunName, err.Error())
		return nil, err
	}

	return trun, nil
}
*/

/*func (c *Controller) getOwnerPipelineRun(taskRun *v1alpha1.TaskRun) (*v1alpha1.PipelineRun, error) {
    for _, ownerRef := range taskRun.OwnerReferences {
        if ownerRef.Kind == "PipelineRun" {
            pipelineRun, err := c.prInformer.Informer().GetIndexer().ByIndex("byName", fmt.Sprintf("%s/%s", taskRun.Namespace, ownerRef.Name))
            if err != nil {
                return nil, err
            }
            if len(pipelineRun) > 0 {
                return pipelineRun[0].(*v1alpha1.PipelineRun), nil
            }
        }
    }
    return nil, nil
}*/
