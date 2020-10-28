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
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/v0/",
		Index,
	},

	Route{
		"AddChannel",
		strings.ToUpper("Post"),
		"/v0/channel",
		AddChannel,
	},

	Route{
		"AddConnector",
		strings.ToUpper("Post"),
		"/v0/connector",
		AddConnector,
	},

	Route{
		"AddEventSink",
		strings.ToUpper("Post"),
		"/v0/eventsink",
		AddEventSink,
	},

	Route{
		"AddEventSource",
		strings.ToUpper("Post"),
		"/v0/eventsource",
		AddEventSource,
	},

	Route{
		"GetChannelByName",
		strings.ToUpper("Get"),
		"/v0/channel/{channelName}",
		GetChannelByName,
	},

	Route{
		"GetChannels",
		strings.ToUpper("Get"),
		"/v0/channel",
		GetChannels,
	},

	Route{
		"GetConnectorByName",
		strings.ToUpper("Get"),
		"/v0/connector/{connectorName}",
		GetConnectorByName,
	},

	Route{
		"GetConnectors",
		strings.ToUpper("Get"),
		"/v0/connector",
		GetConnectors,
	},

	Route{
		"GetEventSinkByName",
		strings.ToUpper("Get"),
		"/v0/eventsink/{eventSinkName}",
		GetEventSinkByName,
	},

	Route{
		"GetEventSinks",
		strings.ToUpper("Get"),
		"/v0/eventsink",
		GetEventSinks,
	},

	Route{
		"GetEventSourceByName",
		strings.ToUpper("Get"),
		"/v0/eventsource/{eventSourceName}",
		GetEventSourceByName,
	},

	Route{
		"GetEventSources",
		strings.ToUpper("Get"),
		"/v0/eventsource",
		GetEventSources,
	},

	Route{
		"UpdateChannel",
		strings.ToUpper("Put"),
		"/v0/channel",
		UpdateChannel,
	},

	Route{
		"UpdateConnector",
		strings.ToUpper("Put"),
		"/v0/connector",
		UpdateConnector,
	},

	Route{
		"UpdateEventSink",
		strings.ToUpper("Put"),
		"/v0/eventsink",
		UpdateEventSink,
	},

	Route{
		"UpdateEventSource",
		strings.ToUpper("Put"),
		"/v0/eventsource",
		UpdateEventSource,
	},
}