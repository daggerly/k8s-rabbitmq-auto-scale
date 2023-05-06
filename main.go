package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"net/http"

	api "github.com/prometheus/prometheus/web/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

const (
	AlertName  string = "messagePileUp"
	PromSVCURL string = "http://expose-prom.monitoring:9001/api/v1/alerts"
)

type AlertData struct {
	Alerts []api.Alert
}

type AlertResponse struct {
	Status string    `json:"status"`
	Data   AlertData `json:"data,omitempty"`
}

type AlertMem struct {
	Namespace      string
	DeploymentName string
}

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	toScaleDownAlert := make(map[string]*AlertMem)
	for {
		time.Sleep(time.Duration(10 * time.Second))

		c := http.Client{Timeout: time.Duration(5 * time.Second)}
		res, err := c.Get(PromSVCURL)
		if err != nil {
			panic(err.Error())
		}
		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		var resData AlertResponse
		err = json.Unmarshal(body, &resData)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("%+v\n", resData)

		const (
			BaseValue          = 3
			ConsumerLowNumber  = 3
			ConsumerHighNumber = 10
		)

		if len(resData.Data.Alerts) == 0 {
			fmt.Println("no alert")
			for alertName, alertMem := range toScaleDownAlert {
				fmt.Printf("%s to scale down\n", alertName)
				deployments := clientset.AppsV1().Deployments(alertMem.Namespace)
				scale, err := deployments.GetScale(context.TODO(), alertMem.DeploymentName, metav1.GetOptions{})
				if err != nil {
					fmt.Printf("Failed to get scale subresource: %v", err)
					panic(err.Error())
				}
				if scale.Spec.Replicas <= ConsumerLowNumber {
					fmt.Printf("No Need to scale down %s\n", scale.Name)
					delete(toScaleDownAlert, alertName)
					continue
				}
				scale.ResourceVersion = ""
				scale.Spec.Replicas = ConsumerLowNumber
				scale, err = deployments.UpdateScale(context.TODO(), alertMem.DeploymentName, scale, metav1.UpdateOptions{})
				if err != nil {
					fmt.Println("Failed to scale subresource:")
					panic(err.Error())
				}
				fmt.Printf("scaled down %s success\n", scale.Name)
				delete(toScaleDownAlert, alertName)
			}

			continue
		}

		for _, alert := range resData.Data.Alerts {
			alertName := alert.Labels.Get("alertname")
			if alertName != AlertName || alert.State != "firing" {
				fmt.Printf("alert %s is not target\n", alertName)
				continue
			}
			deploymentName := "consumer"
			namespace := alert.Labels.Get("namespace")

			if _, ok := toScaleDownAlert[alertName]; ok {
				fmt.Printf("%s already scaled up, waiting for consume\n", alertName)
				continue
			}

			fmt.Printf("%s not in mem, create\n", alertName)
			alertMem := AlertMem{
				Namespace:      namespace,
				DeploymentName: deploymentName,
			}
			toScaleDownAlert[alertName] = &alertMem

			queueName := alert.Labels.Get("queue")
			fmt.Printf("queue pile up: %s. to scale up deploy %s\n", queueName, deploymentName)

			deployments := clientset.AppsV1().Deployments(namespace)

			scale, err := deployments.GetScale(context.TODO(), deploymentName, metav1.GetOptions{})
			if err != nil {
				fmt.Printf("Failed to get scale subresource: %v", err)
				panic(err.Error())
			}
			fmt.Printf("now scale: %+v\n", *scale)
			if scale.Spec.Replicas >= ConsumerHighNumber {
				fmt.Printf("Replicas already scaled to %d, not scale\n", scale.Spec.Replicas)
				continue
			}
			scale.ResourceVersion = ""
			scale.Spec.Replicas = ConsumerHighNumber
			scale, err = deployments.UpdateScale(context.TODO(), deploymentName, scale, metav1.UpdateOptions{})
			if err != nil {
				fmt.Println("Failed to scale subresource:")
				panic(err.Error())
			}
		}
	}
}
