/*
 * hub-kubernetes-agent
 *
 * an agent used to provision and configure Kubernetes resources
 *
 * API version: v1beta
 * Contact: support@appvia.io
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package swagger

// The resource definition for a list of Docker image uris
type ImageUriList struct {

	// A list of Docker image uris
	Uri []string `json:",omitempty"`
}
