package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	// TODO: try this library to see if it generates correct json patch
	// https://github.com/mattbaird/jsonpatch
)

const (
	// MonbanKey is state of lock
	MonbanKey = "koudaiii/monban"
)

// you need clientset to get Namespace from deployments
var clientset *kubernetes.Clientset

// toAdmissionResponse decide response Result and Message
func toAdmissionResponse(allowed bool, err error) *v1beta1.AdmissionResponse {
	response := &v1beta1.AdmissionResponse{
		Allowed: allowed,
	}
	if err != nil {
		response.Result = &metav1.Status{
			Message: err.Error(),
		}
	}
	return response
}

// Disallow edit deployments from specific annotations.
func admitDeployments(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	log.Println("admitting deployments")
	if expect, actual := "deployments", ar.Request.Resource.Resource; expect != actual {
		err := fmt.Errorf("unexpected resource: expect %s, actual %s", expect, actual)
		log.Println(err)
		return toAdmissionResponse(true, nil)
	}

	name, namespace, operation := ar.Request.Name, ar.Request.Namespace, ar.Request.Operation
	log.Printf("operation: %s, name: %s, namespace: %s\n", operation, name, namespace)

	if expect, actual := "UPDATE", string(operation); expect != actual {
		err := fmt.Errorf("unexpected operation: expect %s, actual %s", expect, actual)
		log.Println(err)
		return toAdmissionResponse(true, nil)
	}

	// Get namespace with Annotations.
	ns, err := clientset.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	if err != nil {
		log.Println(err)
		return toAdmissionResponse(true, nil)
	}

	if ns.ObjectMeta.Annotations[MonbanKey] == "enabled" {
		var msg string
		msg = "%s is locked in %s\n"
		msg = msg + "If you want to unlock, Please run command `kubectl annotate namespace/%s %s-`"
		err := fmt.Errorf(msg, name, namespace, namespace, MonbanKey)
		log.Println(err)
		return toAdmissionResponse(false, err)
	}

	return toAdmissionResponse(true, nil)
}

type admitFunc func(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse

func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Printf("contentType=%s, expect application/json\n", contentType)
		return
	}

	var reviewResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		log.Fatal(err)
		reviewResponse = toAdmissionResponse(false, err)
	} else {
		reviewResponse = admit(ar)
	}

	response := v1beta1.AdmissionReview{}
	if reviewResponse != nil {
		response.Response = reviewResponse
		response.Response.UID = ar.Request.UID
	}
	// reset the Object and OldObject, they are not needed in a response.
	ar.Request.Object = runtime.RawExtension{}
	ar.Request.OldObject = runtime.RawExtension{}

	resp, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
	}
	if _, err := w.Write(resp); err != nil {
		log.Println(err)
	}
}

func serveDeployments(w http.ResponseWriter, r *http.Request) {
	serve(w, r, admitDeployments)
}

func main() {
	var config Config
	flag.StringVar(&config.CertFile, "tls-cert-file", config.CertFile, "Destination of cert file")
	flag.StringVar(&config.KeyFile, "tls-key-file", config.KeyFile, "Destination of key file")
	flag.StringVar(&config.CAFile, "tls-ca-file", config.CAFile, "Destination of CA file")
	flag.Parse()

	clientset = getClient()

	http.HandleFunc("/deployments", serveDeployments)
	log.Println("Starting monban...")

	server := &http.Server{
		Addr:      ":443",
		TLSConfig: configTLS(config, clientset),
	}
	server.ListenAndServeTLS("", "")
}
