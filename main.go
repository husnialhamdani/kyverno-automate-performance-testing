package main

/*
CREATE (done)
LIST (done)
UPDATE ()
DELETE (done)
*/
import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"time"

	appsv1 "k8s.io/api/apps/v1"
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

	nodeList, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, n := range nodeList.Items {
		fmt.Println(n.Name)
	}

	/* netpolSpec := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "netpol-ingress-default",
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{},
		},
	}

	deploymentSpec := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deployment-1",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
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

	podSpec := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod-1",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "busybox", Image: "busybox:latest", Command: []string{"sleep", "100000"}},
			},
		},
	}

	cronjobSpec := &v1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cronjob-1",
		},
		Spec: v1beta1.CronJobSpec{
			Schedule: "* * * * *",
			JobTemplate: v1beta1.JobTemplateSpec{
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

	configmapSpec := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cm-1",
		},
		Data: map[string]string{"color.good": "purple", "color.bad": "yellow"},
	}

	secretSpec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "secret-1",
		},
		Data: map[string][]byte{"test": []byte("test")},
	}

	pod, err := clientset.CoreV1().Pods("default").Create(context.Background(), podSpec, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	configMap, err := clientset.CoreV1().ConfigMaps("default").Create(context.Background(), configmapSpec, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	deployment, err := clientset.AppsV1().Deployments("default").Create(context.Background(), deploymentSpec, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	cronjob, err := clientset.BatchV1beta1().CronJobs("default").Create(context.Background(), cronjobSpec, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	secret, err := clientset.CoreV1().Secrets("default").Create(context.Background(), secretSpec, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	networkpolicy, err := clientset.NetworkingV1().NetworkPolicies("default").Create(context.Background(), netpolSpec, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println(pod.GetName())
	fmt.Println(configMap.GetName())
	fmt.Println(deployment.GetName())
	fmt.Println(cronjob.GetName())
	fmt.Println(secret.GetName())
	fmt.Println(networkpolicy.GetName()) */

	//Create resources
	label := make(map[string]string)
	label["app"] = "web"
	/* for i := 0; i < 5; i++ {
		name := "my-deployment" + strconv.Itoa(i)
		CreateDeployment(*clientset, name, "default", "nginx:latest", label)
	} */

	//Delete resources
	//for i := 0; i <= 2; i++ {å
	//	name := "my-deployment" + strconv.Itoa(i)
	//	DeleteDeployment(*clientsået, name, "default")
	//}
	//DeleteDeployment(*clientset, "my-deployment-1", "default")

	/*
		The scenario:
		- Create Cluster (automation.sh)
		- Install metrics server (automation.sh)
		- Install Kyverno & the policies
		- Monitoring started (0m)
		- Create loads of resources(1m)
		- Delete half of resources
		- Monitoring ended
	*/

	mc, err := metrics.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	//Monitoring
	var memoryUsage []int
	duration := 1
	interval := 10

	podmetricGet, err := mc.MetricsV1beta1().PodMetricses(metav1.NamespaceDefault).Get(context.Background(), "nginx", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println("-------")
	memQuantity, ok := podmetricGet.Containers[0].Usage.Memory().AsInt64()
	if !ok {
		return
	}

	sendMail("hello there")

	for len(memoryUsage) < (int(duration) * 60 / interval) {
		memoryUsage = append(memoryUsage, int(memQuantity)/1000000)
		fmt.Println(memoryUsage)
		time.Sleep(time.Duration(interval) * time.Second)
	}

}

func int32Ptr(i int32) *int32 { return &i }

func CreateDeployment(clientset kubernetes.Clientset, name string, namespace string, image string, label map[string]string) {
	deploymentSpec := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
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
		panic(err)
	}
	fmt.Println("deployment sucessfully created:", deployment.GetName())
}

func DeleteDeployment(clientset kubernetes.Clientset, name string, namespace string) {
	deletePolicy := metav1.DeletePropagationForeground
	if err := clientset.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deployment deleted:", name)
}

func sendMail(body string) {
	from := ""
	pass := ""
	to := ""

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Hello there\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Print("email sent")
}
