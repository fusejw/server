/*
 * Jabberwocky
 *
 * Draft version
 *
 * API version: 0.0.1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	strimzi "./strimzi"
)

var kamelClient, kubeClient, clientSet, ctx = localKubeConfiguration()

func localKubeConfiguration() (*versioned.Clientset, client.Client, *kubernetes.Clientset, context.Context) {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	log.Println("Using kubeconfig file: ", kubeconfig)
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

func Health(w http.ResponseWriter, r *http.Request) {
	status := "UP"
	connectorsCount := len(getConnectors())
	channelsCount := len(getChannels())
	eventSourcesCount := len(getEventSources())
	eventSinksCount := len(getEventSinks())
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
	printResponseError(errors.New("Not yet implemented"), w)
}

func GetChannels(w http.ResponseWriter, r *http.Request) {
	channels := getChannels()
	printResponse(channels, http.StatusOK, w)
}

func getChannels() (channels []Channel) {
	kafkaTopics := &unstructured.UnstructuredList{}
	kafkaTopics.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kafka.strimzi.io",
		Kind:    "KafkaTopic",
		Version: "v1beta1",
	})
	_ = kubeClient.List(context.Background(), kafkaTopics)

	for _, kafkaTopic := range kafkaTopics.Items {
		parsedTopic, _ := strimzi.FromUnstructuredObject(kafkaTopic.Object)
		data, _ := json.Marshal(parsedTopic)
		channelName := parsedTopic.Metadata.Name
		if !strings.HasPrefix(channelName, "consumer-offset") {
			eventSources := filterByChannelRef(channelName, getEventSources())
			eventSinks := filterByChannelRef(channelName, getEventSinks())
			channels = append(channels, Channel{channelName, "kafka", eventSources, eventSinks, string(data)})
		}
	}

	return
}

func GetConnectorByName(w http.ResponseWriter, r *http.Request) {
	printResponseError(errors.New("Not yet implemented"), w)
}

func GetConnectors(w http.ResponseWriter, r *http.Request) {
	connectors := getConnectors()
	printResponse(connectors, http.StatusOK, w)
}

func getConnectors() (connectors []Connector) {
	kamelets, _ := kamelClient.CamelV1alpha1().Kamelets("default").List(ctx, metav1.ListOptions{})

	connectors = []Connector{}
	for _, kamelet := range kamelets.Items {
		conf, _ := yaml.Marshal(kamelet)
		connectorName := kamelet.Name
		connectorType := kamelet.Labels["camel.apache.org/kamelet.type"]
		properties := []string{}
		eventSourceInstances := []EventSourceOrSink{}
		eventSinkInstances := []EventSourceOrSink{}
		if connectorType == "source" {
			eventSourceInstances = filterByConnectorRef(connectorName, getEventSources())
		} else if connectorType == "sink" {
			eventSinkInstances = filterByConnectorRef(connectorName, getEventSinks())
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
	printResponseError(errors.New("Not yet implemented"), w)
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
	eventSinks := getEventSinks()
	printResponse(eventSinks, http.StatusOK, w)
}

func getEventSinks() (eventSinks []EventSourceOrSink) {
	eventSinks = []EventSourceOrSink{}
	kameletBindings, _ := kamelClient.CamelV1alpha1().KameletBindings("default").List(ctx, metav1.ListOptions{})
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
	printResponseError(errors.New("Not yet implemented"), w)
}

func GetEventSourceLogByName(w http.ResponseWriter, r *http.Request) {
	eventSourceName := mux.Vars(r)["eventSourceName"]
	logOutputByIntegrationName(w, eventSourceName)
}

func GetEventSources(w http.ResponseWriter, r *http.Request) {
	eventSources := getEventSources()
	printResponse(eventSources, http.StatusOK, w)
}

func getEventSources() (eventSources []EventSourceOrSink) {
	eventSources = []EventSourceOrSink{}
	kameletBindings, _ := kamelClient.CamelV1alpha1().KameletBindings("default").List(ctx, metav1.ListOptions{})
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
