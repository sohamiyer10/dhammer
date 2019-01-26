package main

import (
	"flag"
	"github.com/ipchama/dhammer/config"
	"github.com/ipchama/dhammer/hammer"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	options := &config.Options{}

	options.Handshake = flag.Bool("handshake", true, "Attempt full handshakes")
	options.DhcpInfo = flag.Bool("dhcpinfo", false, "Blast DHCPINFO packets, but don't complete the handshake. NOT YET IMPLEMENTED.")
	options.DhcpBroadcast = flag.Bool("dhcp-broadcast", true, "Set the broadcast bit.")
	options.EthernetBroadcast = flag.Bool("ethernet-broadcast", true, "Use ethernet broadcasting.")
	options.Release = flag.Bool("release", false, "Release leases - NOT YET IMPLEMENTED")

	options.RequestsPerSecond = flag.Int("rps", 0, "Max number of packets per second. 0 == unlimited.")
	options.MaxLifetime = flag.Int("maxlife", 0, "How long to run. 0 == forever")
	options.MacCount = flag.Int("mac-count", 1, "Number of unique MAC addresses to pre-generate.")

	options.StatsRate = flag.Int("stats-rate", 5, "How frequently to display stats (seconds).")

	relayIP := flag.String("relay-source-ip", "", "Source IP for relayed requests.  relay-source-ip AND relay-target-server-ip must be set for relay mode.")
	targetServerIP := flag.String("relay-target-server-ip", "", "Target/Destination IP for relayed requests.  relay-source-ip AND relay-target-server-ip must be set for relay mode.")

	flag.Var(&options.AdditionalDhcpOptions, "dhcp-option", "Additional DHCP option to send out in the discover. Can be used multiple times. Format: <option num>:<RFC4648-base64-encoded-value>")

	options.InterfaceName = flag.String("interface", "eth0", "Interface name for listening and sending.")
	gatewayMAC := flag.String("gateway-mac", "de:ad:be:ef:f0:0d", "MAC of the gateway.")
	flag.Parse()

	var err error

	options.RelaySourceIP = net.ParseIP(*relayIP)
	options.RelayTargetServerIP = net.ParseIP(*targetServerIP)

	if options.RelaySourceIP != nil && options.RelayTargetServerIP != nil {
		options.DhcpRelay = true
	}

	options.GatewayMAC, err = net.ParseMAC(*gatewayMAC)
	if *options.StatsRate <= 0 {
		*options.StatsRate = 5
	}

	if err != nil {
		panic(err)
	}

	Hammer := hammer.New(options)
	err = Hammer.Init()

	if err != nil {
		panic(err)
	}

	osSigChann := make(chan os.Signal)
	signal.Notify(osSigChann, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = <-osSigChann
		Hammer.Stop()
	}()

	err = Hammer.Run()

	if err != nil {
		panic(err)
	}
}