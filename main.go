package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rluisr/mysqlrouter-go"
)

var nameSpace = "mysqlrouter"

var (
	routerStatusGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: nameSpace,
			Name:      "router_status",
			Help:      "MySQL Router information",
		}, []string{"process_id", "product_edition", "time_started", "version", "hostname"})
	metadataGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: nameSpace,
			Name:      "metadata",
			Help:      "metadata list",
		}, []string{"name"})
	metadataConfigGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: nameSpace,
			Name:      "metadata_config",
			Help:      "metadata config",
		}, []string{"name", "cluster_name", "time_refresh_in_ms", "group_replication_id"})
	metadataConfigNodeGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: nameSpace,
			Name:      "metadata_config_node",
			Help:      "metadata config node",
		}, []string{"name", "router_host", "cluster_name", "hostname", "port"})
	metadataStatusGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: nameSpace,
			Name:      "metadata_status",
			Help:      "metadata status",
		}, []string{"name", "refresh_failed", "time_last_refresh_succeeded", "last_refresh_hostname", "last_refresh_port"})
	routeGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: nameSpace,
			Name:      "route",
			Help:      "route name",
		}, []string{"name"})
	routeActiveConnectionsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: nameSpace,
			Name:      "route_active_connections",
			Help:      "route active connections",
		}, []string{"name", "router_hostname"})
	routeTotalConnectionsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: nameSpace,
			Name:      "route_total_connections",
			Help:      "route total connections",
		}, []string{"name", "router_hostname"})
	routeBlockedHostsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: nameSpace,
			Name:      "route_blocked_hosts",
			Help:      "route blocked_hosts",
		}, []string{"name", "router_hostname"})
	routeHealthGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: nameSpace,
			Name:      "route_health",
			Help:      "0: not active, 1: active",
		}, []string{"name", "router_hostname"})
	routeDestinationsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: nameSpace,
			Name:      "route_destinations",
			Help:      "",
		}, []string{"name", "address", "port"})
	routeConnectionsByteFromServerGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "route_connections_byte_from_server",
			Help: "Route connections byte from server",
		}, []string{"name", "router_hostname", "source_address", "destination_address"})
	routeConnectionsByteToServerGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "route_connections_byte_to_server",
			Help: "Route connections byte to server",
		}, []string{"name", "router_hostname", "source_address", "destination_address"})
	routeConnectionsTimeStartedGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "route_connections_time_started",
			Help: "Route connections time started",
		}, []string{"name", "router_hostname", "source_address", "destination_address"})
	routeConnectionsTimeConnectedToServerGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "route_connections_time_connected_to_server",
			Help: "Route connections time connected to server",
		}, []string{"name", "router_hostname", "source_address", "destination_address"})
	routeConnectionsTimeLastSentToServerGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "route_connections_time_last_sent_to_server",
			Help: "Route connections time last sent to server",
		}, []string{"name", "router_hostname", "source_address", "destination_address"})
	routeConnectionsTimeLastReceivedFromServerGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "route_connections_time_last_received_from_server",
			Help: "Route connections time last received from server",
		}, []string{"name", "router_hostname", "source_address", "destination_address"})
)

func init() {
	prometheus.MustRegister(
		routerStatusGauge,
		metadataGauge,
		metadataConfigGauge,
		metadataConfigNodeGauge,
		metadataStatusGauge,
		routeGauge,
		routeActiveConnectionsGauge,
		routeTotalConnectionsGauge,
		routeBlockedHostsGauge,
		routeHealthGauge,
		routeDestinationsGauge,
		routeConnectionsByteFromServerGauge,
		routeConnectionsByteToServerGauge,
		routeConnectionsTimeStartedGauge,
		routeConnectionsTimeConnectedToServerGauge,
		routeConnectionsTimeLastSentToServerGauge,
		routeConnectionsTimeLastReceivedFromServerGauge,
	)
}

var (
	port = os.Getenv("MYSQLROUTER_EXPORTER_PORT")
	url  = os.Getenv("MYSQLROUTER_EXPORTER_URL")
	user = os.Getenv("MYSQLROUTER_EXPORTER_USER")
	pass = os.Getenv("MYSQLROUTER_EXPORTER_PASS")
)

func main() {
	flag.Parse()

	if url == "" || user == "" || pass == "" {
		panic("The environment missing.\n" +
			"MYSQLROUTER_EXPORTER_URL, MYSQLROUTER_EXPORTER_USER and MYSQLROUTER_EXPORTER_PASS is required.")
	}

	mr, err := mysqlrouter.New(url, user, pass)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			// Router
			router, err := mr.GetRouterStatus()
			if err != nil {
				panic(err)
			}
			routerStatusGauge.WithLabelValues(strconv.Itoa(router.ProcessID), router.ProductEdition, router.TimeStarted.String(), router.Version, router.Hostname)

			// Metadata
			metadata, err := mr.GetAllMetadata()
			if err != nil {
				panic(err)
			}
			for _, m := range metadata {
				metadataGauge.WithLabelValues(m.Name)

				// config
				mc, err := mr.GetMetadataConfig(m.Name)
				if err != nil {
					panic(err)
				}
				metadataConfigGauge.WithLabelValues(m.Name, mc.ClusterName, strconv.Itoa(mc.TimeRefreshInMs), mc.GroupReplicationID)

				// config node
				for _, node := range mc.Nodes {
					metadataConfigNodeGauge.WithLabelValues(m.Name, router.Hostname, mc.ClusterName, node.Hostname, strconv.Itoa(node.Port))
				}

				// status
				ms, err := mr.GetMetadataStatus(m.Name)
				if err != nil {
					panic(err)
				}
				metadataStatusGauge.WithLabelValues(m.Name, strconv.Itoa(ms.RefreshFailed), ms.TimeLastRefreshSucceeded.String(), ms.LastRefreshHostname, strconv.Itoa(ms.LastRefreshPort))
			}

			// Routes
			routes, err := mr.GetAllRoutes()
			if err != nil {
				panic(err)
			}
			for _, route := range routes {
				routeGauge.WithLabelValues(route.Name)

				rs, err := mr.GetRouteStatus(route.Name)
				if err != nil {
					panic(err)
				}
				routeActiveConnectionsGauge.WithLabelValues(route.Name, router.Hostname).Set(float64(rs.ActiveConnections))
				routeTotalConnectionsGauge.WithLabelValues(route.Name, router.Hostname).Set(float64(rs.TotalConnections))
				routeBlockedHostsGauge.WithLabelValues(route.Name, router.Hostname).Set(float64(rs.BlockedHosts))

				rh, err := mr.GetRouteHealth(route.Name)
				if err != nil {
					panic(err)
				}
				if rh.IsAlive {
					routeHealthGauge.WithLabelValues(route.Name, router.Hostname).Set(float64(1))
				} else {
					routeHealthGauge.WithLabelValues(route.Name).Set(float64(0))
				}

				rd, err := mr.GetRouteDestinations(route.Name)
				if err != nil {
					panic(err)
				}
				for _, d := range rd {
					routeDestinationsGauge.WithLabelValues(route.Name, d.Address, strconv.Itoa(d.Port))
				}

				rc, err := mr.GetRouteConnections(route.Name)
				if err != nil {
					panic(err)
				}
				for _, c := range rc {
					routeConnectionsByteFromServerGauge.WithLabelValues(route.Name, router.Hostname, c.SourceAddress, c.DestinationAddress).Set(float64(c.BytesFromServer))
					routeConnectionsByteToServerGauge.WithLabelValues(route.Name, router.Hostname, c.SourceAddress, c.DestinationAddress).Set(float64(c.BytesToServer))
					routeConnectionsTimeStartedGauge.WithLabelValues(route.Name, router.Hostname, c.SourceAddress, c.DestinationAddress).Set(float64(c.TimeStarted.Unix() * 1000))
					routeConnectionsTimeConnectedToServerGauge.WithLabelValues(route.Name, router.Hostname, c.SourceAddress, c.DestinationAddress).Set(float64(c.TimeConnectedToServer.Unix() * 1000))
					routeConnectionsTimeLastSentToServerGauge.WithLabelValues(route.Name, router.Hostname, c.SourceAddress, c.DestinationAddress).Set(float64(c.TimeLastSentToServer.Unix() * 1000))
					routeConnectionsTimeLastReceivedFromServerGauge.WithLabelValues(route.Name, router.Hostname, c.SourceAddress, c.DestinationAddress).Set(float64(c.TimeLastReceivedFromServer.Unix() * 1000))
				}
			}
			time.Sleep(60 * time.Second)
		}
	}()

	if port == "" {
		port = "49152"
	}

	log.Printf("listen: %s\n", "0.0.0.0:"+port)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
