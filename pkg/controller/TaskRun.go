package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/Senjuti256/crdextended/pkg/apis/sde.dev/v1alpha1"
	"github.com/kanisterio/kanister/pkg/poll"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
)

func (c *Controller) createPods(prun *v1alpha1.PipelineRun, trun *v1alpha1.TaskRun) error {
	// var podCreate, podDelete bool
	itr := trun.Spec.Count
	// Creates pod
	for i := 0; i < itr; i++ {
		nPod, err := c.kubeClient.CoreV1().Pods(trun.Namespace).Create(context.TODO(), newPod(trun), metav1.CreateOptions{})
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

// total number of 'Completed' pods
func (c *Controller) totalCreatedPods(trun *v1alpha1.TaskRun) int {
	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"controller": trun.Name,
		},
	}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}
	// TODO: Prefer using podLister to reduce the call to K8s API.
	pList, _ := c.kubeClient.CoreV1().Pods(trun.Namespace).List(context.TODO(), listOptions)

	createdPods := 0
	for _, pod := range pList.Items {
		if pod.ObjectMeta.DeletionTimestamp.IsZero() && pod.Status.Phase == "Succeeded" {
			createdPods++
		}
	}
	return createdPods
}

// If the pod doesn't switch to a running state within 5 minutes, shall report.
func (c *Controller) waitForPods(trun *v1alpha1.TaskRun) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	return poll.Wait(ctx, func(ctx context.Context) (bool, error) {
		completedPods := c.totalCreatedPods(trun)
		fmt.Println("Total running pods ", completedPods)

		if completedPods == trun.Spec.Count {
			return true, nil
		}
		return false, nil
	})
}

func (c *Controller) updateTRStatus(trun *v1alpha1.TaskRun) error {
	t, err := c.prClient.SamplecontrollerV1alpha1().TaskRuns(trun.Namespace).Get(context.Background(), trun.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	fmt.Println("about to update TR status", trun)
	klog.Infof("Insider updatetrunstatus: %v ,,, %v", c.totalCreatedPods(trun), t.Spec.Message, t.Spec.Count)

	t.Status.Count = c.totalCreatedPods(trun)
	t.Status.Message = t.Spec.Message
	_, err = c.prClient.SamplecontrollerV1alpha1().TaskRuns(trun.Namespace).UpdateStatus(context.Background(), t, metav1.UpdateOptions{})

	return err
}


//Deleting a particular TR CR all its corresponding pods get deleted.

/*func (c *Controller) deletePods(trun *v1alpha1.TaskRun) error {
	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"controller": trun.Name,
		},
	}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}
	// TODO: Prefer using podLister to reduce the call to K8s API.
	pList, _ := c.kubeClient.CoreV1().Pods(trun.Namespace).List(context.TODO(), listOptions)

	for _, pod := range pList.Items {
		if err := c.kubeClient.CoreV1().Pods(trun.Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{}); err != nil {
			return err
		}
		fmt.Printf("Pod %v deleted successfully!\n", pod.Name)
	}

	return nil
}

func (c *Controller) deleteTaskRun(obj interface{}) {
	trun, ok := obj.(*v1alpha1.TaskRun)
	if !ok {
		klog.Errorf("deleteTaskRun: Expected TaskRun but got %+v", obj)
		return
	}
	prunname := getPipelineRunNameFromTaskRun(trun)

	if prunname != "" {

		prun, err := c.prLister.PipelineRuns(trun.Namespace).Get(prunname)

		klog.Info("owner PipelineRun CR exists ")
		// Recreate TaskRun
		klog.Info("Recreating TaskRun CR")
		err = c.createTaskRun(prun)
		if err != nil {
			klog.Errorf("Failed to create TaskRun %s: %v", trun.Name, err)
		}

	}

	if err := c.deletePods(trun); err != nil {
		klog.Errorf("Failed to delete pods for TaskRun %v: %v", trun.Name, err)
	}

	if err := c.prClient.SamplecontrollerV1alpha1().TaskRuns(trun.Namespace).Delete(context.Background(), trun.Name, metav1.DeleteOptions{}); err != nil {
		klog.Errorf("Failed to delete TaskRun %v: %v", trun.Name, err)
	}
}
*/