package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Config contains the server (the webhook) cert and key and CA.
type Config struct {
	CertFile string
	KeyFile  string
	CAFile   string
}

// Get a clientset with in-cluster config.
func getClient() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return clientset
}

func configTLS(config Config, clientset *kubernetes.Clientset) *tls.Config {
	sCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		log.Fatal(err)
	}

	clientCACert, err := ioutil.ReadFile(config.CAFile)
	if err != nil {
		log.Fatal(err)
	}

	clientCertPool := x509.NewCertPool()
	clientCertPool.AppendCertsFromPEM(clientCACert)

	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
		// TODO: uses mutual tls after we agree on what cert the apiserver should use.
		ClientAuth: tls.NoClientCert,
		RootCAs:    clientCertPool,
	}
}
