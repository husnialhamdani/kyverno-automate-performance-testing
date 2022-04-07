#!/bin/bash

iteratorStart=()
iteratorEnd=()
resourcetoCreate=("ConfigMap" "Secret" "Deployment" "Cronjob" "Pod")


function executeCommand() {
    #create resource
    for((i=$1;i<=$2;i++))
    do 
        case $3 in
            ConfigMap) kubectl create configmap my-config-$i --from-file=ssh-privatekey=/root/.ssh/authorized_keys -n $4;;
            Secret) kubectl create secret generic my-secret-$i --from-literal=username=admin --from-literal=password='admin' -n $4;;
            Deployment) kubectl create deployment my-deployment-$i --image=nginx -n $4;;
            Pod) kubectl run pod-$i --image=nginx -n $4;;
            Cronjob) kubectl create cronjob my-cronjob-$i --image=busybox --schedule="* * * * *" -- "/bin/sh" "-c" "sleep 60" -n $4;;
        esac
    done
}

#Cluster setup & monitoring tools
#minikube start
sysctl -w fs.inotify.max_queued_events=1048576
sysctl -w fs.inotify.max_user_watches=1048576
sysctl -w fs.inotify.max_user_instances=1048576
kind create cluster --config config.yaml
kubectl apply -f metricserver.yaml

#Install prometheus and grafana
git clone https://github.com/prometheus-operator/kube-prometheus.git
kubectl apply --server-side -f kube-prometheus/manifests/setup
kubectl apply -f kube-prometheus/manifests/

#Install Kyverno & the policies
helm install kyverno kyverno/kyverno --namespace kyverno --create-namespace
helm install kyverno-policies kyverno/kyverno-policies --namespace kyverno

#Set target namespace
echo "Target namespace:"
read namespace
kubectl create ns $namespace

#Get number of resource that are going to create
for l in ${!resourcetoCreate[@]]}; do
    echo "Resource to create for" ${resourcetoCreate[l]}":"
    echo "start:"; read start; iteratorStart+=($start)
    echo "end:"; read end; iteratorEnd+=($end)
done

#Execute the command in range of user defined
for i in ${!resourcetoCreate[@]]}; do
    executeCommand ${iteratorStart[$i]} ${iteratorEnd[$i]} ${resourcetoCreate[i]} $namespace
done