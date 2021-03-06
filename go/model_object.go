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

// The definition of a object
type Object struct {

	// A globally unique human readible resource name
	Name string `json:"name"`

	// A cryptographic signature used to verify the request payload
	Signature string `json:"signature,omitempty"`
}
