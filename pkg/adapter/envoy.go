package adapter

import (
	"io"
	"net/http"
	"service-mesh-api-discovery/pkg/k8s"

	accesslogv3 "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v3"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

// envoyAdapter implements accesslogv3.AccessLogServiceServer
type envoyAdapter struct {
	k8s            k8s.K8sCollector
	discoveredAPIs map[string][]string // structore to store discovered api paths
}

func NewAdapter(collector k8s.K8sCollector) accesslogv3.AccessLogServiceServer {
	return &envoyAdapter{
		k8s:            collector,
		discoveredAPIs: make(map[string][]string),
	}
}

// StreamAccessLogs receives and process envoy logs
func (a *envoyAdapter) StreamAccessLogs(logs accesslogv3.AccessLogService_StreamAccessLogsServer) error {
	for {
		entry, err := logs.Recv()
		slog.Info("receiving data...")
		if err == io.EOF {
			slog.Warn("received EOF, exiting")
			return nil
		}
		if err != nil {
			return err
		}

		a.process(entry.GetHttpLogs())
	}
}

// process updates discovered API endpoints based on incomming envoy log entries
func (a *envoyAdapter) process(httpLogs *accesslogv3.StreamAccessLogsMessage_HTTPAccessLogEntries) {
	for _, e := range httpLogs.GetLogEntry() {
		srcAddr := e.CommonProperties.GetDownstreamRemoteAddress().GetSocketAddress().GetAddress()
		dstAddr := e.CommonProperties.GetUpstreamRemoteAddress().GetSocketAddress().GetAddress()

		dest := a.k8s.LookupFor(dstAddr)
		src := a.k8s.LookupFor(srcAddr)

		slog.Debug("incomming HTTP request", "name", src)

		respCode := e.Response.GetResponseCode().GetValue()

		// perhaps this could be tighten up to more realistic conditions
		if dest != "" && respCode != http.StatusNotFound && respCode != http.StatusGatewayTimeout {
			api, ok := a.discoveredAPIs[dest]
			apiPath := e.Request.GetPath()

			if slices.Contains(api, apiPath) {
				continue
			}

			slog.Info("found new API endpoint", "name", dest, "endpoint", apiPath)
			if !ok {
				a.discoveredAPIs[dest] = []string{apiPath}
			}
			a.discoveredAPIs[dest] = append(api, apiPath)
		}
	}
}
