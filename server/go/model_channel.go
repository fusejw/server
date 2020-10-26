/*
 * Jabberwocky
 *
 * Draft version
 *
 * API version: 0.0.1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type Channel struct {

	Name string `json:"name"`
	// the channel type, a knative destination, a kafka topic or a generic URI
	Type_ string `json:"type"`

	Configuration string `json:"configuration"`
}
