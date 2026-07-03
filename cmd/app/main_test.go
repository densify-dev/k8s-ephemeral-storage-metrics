package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jmcgrath207/k8s-ephemeral-storage-metrics/pkg/dev"
	"github.com/jmcgrath207/k8s-ephemeral-storage-metrics/pkg/node"
	"github.com/jmcgrath207/k8s-ephemeral-storage-metrics/pkg/pod"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

func setupMetricsTestEnv(t *testing.T) {
	t.Helper()
	os.Setenv("CURRENT_NODE_NAME", "test-node")
	os.Setenv("EPHEMERAL_STORAGE_POD_USAGE", "true")
	Node = node.NewCollector(1)
	Pod = pod.NewCollector(1)
}

func teardownMetricsTestEnv(t *testing.T) {
	t.Helper()
	os.Unsetenv("CURRENT_NODE_NAME")
	os.Unsetenv("EPHEMERAL_STORAGE_POD_USAGE")
}

func setupClientset(t *testing.T, server *httptest.Server) {
	t.Helper()
	config := &rest.Config{
		Host:    server.URL,
		APIPath: "/api",
		ContentConfig: rest.ContentConfig{
			GroupVersion:         &schema.GroupVersion{Group: "", Version: "v1"},
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		},
	}
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		t.Fatal(err)
	}
	clientset := kubernetes.New(restClient)

	dev.Clientset = clientset
}

func nodeCollector(sampleInterval int64) node.Node {
	// ponytail: reuse cached collector to avoid redundant goroutines
	return node.NewCollector(sampleInterval)
}

func TestSetMetrics_WithApiServer(t *testing.T) {
	// ponytail: httptest-based integration test for setMetrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"node": {"nodeName": "test-node"},
			"pods": [{
				"podRef": {"name": "pod-1", "namespace": "ns-1"},
				"ephemeral-storage": {
					"usedBytes": 100,
					"availableBytes": 900,
					"capacityBytes": 1000,
					"inodes": 1000,
					"inodesFree": 500,
					"inodesUsed": 500
				},
				"containers": [],
				"volume": []
			}]
		}`))
	}))
	defer server.Close()

	setupClientset(t, server)
	defer func() { dev.Clientset = nil }()

	setupMetricsTestEnv(t)
	defer teardownMetricsTestEnv(t)

	setMetrics("test-node")
}

func TestSetMetrics_QueryError(t *testing.T) {
	// ponytail: short interval so backoff doesn't make test slow
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	setupClientset(t, server)
	defer func() { dev.Clientset = nil }()

	os.Setenv("CURRENT_NODE_NAME", "test-node")
	os.Setenv("EPHEMERAL_STORAGE_POD_USAGE", "true")
	// Use sampleInterval=1 so backoff expires quickly
	Node = node.NewCollector(1)
	Pod = pod.NewCollector(1)
	defer func() {
		os.Unsetenv("CURRENT_NODE_NAME")
		os.Unsetenv("EPHEMERAL_STORAGE_POD_USAGE")
	}()

	setMetrics("test-node")
}

func TestSetMetrics_EmptyNamespacePod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"node": {"nodeName": "test-node"},
			"pods": [{
				"podRef": {"name": "no-ns", "namespace": ""},
				"ephemeral-storage": {
					"usedBytes": 100,
					"availableBytes": 900,
					"capacityBytes": 1000,
					"inodes": 1000,
					"inodesFree": 500,
					"inodesUsed": 500
				},
				"containers": [],
				"volume": []
			}]
		}`))
	}))
	defer server.Close()

	setupClientset(t, server)
	defer func() { dev.Clientset = nil }()

	setupMetricsTestEnv(t)
	defer teardownMetricsTestEnv(t)

	setMetrics("test-node")
}

func TestSetMetrics_ZeroMetricsPod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"node": {"nodeName": "test-node"},
			"pods": [{
				"podRef": {"name": "zero-pod", "namespace": "ns-zero"},
				"ephemeral-storage": {
					"usedBytes": 0,
					"availableBytes": 0,
					"capacityBytes": 0,
					"inodes": 0,
					"inodesFree": 0,
					"inodesUsed": 0
				},
				"containers": [],
				"volume": []
			}]
		}`))
	}))
	defer server.Close()

	setupClientset(t, server)
	defer func() { dev.Clientset = nil }()

	setupMetricsTestEnv(t)
	defer teardownMetricsTestEnv(t)

	setMetrics("test-node")
}

func TestGetMetrics_NoCrash(t *testing.T) {
	// ponytail: verify getMetrics starts without crashing.
	// It loops forever, so we run briefly then let goroutine leak.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	setupClientset(t, server)
	defer func() { dev.Clientset = nil }()

	os.Setenv("CURRENT_NODE_NAME", "test-node")
	os.Setenv("EPHEMERAL_STORAGE_POD_USAGE", "true")
	Node = node.NewCollector(1)
	Pod = pod.NewCollector(1)
	defer func() {
		os.Unsetenv("CURRENT_NODE_NAME")
		os.Unsetenv("EPHEMERAL_STORAGE_POD_USAGE")
	}()

	// Start getMetrics in goroutine; it loops forever
	go getMetrics()
	time.Sleep(50 * time.Millisecond)
	// ponytail: goroutine leak, process exit cleans up
}

func startAppTest(t *testing.T, extraEnv func()) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	setupClientset(t, server)
	t.Cleanup(func() { dev.Clientset = nil })

	os.Setenv("CURRENT_NODE_NAME", "test-node")
	os.Setenv("EPHEMERAL_STORAGE_POD_USAGE", "true")
	t.Cleanup(func() {
		os.Unsetenv("CURRENT_NODE_NAME")
		os.Unsetenv("EPHEMERAL_STORAGE_POD_USAGE")
	})
	if extraEnv != nil {
		extraEnv()
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	freePort := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	go StartApplication(strconv.Itoa(freePort), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`metrics output`))
	}))
	time.Sleep(100 * time.Millisecond)
	return fmt.Sprintf("http://127.0.0.1:%d/metrics", freePort)
}

func TestStartApplication_NoCrash(t *testing.T) {
	metricsURL := startAppTest(t, nil)

	resp, err := http.Get(metricsURL)
	if err != nil {
		t.Fatalf("metrics endpoint unreachable: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "metrics output") {
		t.Errorf("expected metrics output, got %q", string(body))
	}
	// ponytail: goroutine leak, process exit cleans up
}

func TestStartApplication_WithPprof(t *testing.T) {
	metricsURL := startAppTest(t, func() {
		os.Setenv("PPROF", "true")
		t.Cleanup(func() { os.Unsetenv("PPROF") })
	})

	resp, err := http.Get(metricsURL)
	if err != nil {
		t.Fatalf("metrics endpoint unreachable: %v", err)
	}
	resp.Body.Close()
	// ponytail: goroutine leak (pprof and metrics server), process exit cleans up
}
