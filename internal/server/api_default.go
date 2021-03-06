/*
 * Jabberwocky
 *
 * Draft version
 *
 * API version: 0.0.1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	v1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	"github.com/gorilla/mux"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/apache/camel-k/pkg/client/camel/clientset/versioned"
	kamel "github.com/apache/camel-k/pkg/client/camel/clientset/versioned"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	strimzi "github.com/fusejw/server/internal/pkg/strimzi"
)

var kamelClient, kubeClient, clientSet, ctx = localKubeConfiguration()

func localKubeConfiguration() (*versioned.Clientset, client.Client, *kubernetes.Clientset, context.Context) {

	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		log.Println("Trying to connect to in cluster configuration")
		kubeconfig = ""
	} else {
		log.Println("Using local kubeconfig file: ", kubeconfig)
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	checkFatalError(err)

	kamelClient, err := kamel.NewForConfig(cfg)
	checkFatalError(err)

	kubeClient, err := client.New(cfg, client.Options{})
	checkFatalError(err)

	clientSet, err := kubernetes.NewForConfig(cfg)
	checkFatalError(err)

	ctx := context.Background()

	return kamelClient, kubeClient, clientSet, ctx
}

func checkFatalError(err error) {
	if err != nil {
		log.Fatal(err)
		// Non recoverable error
		os.Exit(1)
	}
}

func printResponse(obj interface{}, status int, w http.ResponseWriter) {
	data, _ := json.Marshal(obj)
	printResponseRaw(string(data), "application/json; charset=UTF-8", status, w)
}

func printResponseRaw(data string, contentType string, status int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	fmt.Fprintf(w, data)
}

func printResponseError(err error, w http.ResponseWriter) {
	w.WriteHeader(500)
	fmt.Fprintf(w, err.Error())
}

func OpenAPI(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("api/swagger.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		printResponseError(err, w)
		return
	}
	printResponseRaw(string(data), "text/plain; charset=UTF-8", 200, w)
}

func Health(w http.ResponseWriter, r *http.Request) {
	status := "UP"
	connectorsCount := len(getConnectors(metav1.ListOptions{}))
	channelsCount := len(getChannels(&client.ListOptions{Namespace: "default"}))
	eventSourcesCount := len(getEventSources(metav1.ListOptions{}))
	eventSinksCount := len(getEventSinks(metav1.ListOptions{}))
	healthResponse := fmt.Sprintf("{\"status\": \"%s\", \"connectors\": %d, \"channels\": %d, \"eventSources\": %d, \"eventSinks\": %d}",
		status, connectorsCount, channelsCount, eventSourcesCount, eventSinksCount)

	printResponseRaw(healthResponse, "application/json; charset=UTF-8", 200, w)
}

func AddChannel(w http.ResponseWriter, r *http.Request) {
	printResponseError(errors.New("Not yet implemented"), w)
}

func AddConnector(w http.ResponseWriter, r *http.Request) {
	result, createError := createConnector(r)

	if createError != nil {
		printResponseError(createError, w)
		return
	}
	printResponse(result, 201, w)
}

func createConnector(r *http.Request) (*v1alpha1.Kamelet, error) {
	var connector Connector
	_ = json.NewDecoder(r.Body).Decode(&connector)

	kameletSpec, _ := json.Marshal(connector.Configuration)
	fmt.Println("Spec: ", string(kameletSpec))

	kamelet := v1alpha1.Kamelet{}
	kamelet.Name = connector.Name
	kamelet.Spec.Definition.Default = &v1alpha1.JSON{kameletSpec}

	return kamelClient.CamelV1alpha1().Kamelets("default").Create(ctx, &kamelet, metav1.CreateOptions{
		metav1.TypeMeta{
			Kind:       "Kamelet",
			APIVersion: "camel.apache.org/v1alpha1"}, nil, ""})
}

func AddEventSink(w http.ResponseWriter, r *http.Request) {
	result, createError := createEventSourceOrSink(r, "destination")

	if createError != nil {
		printResponseError(createError, w)
		return
	}

	printResponse(result, 201, w)
}

func AddEventSource(w http.ResponseWriter, r *http.Request) {
	result, createError := createEventSourceOrSink(r, "source")

	if createError != nil {
		printResponseError(createError, w)
		return
	}
	printResponse(result, 201, w)
}

func createEventSourceOrSink(r *http.Request, kameletType string) (*v1alpha1.KameletBinding, error) {
	var eventSourceOrSink EventSourceOrSink
	_ = json.NewDecoder(r.Body).Decode(&eventSourceOrSink)

	convertedProperties := []byte(convertProperties(eventSourceOrSink.Properties))
	eventOrigin := v1alpha1.Endpoint{
		Ref: &corev1.ObjectReference{
			Kind:       "Kamelet",
			APIVersion: "camel.apache.org/v1alpha1",
			Name:       eventSourceOrSink.ConnectorRef},
		// default, empty properties
		Properties: v1alpha1.EndpointProperties{[]byte("{}")}}
	eventDestination := v1alpha1.Endpoint{
		Ref: &corev1.ObjectReference{
			Kind:       "KafkaTopic",
			APIVersion: "kafka.strimzi.io/v1beta1",
			Name:       eventSourceOrSink.ChannelRef},
		Properties: v1alpha1.EndpointProperties{[]byte("{}")}}

	kameletBinding := v1alpha1.NewKameletBinding("default", eventSourceOrSink.Name)

	// It's either a source or a sink, based on the origin of events
	if kameletType == "source" {
		eventOrigin.Properties = v1alpha1.EndpointProperties{convertedProperties}
		kameletBinding.Spec = v1alpha1.KameletBindingSpec{Source: eventOrigin, Sink: eventDestination}
	} else if kameletType == "destination" {
		eventDestination.Properties = v1alpha1.EndpointProperties{convertedProperties}
		kameletBinding.Spec = v1alpha1.KameletBindingSpec{Source: eventDestination, Sink: eventOrigin}
	} else {
		return nil, errors.New("You need to specify either a source or sink type, provided: " + kameletType)
	}

	return kamelClient.CamelV1alpha1().KameletBindings("default").Create(ctx, &kameletBinding, metav1.CreateOptions{
		metav1.TypeMeta{
			Kind:       "KameletBinding",
			APIVersion: "camel.apache.org/v1alpha1"}, nil, ""})
}

func convertProperties(properties []Property) (toRawProperties string) {
	toRawProperties = "{"
	for i, property := range properties {
		toRawProperties += "\"" + property.Name + "\": \"" + property.Value + "\""
		if i < len(properties)-1 {
			toRawProperties += ", "
		}
	}
	toRawProperties += "}"
	return
}

func GetChannelByName(w http.ResponseWriter, r *http.Request) {
	channelName := mux.Vars(r)["channelName"]
	channels := getChannels(&client.ListOptions{Namespace: "default",
		Raw: &metav1.ListOptions{FieldSelector: "metadata.name==" + channelName}})
	if len(channels) == 0 {
		// 404
		printResponse("Not found: "+channelName, http.StatusNotFound, w)
	} else if len(channels) > 1 {
		// 500
		printResponseError(errors.New("Found more than 1 channel with name "+channelName), w)
	} else {
		// 200
		printResponse(channels[0], http.StatusOK, w)
	}
}

func GetChannels(w http.ResponseWriter, r *http.Request) {
	channels := getChannels(&client.ListOptions{Namespace: "default"})
	printResponse(channels, http.StatusOK, w)
}

func getChannels(listOptions *client.ListOptions) (channels []Channel) {
	kafkaTopics := &unstructured.UnstructuredList{}
	kafkaTopics.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kafka.strimzi.io",
		Kind:    "KafkaTopic",
		Version: "v1beta1",
	})
	err := kubeClient.List(context.Background(), kafkaTopics, listOptions)
	if err != nil {
		fmt.Println("[ERROR] ", err.Error())
		return
	}

	for _, kafkaTopic := range kafkaTopics.Items {
		parsedTopic, _ := strimzi.FromUnstructuredObject(kafkaTopic.Object)
		data, _ := json.Marshal(parsedTopic)
		channelName := parsedTopic.Metadata.Name
		if !strings.HasPrefix(channelName, "consumer-offset") {
			eventSources := filterByChannelRef(channelName, getEventSources(metav1.ListOptions{}))
			eventSinks := filterByChannelRef(channelName, getEventSinks(metav1.ListOptions{}))
			channels = append(channels, Channel{channelName, "kafka", eventSources, eventSinks, string(data)})
		}
	}

	return
}

func GetConnectorByName(w http.ResponseWriter, r *http.Request) {
	connectorName := mux.Vars(r)["connectorName"]
	connectors := getConnectors(metav1.ListOptions{FieldSelector: "metadata.name==" + connectorName})
	if len(connectors) == 0 {
		// 404
		printResponse("Not found: "+connectorName, http.StatusNotFound, w)
	} else if len(connectors) > 1 {
		// 500
		printResponseError(errors.New("Found more than 1 connector with name "+connectorName), w)
	} else {
		// 200
		printResponse(connectors[0], http.StatusOK, w)
	}
}

func GetConnectors(w http.ResponseWriter, r *http.Request) {
	connectors := getConnectors(metav1.ListOptions{})
	printResponse(connectors, http.StatusOK, w)
}

func getConnectors(listOptions metav1.ListOptions) (connectors []Connector) {
	kamelets, _ := kamelClient.CamelV1alpha1().Kamelets("default").List(ctx, listOptions)

	connectors = []Connector{}
	for _, kamelet := range kamelets.Items {
		conf, _ := yaml.Marshal(kamelet)
		connectorName := kamelet.Name
		connectorType := kamelet.Labels["camel.apache.org/kamelet.type"]
		properties := []string{}
		eventSourceInstances := []EventSourceOrSink{}
		eventSinkInstances := []EventSourceOrSink{}
		if connectorType == "source" {
			eventSourceInstances = filterByConnectorRef(connectorName, getEventSources(metav1.ListOptions{}))
		} else if connectorType == "sink" {
			eventSinkInstances = filterByConnectorRef(connectorName, getEventSinks(metav1.ListOptions{}))
		}
		// Just get the name of the property
		for key, _ := range kamelet.Spec.Definition.Properties {
			properties = append(properties, key)
		}
		connectors = append(connectors, Connector{kamelet.Name, connectorType, properties,
			eventSourceInstances, eventSinkInstances, string(conf)})
	}

	return
}

func GetEventSinkByName(w http.ResponseWriter, r *http.Request) {
	eventSinkName := mux.Vars(r)["eventSinkName"]
	eventSinks := getEventSinks(metav1.ListOptions{FieldSelector: "metadata.name==" + eventSinkName})
	if len(eventSinks) == 0 {
		// 404
		printResponse("Not found: "+eventSinkName, http.StatusNotFound, w)
	} else if len(eventSinks) > 1 {
		// 500
		printResponseError(errors.New("Found more than 1 event sinks with name "+eventSinkName), w)
	} else {
		// 200
		printResponse(eventSinks[0], http.StatusOK, w)
	}
}

func GetEventSinkLogByName(w http.ResponseWriter, r *http.Request) {
	eventSinkName := mux.Vars(r)["eventSinkName"]
	logOutputByIntegrationName(w, eventSinkName)
}

func logOutputByIntegrationName(w http.ResponseWriter, integration string) {
	pods, err := clientSet.CoreV1().Pods("default").List(ctx, metav1.ListOptions{
		LabelSelector: "camel.apache.org/integration=" + integration})
	if err != nil {
		log.Fatal(err)
	}
	firstPod := pods.Items[0]
	req := clientSet.CoreV1().Pods("default").GetLogs(firstPod.GetName(), &corev1.PodLogOptions{})

	readCloser, err := req.Stream(ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer readCloser.Close()
	_, err = io.Copy(w, readCloser)
	if err != nil {
		printResponseError(err, w)
	}
}

func GetEventSinks(w http.ResponseWriter, r *http.Request) {
	eventSinks := getEventSinks(metav1.ListOptions{})
	printResponse(eventSinks, http.StatusOK, w)
}

func getEventSinks(listOptions metav1.ListOptions) (eventSinks []EventSourceOrSink) {
	eventSinks = []EventSourceOrSink{}
	kameletBindings, _ := kamelClient.CamelV1alpha1().KameletBindings("default").List(ctx, listOptions)
	for _, kameletBinding := range kameletBindings.Items {
		if kameletBinding.Spec.Sink.Ref.Kind == "Kamelet" {
			// From sink perspective, Channel is the source, Source is the destination
			properties := fromRawProperties(kameletBinding.Spec.Sink.Properties.RawMessage)
			eventSinks = append(eventSinks, EventSourceOrSink{
				kameletBinding.Name,
				kameletBinding.Spec.Sink.Ref.Name,
				kameletBinding.Spec.Source.Ref.Name,
				properties})
		}
	}

	return
}

func fromRawProperties(data json.RawMessage) (properties []Property) {
	var objmap map[string]json.RawMessage
	_ = json.Unmarshal(data, &objmap)
	for key, element := range objmap {
		var value string
		_ = json.Unmarshal(element, &value)
		properties = append(properties, Property{key, value})
	}
	return
}

func GetEventSourceByName(w http.ResponseWriter, r *http.Request) {
	eventSourceName := mux.Vars(r)["eventSourceName"]
	eventSources := getEventSources(metav1.ListOptions{FieldSelector: "metadata.name==" + eventSourceName})
	if len(eventSources) == 0 {
		// 404
		printResponse("Not found: "+eventSourceName, http.StatusNotFound, w)
	} else if len(eventSources) > 1 {
		// 500
		printResponseError(errors.New("Found more than 1 event sources with name "+eventSourceName), w)
	} else {
		// 200
		printResponse(eventSources[0], http.StatusOK, w)
	}
}

func GetEventSourceLogByName(w http.ResponseWriter, r *http.Request) {
	eventSourceName := mux.Vars(r)["eventSourceName"]
	logOutputByIntegrationName(w, eventSourceName)
}

func GetEventSources(w http.ResponseWriter, r *http.Request) {
	eventSources := getEventSources(metav1.ListOptions{})
	printResponse(eventSources, http.StatusOK, w)
}

func getEventSources(listOptions metav1.ListOptions) (eventSources []EventSourceOrSink) {
	eventSources = []EventSourceOrSink{}
	kameletBindings, _ := kamelClient.CamelV1alpha1().KameletBindings("default").List(ctx, listOptions)
	for _, kameletBinding := range kameletBindings.Items {
		if kameletBinding.Spec.Source.Ref.Kind == "Kamelet" {
			properties := fromRawProperties(kameletBinding.Spec.Source.Properties.RawMessage)
			eventSources = append(eventSources, EventSourceOrSink{
				kameletBinding.Name,
				kameletBinding.Spec.Source.Ref.Name,
				kameletBinding.Spec.Sink.Ref.Name,
				properties})
		}
	}

	return
}

func filterByChannelRef(channelRef string, eventsIn []EventSourceOrSink) (eventsOut []EventSourceOrSink) {
	eventsOut = []EventSourceOrSink{}
	for _, event := range eventsIn {
		if event.ChannelRef == channelRef {
			eventsOut = append(eventsOut, event)
		}
	}

	return
}

func filterByConnectorRef(connectorRef string, eventsIn []EventSourceOrSink) (eventsOut []EventSourceOrSink) {
	eventsOut = []EventSourceOrSink{}
	for _, event := range eventsIn {
		if event.ConnectorRef == connectorRef {
			eventsOut = append(eventsOut, event)
		}
	}

	return
}

func UpdateChannel(w http.ResponseWriter, r *http.Request) {
	printResponseError(errors.New("Not yet implemented"), w)
}

func UpdateConnector(w http.ResponseWriter, r *http.Request) {
	printResponseError(errors.New("Not yet implemented"), w)
}

func UpdateEventSink(w http.ResponseWriter, r *http.Request) {
	printResponseError(errors.New("Not yet implemented"), w)
}

func UpdateEventSource(w http.ResponseWriter, r *http.Request) {
	printResponseError(errors.New("Not yet implemented"), w)
}

func DeleteConnectorByName(w http.ResponseWriter, r *http.Request) {
	connectorName := mux.Vars(r)["connectorName"]
	connectors := getConnectors(metav1.ListOptions{FieldSelector: "metadata.name==" + connectorName})
	if len(connectors) == 0 {
		printResponse("Not found: "+connectorName, http.StatusNotFound, w)
	} else if len(connectors) > 1 {
		printResponseError(errors.New("Found more than 1 connector with name "+connectorName), w)
	} else {
		connector := connectors[0]
		eventSinks := len(connector.EventSinks)
		eventSources := len(connector.EventSources)
		if eventSinks > 0 || eventSources > 0 {
			printResponseError(errors.New(fmt.Sprintf("The connector has %d event sources and %d event sinks bound!", eventSources, eventSinks)), w)
		} else {
			err := kamelClient.CamelV1alpha1().Kamelets("default").Delete(ctx, connectorName, v1.DeleteOptions{})
			if err != nil {
				printResponseError(err, w)
			} else {
				printResponse("Deleted", http.StatusNoContent, w)
			}
		}
	}
}

func DeleteEventSinkByName(w http.ResponseWriter, r *http.Request) {
	eventSinkName := mux.Vars(r)["eventSinkName"]
	eventSinks := getEventSinks(metav1.ListOptions{FieldSelector: "metadata.name==" + eventSinkName})
	if len(eventSinks) == 0 {
		printResponse("Not found: "+eventSinkName, http.StatusNotFound, w)
	} else if len(eventSinks) > 1 {
		printResponseError(errors.New("Found more than 1 event sink with name "+eventSinkName), w)
	} else {
		err := kamelClient.CamelV1alpha1().KameletBindings("default").Delete(ctx, eventSinkName, v1.DeleteOptions{})
		if err != nil {
			printResponseError(err, w)
		} else {
			printResponse("Deleted", http.StatusNoContent, w)
		}
	}
}

func DeleteEventSourceByName(w http.ResponseWriter, r *http.Request) {
	eventSourceName := mux.Vars(r)["eventSourceName"]
	eventSources := getEventSources(metav1.ListOptions{FieldSelector: "metadata.name==" + eventSourceName})
	if len(eventSources) == 0 {
		printResponse("Not found: "+eventSourceName, http.StatusNotFound, w)
	} else if len(eventSources) > 1 {
		printResponseError(errors.New("Found more than 1 event source with name "+eventSourceName), w)
	} else {
		err := kamelClient.CamelV1alpha1().KameletBindings("default").Delete(ctx, eventSourceName, v1.DeleteOptions{})
		if err != nil {
			printResponseError(err, w)
		} else {
			printResponse("Deleted", http.StatusNoContent, w)
		}
	}
}
