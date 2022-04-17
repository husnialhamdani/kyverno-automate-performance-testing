package main

/*
	The scenario:
	- Create Cluster (automation.sh)
	- Install metrics server (automation.sh)
	- Install Kyverno & the policies (automation.sh)
	- Monitoring started (0m) (this script)
	- Create loads of resources (1m) (this script)
	- Delete half of resources (this script)
	- Monitoring ended


	The scale:
	small: 500
	medium: 1000
	large: 2000
	extra large: 3000

	The CLI:
	...
*/

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/gomail.v2"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

func main() {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		panic(err)
	}
	clientset := kubernetes.NewForConfigOrDie(config)

	//Scales mapping
	scales := map[string]int{"xs": 100, "small": 500, "medium": 1000, "large": 2000, "xl": 3000}
	scalesPtr := flag.String("scales", "xs", "choose the scale size (small/medium/large/xl) default: xs")

	flag.Parse()
	log.Print("scales selected: ", *scalesPtr, ": ", scales[*scalesPtr])
	size := scales[*scalesPtr] / 5

	//Get usage
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go getMetrics(wg, 10, 10, "kyverno-856b9c4fbd-fxqqs", "kyverno")

	//dependencies
	label := map[string]string{"app": "web"}
	namespace := "default"

	//Create resources - steps up
	fmt.Println("Creating resource..")
	for i := 0; i < size; i++ {
		counter := strconv.Itoa(i)
		createNamespace(*clientset, counter)
		createDeployment(*clientset, counter, namespace, "nginx:latest", label)
		createConfigmap(*clientset, counter, namespace)
		createPod(*clientset, counter, namespace, "nginx")
		createSecret(*clientset, counter, namespace)
		createCronjob(*clientset, counter, namespace, "* * * * *")
	}

	time.Sleep(time.Duration(10) * time.Minute)

	//Delete resources - steps down
	fmt.Println("Deleting resource..")
	for i := size - 1; i >= size/2; i-- {
		counter := strconv.Itoa(i)
		deleteNamespace(*clientset, counter)
		deleteDeployment(*clientset, counter, namespace)
		deleteConfigmap(*clientset, counter, namespace)
		deletePod(*clientset, counter, namespace)
		deleteSecret(*clientset, counter, namespace)
		deleteCronjob(*clientset, counter, namespace)
	}

	wg.Wait()
	visualizeAnomaly()

	//cleanup(*clientset, size, "default")
	sendReport(os.Getenv("EMAILFROM"), os.Getenv("EMAILTO"), "Kyverno Automation Performance Testing report")
}

func getMetrics(wg *sync.WaitGroup, duration int, interval int, name string, namespace string) {
	defer wg.Done()
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		panic(err)
	}

	mc, err := metrics.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	var memoryUsage [][]int
	for len(memoryUsage) < (int(duration) * 60 / interval) {
		podmetricGet, err := mc.MetricsV1beta1().PodMetricses(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			panic(err)
		}
		memQuantity, ok := podmetricGet.Containers[0].Usage.Memory().AsInt64()
		if !ok {
			panic(!ok)
		}
		memoryUsage = append(memoryUsage, []int{len(memoryUsage), int(memQuantity) / 1000000})
		fmt.Println(memoryUsage)
		time.Sleep(time.Duration(interval) * time.Second)
	}

	csvfile, err := os.Create("usage.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	csvwriter := csv.NewWriter(csvfile)
	for _, row := range memoryUsage {
		st := strings.Fields(strings.Trim(fmt.Sprint(row), "[]"))
		_ = csvwriter.Write(st)
	}
	csvwriter.Flush()
	csvfile.Close()

}

func cleanup(clientset kubernetes.Clientset, size int, namespace string) {
	log.Print("Cleaning up resources...")
	for i := size - 1; i >= 0; i-- {
		counter := strconv.Itoa(i)
		deleteNamespace(clientset, counter)
		deleteDeployment(clientset, counter, namespace)
		deleteConfigmap(clientset, counter, namespace)
		deletePod(clientset, counter, namespace)
		deleteSecret(clientset, counter, namespace)
		deleteCronjob(clientset, counter, namespace)
	}
}

func int32Ptr(i int32) *int32 { return &i }

func createNamespace(clientset kubernetes.Clientset, name string) {
	nsSpec := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "namespace-" + name,
		},
	}
	namespace, err := clientset.CoreV1().Namespaces().Create(context.Background(), nsSpec, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Namespace successfully created:", namespace.GetName())
}

func deleteNamespace(clientset kubernetes.Clientset, name string) {
	deletePolicy := metav1.DeletePropagationForeground
	if err := clientset.CoreV1().Namespaces().Delete(context.TODO(), "namespace-"+name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Namespace deleted:", "namespace-"+name)
}

func createConfigmap(clientset kubernetes.Clientset, name string, namespace string) {
	configmapSpec := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm-" + name,
			Namespace: namespace,
		},
		Data: map[string]string{"color.good": "purple", "color.bad": "yellow"},
	}
	configMap, err := clientset.CoreV1().ConfigMaps("default").Create(context.Background(), configmapSpec, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ConfigMap created:", configMap.GetName())
}

func deleteConfigmap(clientset kubernetes.Clientset, name string, namespace string) {
	deletePolicy := metav1.DeletePropagationForeground
	if err := clientset.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), "cm-"+name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ConfigMap deleted:", "cm-"+name)
}

func createSecret(clientset kubernetes.Clientset, name string, namespace string) {
	secretSpec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-" + name,
			Namespace: namespace,
		},
		Data: map[string][]byte{"test": []byte("test")},
	}

	secret, err := clientset.CoreV1().Secrets("default").Create(context.Background(), secretSpec, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Secret created", secret.GetName())
}

func deleteSecret(clientset kubernetes.Clientset, name string, namespace string) {
	deletePolicy := metav1.DeletePropagationForeground
	if err := clientset.CoreV1().Secrets(namespace).Delete(context.TODO(), "secret-"+name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Secret deleted:", "secret-"+name)
}

func createPod(clientset kubernetes.Clientset, name string, namespace string, image string) {
	podSpec := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-" + name,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "busybox", Image: "busybox:latest", Command: []string{"sleep", "100000"}},
			},
		},
	}
	pod, err := clientset.CoreV1().Pods("default").Create(context.Background(), podSpec, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Pod created:", pod.GetName())
}

func deletePod(clientset kubernetes.Clientset, name string, namespace string) {
	deletePolicy := metav1.DeletePropagationForeground
	if err := clientset.CoreV1().Pods(namespace).Delete(context.TODO(), "pod-"+name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Pod deleted:", "pod-"+name)
}

func createCronjob(clientset kubernetes.Clientset, name string, namespace string, schedule string) {
	cronjobSpec := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cronjob-" + name,
			Namespace: namespace,
		},
		Spec: batchv1.CronJobSpec{
			Schedule: schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Name: "busybox", Image: "busybox", Command: []string{"sleep", "60"}},
							},
							RestartPolicy: "OnFailure",
						},
					},
				},
			},
		},
	}
	cronjob, err := clientset.BatchV1().CronJobs("default").Create(context.Background(), cronjobSpec, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Cronjob created :", cronjob.GetName())
}

func deleteCronjob(clientset kubernetes.Clientset, name string, namespace string) {
	deletePolicy := metav1.DeletePropagationForeground
	if err := clientset.BatchV1().CronJobs(namespace).Delete(context.TODO(), "cronjob-"+name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Cronjob deleted:", "cronjob-"+name)
}

func createDeployment(clientset kubernetes.Clientset, name string, namespace string, image string, label map[string]string) {
	deploymentSpec := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deployment-" + name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: label,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: label,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "web",
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	deployment, err := clientset.AppsV1().Deployments("default").Create(context.Background(), deploymentSpec, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("deployment sucessfully created:", deployment.GetName())
}

func deleteDeployment(clientset kubernetes.Clientset, name string, namespace string) {
	deletePolicy := metav1.DeletePropagationForeground
	if err := clientset.AppsV1().Deployments(namespace).Delete(context.TODO(), "deployment-"+name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Deployment deleted:", "deployment-"+name)
}

func visualizeAnomaly() {
	cmd := exec.Command("python3", "anomalydetection.py")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	go copyOutput(stdout)
	go copyOutput(stderr)

	cmd.Wait()
	fmt.Println("report generated in report.png")
}

func copyOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func sendReport(from string, to string, subject string) {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", "Kyverno Automation Performance Testing result:")
	m.Attach("report.png")

	d := gomail.NewPlainDialer("smtp.gmail.com", 587, from, os.Getenv("EMAILPASS"))

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	} else {
		fmt.Println("email sent")
	}
}
