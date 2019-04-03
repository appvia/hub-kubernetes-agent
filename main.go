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
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/urfave/cli"

	sw "github.com/appvia/hub-kubernetes-agent/go"
	muxlogrus "github.com/pytimer/mux-logrus"
	logrus "github.com/sirupsen/logrus"
)

var (
	release = "v0.0.1"
)

func invokeServerAction(ctx *cli.Context) error {
	router := sw.NewRouter()
	router.Use(Middleware)

	var logoptions muxlogrus.LogOptions
	logoptions = muxlogrus.LogOptions{Formatter: new(logrus.JSONFormatter), EnableStarting: true}
	router.Use(muxlogrus.NewLogger(logoptions).Middleware)

	if ctx.String("https-port") != "" && ctx.String("tls-key") != "" && ctx.String("tls-cert") != "" {
		cfg := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}
		srv := &http.Server{
			Addr:         ctx.String("listen")+":"+ctx.String("https-port"),
			Handler:      router,
			TLSConfig:    cfg,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}
		go func() {
			if err := srv.ListenAndServeTLS(ctx.String("tls-cert"), ctx.String("tls-key")); err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("failed to start the api service")
			}
		}()
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-signalChannel

		return nil
	} else {
		go func() {
			if err := http.ListenAndServe(ctx.String("listen")+":"+ctx.String("http-port"), router); err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("failed to start the api service")
			}
		}()
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-signalChannel

		return nil
	}

}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "healthz") == true {
			next.ServeHTTP(w, r)
			return
		}
		tokenHeader := r.Header.Get("Authorization")
		if len(tokenHeader) == 0 {
			logrus.Infof("Missing Authorization header")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		token := strings.Replace(tokenHeader, "Bearer ", "", 1)
		if token != os.Getenv("AUTH_TOKEN") {
			logrus.Infof("Incorrect Authorization header")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if r.Header.Get("X-Kube-API-URL") == "" || r.Header.Get("X-Kube-Token") == "" || r.Header.Get("X-Kube-CA") == "" {
			logrus.Infof("Missing Kube header")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func main() {
	app := &cli.App{
		Name:    "hub-kubernetes-agent",
		Author:  "Daniel Whatmuff",
		Email:   "daniel.whatmuff@appvia.io",
		Usage:   "A backend agent used to provision resources within Kubernetes clusters",
		Version: release,

		OnUsageError: func(context *cli.Context, err error, _ bool) error {
			fmt.Fprintf(os.Stderr, "[error] invalid options %s\n", err)
			return err
		},

		Action: func(ctx *cli.Context) error {
			if ctx.String("auth-token") == "" {
				return cli.NewExitError("Missing AUTH_TOKEN", 1)
			}
			os.Setenv("AUTH_TOKEN", ctx.String("auth-token"))
			logrus.Info("Starting server...")
			return invokeServerAction(ctx)
		},

		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "listen",
				Usage:  "the interface to bind the service to `INTERFACE`",
				Value:  "127.0.0.1",
				EnvVar: "LISTEN",
			},
			cli.StringFlag{
				Name:   "http-port",
				Usage:  "network interface the service should listen on `PORT`",
				Value:  "10080",
				EnvVar: "HTTP_PORT",
			},
			cli.StringFlag{
				Name:   "https-port",
				Usage:  "network interface the service should listen on `PORT`",
				Value:  "10443",
				EnvVar: "HTTPS_PORT",
			},
			cli.StringFlag{
				Name:   "auth-token",
				Usage:  "authentication token used to verifier the caller `TOKEN`",
				EnvVar: "AUTH_TOKEN",
			},
			cli.StringFlag{
				Name:   "tls-cert",
				Usage:  "the path to the file containing the certificate pem `PATH`",
				EnvVar: "TLS_CERT",
			},
			cli.StringFlag{
				Name:   "tls-key",
				Usage:  "the path to the file containing the private key pem `PATH`",
				EnvVar: "TLS_KEY",
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}
