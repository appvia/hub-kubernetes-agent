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

// The resource specification of a global service account
type GlobalServiceAccountSpec struct {

	// The name of the global service account
	Name string `json:"global_service_account_name"`

	// The token associated with the global service account
	Token string `json:"token,omitempty"`
}
