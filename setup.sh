#!/bin/bash

#Cluster setup & monitoring tools
sysctl -w fs.inotify.max_queued_events=1048576
sysctl -w fs.inotify.max_user_watches=1048576
sysctl -w fs.inotify.max_user_instances=1048576
kind create cluster --config dir/cluster-config/kind-config.yaml
kubectl apply -f dir/cluster-config/metricserver.yaml

#Install prometheus and grafana
git clone https://github.com/prometheus-operator/kube-prometheus.git
kubectl apply --server-side -f kube-prometheus/manifests/setup
kubectl apply -f kube-prometheus/manifests/

#Install Kyverno & the policies
helm install kyverno kyverno/kyverno --namespace kyverno --create-namespace
helm install kyverno-policies kyverno/kyverno-policies --namespace kyverno