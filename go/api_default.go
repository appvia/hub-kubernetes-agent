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

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	logrus "github.com/sirupsen/logrus"
	apicorev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var clientset *kubernetes.Clientset

func getClient(server, token, caCert string) (*kubernetes.Clientset, error) {
	decodedCert, err := base64.StdEncoding.DecodeString(caCert)

	if err != nil {
		logrus.Println("decode error:", err)
		return nil, err
	}

	config := &rest.Config{
		Host:            server,
		BearerToken:     token,
		TLSClientConfig: rest.TLSClientConfig{CAData: decodedCert},
	}

	client, err := kubernetes.NewForConfig(config)
	_ = client

	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func handleSuccess(w http.ResponseWriter, payload []byte) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(payload)
}

func handleDelete(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func handleInternalServerError(w http.ResponseWriter, reason string, err error) {
	logrus.Println(err.Error())
	var apiError ApiError
	apiError = ApiError{Reason: reason, Detail: err.Error()}
	payload, err := json.Marshal(apiError)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(payload)
}

func handleNotFoundError(w http.ResponseWriter, err error) {
	logrus.Println(err.Error())
	var apiError ApiError
	apiError = ApiError{Reason: "not found", Detail: err.Error()}
	payload, err := json.Marshal(apiError)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write(payload)
}

func NamespacesList(w http.ResponseWriter, r *http.Request) {
	clientset, err := getClient(r.Header.Get("X-Kube-API-URL"), r.Header.Get("X-Kube-Token"), r.Header.Get("X-Kube-CA"))

	if err != nil {
		logrus.Println(err)
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})

	if err != nil {
		handleInternalServerError(w, "error listing namespaces", err)
		return
	} else {
		var namespaceList []Namespace
		for _, namespace := range namespaces.Items {
			namespaceList = append(namespaceList, Namespace{Name: namespace.Name})
		}
		payload, err := json.Marshal(namespaceList)
		if err != nil {
			logrus.Println(err)
		}
		handleSuccess(w, payload)
		return
	}
}

func NamespacesNameGet(w http.ResponseWriter, r *http.Request) {
	clientset, err := getClient(r.Header.Get("X-Kube-API-URL"), r.Header.Get("X-Kube-Token"), r.Header.Get("X-Kube-CA"))

	vars := mux.Vars(r)
	name := vars["name"]
	namespace, err := clientset.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
	_ = namespace

	if errors.IsNotFound(err) {
		handleNotFoundError(w, err)
		return
	}

	if err != nil {
		handleInternalServerError(w, "Error getting namespace", err)
		return
	}

	serviceAccounts, err := clientset.CoreV1().ServiceAccounts(name).List(metav1.ListOptions{})
	_ = serviceAccounts

	if err != nil {
		handleInternalServerError(w, "error getting service accounts for namespace", err)
		return
	}

	namespaceServiceAccounts := make([]map[string]string, len(serviceAccounts.Items))

	for _, sa := range serviceAccounts.Items {
		item := map[string]string{"name": sa.Name}
		namespaceServiceAccounts = append(namespaceServiceAccounts, item)
	}

	var namespaceResponse Namespace
	namespaceResponse = Namespace{Name: name, Spec: &NamespaceSpec{ServiceAccounts: namespaceServiceAccounts}}
	payload, err := json.Marshal(namespaceResponse)
	if err != nil {
		logrus.Println(err)
	}
	handleSuccess(w, payload)
	return
}

func NamespacesNameDelete(w http.ResponseWriter, r *http.Request) {
	clientset, err := getClient(r.Header.Get("X-Kube-API-URL"), r.Header.Get("X-Kube-Token"), r.Header.Get("X-Kube-CA"))

	if err != nil {
		logrus.Println(err)
	}

	vars := mux.Vars(r)
	name := vars["name"]

	if err := clientset.CoreV1().Namespaces().Delete(name, &metav1.DeleteOptions{}); errors.IsNotFound(err) || err == nil {
		logrus.Infof("Deleted namespace: %s", name)
		handleDelete(w)
		return
	} else {
		handleInternalServerError(w, "error deleting namespace", err)
		return
	}
}

func NamespacesNamePut(w http.ResponseWriter, r *http.Request) {
	clientset, err := getClient(r.Header.Get("X-Kube-API-URL"), r.Header.Get("X-Kube-Token"), r.Header.Get("X-Kube-CA"))

	if err != nil {
		panic(err)
	}

	vars := mux.Vars(r)
	name := vars["name"]
	_ = name

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	} else {
		logrus.Println(string(body))
	}

	var n Namespace
	err = json.Unmarshal(body, &n)

	if err != nil {
		logrus.Println(err)
		handleInternalServerError(w, "client error", err)
		return
	}

	namespaceName := n.Name
	namespaceServiceAccounts := n.Spec.ServiceAccounts

	logrus.Infof("Attempting to create namespace: %s", namespaceName)

	namespace, err := clientset.CoreV1().Namespaces().Create(&apicorev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	})
	_ = namespace

	if err == nil {
		logrus.Infof("Namespace successfully created: %s", namespaceName)
	} else if errors.IsAlreadyExists(err) {
		logrus.Infof("Namespace already exists: %s", namespaceName)
	} else if err != nil {
		handleInternalServerError(w, "error creating namespace", err)
		return
	}

	for _, sa := range namespaceServiceAccounts {
		subject := rbacv1.Subject{
			Kind:      "ServiceAccount",
			Name:      sa["name"],
			Namespace: "default",
		}

		var subjects []rbacv1.Subject

		subjects = append(subjects, subject)

		roleRef := rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "admin",
		}

		roleBinding := rbacv1.RoleBinding{
			Subjects: subjects,
			RoleRef:  roleRef,
			ObjectMeta: metav1.ObjectMeta{
				Name: sa["name"] + "-admin-" + namespaceName,
			},
		}

		roleBindingReponse, err := clientset.Rbac().RoleBindings(namespaceName).Create(&roleBinding)
		_ = roleBindingReponse

		if err == nil {
			logrus.Infof("Created role binding: %s-admin-%s", sa["name"], namespaceName)
		} else if errors.IsAlreadyExists(err) {
			logrus.Infof("Role binding already exists: %s-admin-%s", sa["name"], namespaceName)
		} else {
			logrus.Infof("Failed to create role binding: %s-admin-%s", sa["name"], namespaceName)
			handleInternalServerError(w, "error creating rolebinding for namespace", err)
			return
		}
	}
	var namespaceItem Namespace
	namespaceItem = Namespace{Name: namespaceName, Spec: &NamespaceSpec{ServiceAccounts: namespaceServiceAccounts}}
	payload, err := json.Marshal(namespaceItem)
	if err != nil {
		logrus.Println(err)
	}
	handleSuccess(w, payload)
	return
}

func ServiceAccountsNamespaceGet(w http.ResponseWriter, r *http.Request) {
	clientset, err := getClient(r.Header.Get("X-Kube-API-URL"), r.Header.Get("X-Kube-Token"), r.Header.Get("X-Kube-CA"))

	vars := mux.Vars(r)
	namespace := vars["namespace"]

	if namespace == "" {
		namespace = "default"
	}
	namespaceCheck, err := clientset.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	_ = namespaceCheck

	if errors.IsNotFound(err) {
		logrus.Infof("Namespace: %s not found\n", namespace)
		handleNotFoundError(w, err)
		return
	}

	serviceAccounts, err := clientset.CoreV1().ServiceAccounts(namespace).List(metav1.ListOptions{})
	_ = serviceAccounts

	if err != nil {
		handleInternalServerError(w, "error getting service accounts", err)
		return
	}

	var serviceAccountsList []string
	for _, sa := range serviceAccounts.Items {
		serviceAccountsList = append(serviceAccountsList, sa.Name)
	}

	payload, err := json.Marshal(serviceAccountsList)

	if err != nil {
		logrus.Println(err)
	}

	logrus.Infof("Listing service accounts for namespace: %s", namespace)
	handleSuccess(w, payload)
	return
}

func ServiceAccountsNamespaceNameDelete(w http.ResponseWriter, r *http.Request) {
	clientset, err := getClient(r.Header.Get("X-Kube-API-URL"), r.Header.Get("X-Kube-Token"), r.Header.Get("X-Kube-CA"))

	if err != nil {
		logrus.Println(err)
	}

	vars := mux.Vars(r)
	name := vars["name"]
	namespace := vars["namespace"]

	if namespace == "" {
		namespace = "default"
	}

	if err := clientset.CoreV1().ServiceAccounts(namespace).Delete(name, &metav1.DeleteOptions{}); errors.IsNotFound(err) || err == nil {
		logrus.Infof("Deleted service account: %s from namespace: %s", name, namespace)
		handleDelete(w)
		return
	} else if err != nil {
		handleInternalServerError(w, "error deleting service account", err)
		return
	}
}

func ServiceAccountsNamespaceNameGet(w http.ResponseWriter, r *http.Request) {
	clientset, err := getClient(r.Header.Get("X-Kube-API-URL"), r.Header.Get("X-Kube-Token"), r.Header.Get("X-Kube-CA"))

	vars := mux.Vars(r)
	name := vars["name"]
	namespace := vars["namespace"]

	if namespace == "" {
		namespace = "default"
	}

	serviceAccount, err := clientset.CoreV1().ServiceAccounts(namespace).Get(name, metav1.GetOptions{})
	_ = serviceAccount

	if errors.IsNotFound(err) {
		logrus.Infof("Service account %s not found\n", name)
		handleNotFoundError(w, err)
		return
	}

	if err != nil {
		handleInternalServerError(w, "error getting service account", err)
		return
	}

	secret, err := clientset.CoreV1().Secrets(namespace).Get(serviceAccount.Secrets[0].Name, metav1.GetOptions{})
	_ = secret

	if err != nil {
		logrus.Infof("Error getting service account token for %s", name)
		logrus.Println(err)
	}

	// TODO: add token to response
	// ServiceAccountSpec{Name: name, Token: string(secret.Data["token"])}

	var serviceAccountItem ServiceAccount
	serviceAccountItem = ServiceAccount{Name: name}
	payload, err := json.Marshal(serviceAccountItem)
	if err != nil {
		logrus.Println(err)
	}
	logrus.Infof("Found service account: %s", name)
	handleSuccess(w, payload)
	return
}

func ServiceAccountsNamespaceNamePut(w http.ResponseWriter, r *http.Request) {
	clientset, err := getClient(r.Header.Get("X-Kube-API-URL"), r.Header.Get("X-Kube-Token"), r.Header.Get("X-Kube-CA"))

	vars := mux.Vars(r)
	name := vars["name"]
	namespace := vars["namespace"]

	serviceAccount, err := clientset.CoreV1().ServiceAccounts(namespace).Create(&apicorev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
	_ = serviceAccount

	if errors.IsAlreadyExists(err) || err == nil {
		// TODO: add token to response
		// ServiceAccountSpec{Name: name, Token: string(secret.Data["token"])}
		var serviceAccountItem ServiceAccount
		serviceAccountItem = ServiceAccount{Name: name}
		payload, err := json.Marshal(serviceAccountItem)
		if err != nil {
			logrus.Println(err)
		}
		logrus.Infof("Created service account: %s in namespace: %s", name, namespace)
		handleSuccess(w, payload)
		return
	} else {
		logrus.Println(err)
		handleInternalServerError(w, "error creating service account", err)
		return
	}
}
