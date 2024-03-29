package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	RenderChi "github.com/go-chi/render"
	RenderPkg "github.com/unrolled/render"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var render *RenderPkg.Render
var waitGroup sync.WaitGroup

func main() {
	start()
}

func start() {
	contentType := middleware.AllowContentType("application/json")
	render = RenderPkg.New()
	route := chi.NewRouter()
	route.Use(middleware.RequestID)
	route.Use(middleware.RealIP)
	route.Use(middleware.Recoverer)
	route.Use(contentType)
	route.Use(RenderChi.SetContentType(RenderChi.ContentTypeJSON))
	route.Use(middleware.Timeout(60 * time.Second))

	route.Get("/go-app/health", health)
	route.Post("/node-app/pods/delete/name/{identifier}", controller)
	route.Post("/go-app/message/{podName}", sendMessageToPod)

	panic(http.ListenAndServe(":8081", route))
}

func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "App up")
}

type MessageRequest struct {
	Message string `json:"message"`
}

// Bind implements render.Binder.
func (*MessageRequest) Bind(r *http.Request) error {
	return nil
}

func sendMessageToPod(w http.ResponseWriter, r *http.Request) {

	clienthttp := &http.Client{}

	podName := chi.URLParam(r, "podName")
	podIp, podErr := getIpFromPodByItsName("default", podName)

	if podErr != nil {
		fmt.Fprintf(w, "Error getting pod IP: %v", podErr)
		return
	}

	url := fmt.Sprintf("http://%s:8080/node-app/", podIp)

	fmt.Println(url)

	message := &MessageRequest{}

	if err := RenderChi.Bind(r, message); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	bytesMessage, _ := json.Marshal(message)

	request, requestErr := http.NewRequest("POST", url, bytes.NewBuffer(bytesMessage))

	if requestErr != nil {
		fmt.Fprintf(w, requestErr.Error())
		return
	}

	response, responseErr := clienthttp.Do(request)

	if response != nil {
		fmt.Print(response)

		fmt.Fprintf(w, "Message propagated successfully")
		return

	}

	if responseErr != nil {
		fmt.Fprintf(w, responseErr.Error())
		return
	}

	fmt.Fprintf(w, "Request couldn't reach recipient")
}

func getIpFromPodByItsName(namespace, podName string) (string, error) {

	clientset, clientsetErr := kubeconfig()

	if clientsetErr != nil {
		return "", clientsetErr
	}

	podsClient := clientset.CoreV1().Pods(namespace)
	pod, getErr := podsClient.Get(context.TODO(), podName, metav1.GetOptions{})

	if getErr != nil {
		return "", getErr
	}

	return pod.Status.PodIP, nil
}

func controller(w http.ResponseWriter, r *http.Request) {

	podName := chi.URLParam(r, "identifier")

	err := deletePodByName(podName)

	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	fmt.Fprintf(w, "Pod deleted successfully")

}

func deletePodByName(podName string) error {

	clientset, clientsetErr := kubeconfig()

	if clientsetErr != nil {
		return clientsetErr
	}

	podsClient := clientset.CoreV1().Pods("default")

	deletePropagation := metav1.DeletePropagationForeground
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &deletePropagation,
	}

	deleteErr := podsClient.Delete(context.TODO(), podName, *deleteOptions)

	if deleteErr != nil {
		return deleteErr
	}

	scallingErr := decreasePods("default", "node-app")

	if scallingErr != nil {
		return scallingErr
	}

	return nil
}

func decreasePods(namespace, deploymentName string) error {
	clientset, clientsetErr := kubeconfig()

	if clientsetErr != nil {
		return clientsetErr
	}

	deploymentClient := clientset.AppsV1().Deployments(namespace)

	deployment, deplErr := deploymentClient.Get(context.TODO(), deploymentName, metav1.GetOptions{})

	if deplErr != nil {
		return errors.New(fmt.Sprintf("Error getting deployment. Reason: %v", deplErr.Error()))
	}

	*deployment.Spec.Replicas -= 1

	_, updateErr := deploymentClient.Update(context.TODO(), deployment, metav1.UpdateOptions{})

	if updateErr != nil {
		return errors.New(fmt.Sprintf("Error updating deployment. Reason: %v", updateErr.Error()))
	}

	return nil
}

func kubeconfig() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error config in cluster config. Reason: %v", err.Error()))
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error config kubernetes new config. Reason: %v", err.Error()))
	}

	return clientset, nil
}
