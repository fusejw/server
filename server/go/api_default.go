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
	"encoding/json"
	"fmt"
	"net/http"
)

func AddChannel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func AddConnector(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func AddEventSink(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func AddEventSource(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func GetChannelByName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	channel := Channel{"my-channel", "knative", "some yaml configuration here!"}
	w.WriteHeader(http.StatusOK)
	printResponse(channel, w)
}

func GetChannels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	channels := make([]Channel, 2)
	channels[0] = Channel{"my-channel-1", "knative", "some yaml configuration here!"}
	channels[1] = Channel{"my-channel-2", "kafka", "some yaml configuration here!"}
	w.WriteHeader(http.StatusOK)
	printResponse(channels, w)
}

func GetConnectorByName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	connector := Connector{"my-connector", "sink", "some yaml configuration here!"}
	w.WriteHeader(http.StatusOK)
	printResponse(connector, w)
}

func printResponse(obj interface{}, w http.ResponseWriter) {
	data, _ := json.Marshal(obj)
	fmt.Fprintf(w, string(data))
}

func GetConnectors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	connectors := make([]Connector, 2)
	connectors[0] = Connector{"my-connector-1", "source", "some yaml configuration here!"}
	connectors[1] = Connector{"my-connector-2", "sink", "some yaml configuration here!"}
	w.WriteHeader(http.StatusOK)
	printResponse(connectors, w)
}

func GetEventSinkByName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	eventSink := EventSourceOrSink{"my-event-sink", "my-connector-sink", "my-channel", nil}
	w.WriteHeader(http.StatusOK)
	printResponse(eventSink, w)
}

func GetEventSinks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	eventSinks := make([]EventSourceOrSink, 2)
	eventSinks[0] = EventSourceOrSink{"my-event-sink-1", "my-connector-sink", "my-channel", nil}
	eventSinks[1] = EventSourceOrSink{"my-event-sink-2", "my-connector-sink", "my-channel", nil}
	w.WriteHeader(http.StatusOK)
	printResponse(eventSinks, w)
}

func GetEventSourceByName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	eventSource := EventSourceOrSink{"my-event-source", "my-connector-source", "my-channel", nil}
	w.WriteHeader(http.StatusOK)
	printResponse(eventSource, w)
}

func GetEventSources(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	eventSources := make([]EventSourceOrSink, 2)
	eventSources[0] = EventSourceOrSink{"my-event-source", "my-connector-source", "my-channel", nil}
	eventSources[1] = EventSourceOrSink{"my-event-source", "my-connector-source", "my-channel", nil}
	w.WriteHeader(http.StatusOK)
	printResponse(eventSources, w)
}

func UpdateChannel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func UpdateConnector(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func UpdateEventSink(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func UpdateEventSource(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}