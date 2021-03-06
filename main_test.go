/*
 * hub-kubernetes-agent
 *
 * an agent used to provision and configure Kubernetes resources
 *
 * API version: v1beta
 * Contact: support@appvia.io
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sw "github.com/appvia/hub-kubernetes-agent/go"
	"github.com/gorilla/mux"
	muxlogrus "github.com/pytimer/mux-logrus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Router() *mux.Router {
	router := sw.NewRouter()
	router.Use(Middleware)

	var logoptions muxlogrus.LogOptions
	logoptions = muxlogrus.LogOptions{Formatter: new(logrus.JSONFormatter), EnableStarting: true}
	router.Use(muxlogrus.NewLogger(logoptions).Middleware)

	return router
}

func TestHealthEndpoint(t *testing.T) {
	request, _ := http.NewRequest("GET", "/api/v1beta/healthz", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
}

func TestUnauthorized(t *testing.T) {
	request, _ := http.NewRequest("GET", "/api/v1beta/namespaces", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 401, response.Code, "Unauthorized response is expected")
}
