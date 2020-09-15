package kube

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	Config    *rest.Config
	Clientset *kubernetes.Clientset
}

func NewClient(kubeConfigPath string) (*Client, error) {
	config, err := NewConfig(kubeConfigPath)
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build a kube clientset: %w", err)
	}

	return &Client{
		Config:    config,
		Clientset: clientSet,
	}, nil
}
