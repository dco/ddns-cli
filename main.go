package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const Banner = `
	┌─┐┌─┐┌┬┐┬┌─┐
	└─┐├┤  │ │├─┘
	└─┘└─┘ ┴ ┴┴      %s
___________________________________
`

var (
	Version   string
	ipsMap    = sync.Map{}
	APIURL, _ = base64.StdEncoding.DecodeString("aHR0cHM6Ly9kZG5zLnNldGlwLmV1Lm9yZy9uaWM=")
	WEB, _    = base64.StdEncoding.DecodeString("aHR0cHM6Ly9zZXRpcC5ldS5vcmc=")
	cid       string
	loglevel  string
	version   bool
	DNS       = []string{
		"localhost",
		"1.1.1.1",
		"223.5.5.5",
		"119.29.29.29",
		"8.8.8.8",
	}
)

func init() {
	// Define command-line flags
	flag.StringVar(&cid, "cid", "", "Required parameter: Client ID")
	flag.StringVar(&loglevel, "loglevel", "error", "Set log level Supported log levels: debug, info, warn, error")
	flag.BoolVar(&version, "version", false, "Print version number")
	fmt.Printf(Banner, Version)
	fmt.Println("Author: dco")
	fmt.Printf("Document: %s\n", string(WEB))
	fmt.Println(strings.Repeat("_", 35))
	forQueryDNS()
}

func main() {
	flag.Parse()

	// Check for required parameter: cid
	if cid == "" && !version {
		fmt.Println("Error: cid is a required parameter")
		flag.Usage()
		os.Exit(1)
	}

	// Handle version flag
	if version {
		fmt.Println("Version:", Version)
		os.Exit(1)
	}

	if loglevel, err := parseLogLevel(strings.ToLower(loglevel)); err != nil {
		fmt.Println("Parse log level error", err)
		fmt.Println("Supported log levels: debug, info, warn, error")
		os.Exit(1)
	} else {
		slog.SetLogLoggerLevel(loglevel)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Printf("\nGoodbye! Thank you for using [%s]!\n", WEB)
		os.Exit(0)
	}()

	run()
}

func run() {
	ipStr := "ipv4"
	var (
		ips        []string
		ipInfo     *IPInfo
		ipInfoJSON []byte
		err        error
		totalIPs   int
	)

	for {
		slog.Info("The new information push starts", "type", ipStr)

		if ipInfo, err = getIPInfo(); err != nil {
			slog.Error("get ip info error", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		ipInfo.ClientId = cid

		if ipInfoJSON, err = toJSON(ipInfo); err != nil {
			slog.Error("to json error")
			time.Sleep(5 * time.Second)
			continue
		}

		totalIPs = 0

		ipsMap.Range(func(key, value interface{}) bool {
			if ips, ok := value.([]string); ok {
				totalIPs += len(ips)
			}
			return true
		})

		if totalIPs == 0 {
			slog.Warn("No IPs found in the map, attempting to query DNS")

			forQueryDNS()
		}

		if ips, err = GetIPsFromMap(ipStr); err != nil {
			slog.Error("get ips from map error", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, ip := range ips {
			if _, err := RequestClient(string(APIURL), ip, "POST", ipInfoJSON); err != nil {
				slog.Warn("API request failed, attempting to retry")
				RemoveIPFromMap(ipStr, ip)
			} else {
				slog.Info("Information has been pushed")
				fmt.Printf("%v Information has been pushed\n", time.Now().Format("2006-01-02 15:04:05"))
				break
			}
		}

		if ipStr == "ipv4" {
			ipStr = "ipv6"
		} else {
			ipStr = "ipv4"
		}

		time.Sleep(30 * time.Second)
	}
}

// DNSQueryResult represents the result of a DNS query, containing both IPv4 and IPv6 addresses.
type DNSQueryResult struct {
	IPv4 []string
	IPv6 []string
}

// QueryDNS queries the IP addresses of the specified domain using a custom DNS server,
// and distinguishes between IPv4 and IPv6 addresses.
func QueryDNS(dnsServer, domain string) (DNSQueryResult, error) {
	var (
		err    error
		result = DNSQueryResult{}
		ips    = []net.IP{}
	)
	if dnsServer == "localhost" {
		if ips, err = net.LookupIP(domain); err != nil {
			return result, fmt.Errorf("failed to lookup IP locally for %s: %v", domain, err)
		}
	} else {
		resolver := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: 5 * time.Second,
				}
				return d.DialContext(ctx, "udp", dnsServer+":53")
			},
		}

		if ips, err = resolver.LookupIP(context.Background(), "ip", domain); err != nil {
			return result, fmt.Errorf("failed to lookup IP for %s: %v", domain, err)
		}
	}
	for _, ip := range ips {
		if ip.To4() != nil {
			result.IPv4 = append(result.IPv4, ip.String())
		} else if ip.To16() != nil {
			result.IPv6 = append(result.IPv6, ip.String())
		}
	}
	return result, nil
}

func forQueryDNS() {
	host, _ := url.Parse(string(APIURL))
	for _, dns := range DNS {
		if dnsIPs, err := QueryDNS(dns, host.Host); err != nil {
			slog.Warn("Failed to query DNS", "Warn", err)
			continue
		} else {
			AddIPToMap("ipv4", dnsIPs.IPv4)
			AddIPToMap("ipv6", dnsIPs.IPv6)
		}
	}
}

// RequestClient sends an HTTP request to the specified URL using a custom IP address
// and HTTP method. It supports both IPv4 and IPv6 addresses.
func RequestClient(urlStr, ip, method string, data []byte) (*http.Response, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("parsing URL error: %w", err)
	}

	var network string
	useIPv6 := strings.Contains(ip, ":")

	if useIPv6 {
		network = "tcp6"
		if ip[0] != '[' {
			ip = "[" + ip + "]"
		}
	} else {
		network = "tcp4"
	}

	addr := ip + ":443"

	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: 5 * time.Second}
			return dialer.DialContext(ctx, network, addr)
		},
	}

	client := &http.Client{Transport: transport}
	if method == "" {
		method = "GET"
	}

	if data == nil {
		data = []byte("")
	}
	req, err := http.NewRequest(method, urlStr, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("creating request error: %w", err)
	}
	req.Header.Set("content-type", "application/json")

	req.Host = u.Host

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request error: %w", err)
	}

	return resp, nil
}

// NICInfo stores information about a network interface, including its name and associated IP addresses.
type NICInfo struct {
	Name string   `json:"name"`
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
}

// IPInfo stores information about the machine's network interfaces, including their names
type IPInfo struct {
	NICs     []NICInfo `json:"nics"`
	ClientId string    `json:"client_id"`
}

// toJSON converts the IPInfo struct to a JSON byte slice.
func toJSON(ipInfo *IPInfo) ([]byte, error) {
	return json.Marshal(ipInfo)
}

// getIPInfo retrieves the IP addresses of all network interfaces on the machine,
// filtering out loopback and private IP addresses for both IPv4 and IPv6.
func getIPInfo() (*IPInfo, error) {
	// Retrieve all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %v", err)
	}

	ipInfo := &IPInfo{}

	// Iterate through each network interface
	for _, iface := range interfaces {
		nicInfo := NICInfo{
			Name: iface.Name,
		}

		// Retrieve the addresses associated with the interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Iterate through each address
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP

			// Filter out loopback and private addresses
			if ip.To4() != nil && !ip.IsLoopback() && !ip.IsPrivate() && isOtherLocalIPv4(ip) {
				nicInfo.IPv4 = append(nicInfo.IPv4, ip.String())
			} else if ip.To16() != nil && !ip.IsLoopback() && !ip.IsPrivate() && isGlobalUnicastIPv6(ip) {
				nicInfo.IPv6 = append(nicInfo.IPv6, ip.String())
			}
		}

		// Add the NICInfo to the result if it has any IP addresses
		if len(nicInfo.IPv4) > 0 || len(nicInfo.IPv6) > 0 {
			ipInfo.NICs = append(ipInfo.NICs, nicInfo)
		}
	}

	return ipInfo, nil
}

// isGlobalUnicastIPv6 checks if the given IPv6 address is a global unicast address.
// It filters out link-local, unique-local, multicast, loopback, unspecified, and documentation addresses.
func isGlobalUnicastIPv6(ip net.IP) bool {
	// Ensure the IP is an IPv6 address
	if ip.To16() == nil || ip.To4() != nil {
		return false
	}

	// Check for link-local addresses (fe80::/10)
	if ip[0] == 0xfe && (ip[1]&0xc0) == 0x80 {
		return false
	}

	// Check for unique-local addresses (fc00::/7)
	if ip[0] == 0xfc || ip[0] == 0xfd {
		return false
	}

	// Check for multicast addresses (ff00::/8)
	if ip[0] == 0xff {
		return false
	}

	// Check for loopback address (::1/128)
	if ip.Equal(net.IPv6loopback) {
		return false
	}

	// Check for unspecified address (::/128)
	if ip.Equal(net.IPv6unspecified) {
		return false
	}

	// Check for documentation addresses (2001:db8::/32)
	if len(ip) >= 4 && ip[0] == 0x20 && ip[1] == 0x01 && ip[2] == 0x0d && ip[3] == 0xb8 {
		return false
	}

	return true
}

func parseLogLevel(levelStr string) (slog.Level, error) {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level: %s", levelStr)
	}
}

// AddIPToMap adds an IP to the ipsMap
func AddIPToMap(ipType string, ip []string) {
	if value, ok := ipsMap.Load(ipType); !ok {
		ipsMap.Store(ipType, ip)
	} else {
		newValue := append(value.([]string), ip...)
		ipsMap.Store(ipType, newValue)
	}
}

// RemoveIPFromMap removes a specified IP from the ipsMap
func RemoveIPFromMap(ipType string, ipToRemove string) {
	value, ok := ipsMap.Load(ipType)
	if !ok {
		slog.Error("IP type does not exist", "type", ipType)
		return
	}

	ips, ok := value.([]string)
	if !ok {
		slog.Error("IP list format is incorrect", "type", ipType)
		return
	}

	// Filter out the IP to remove
	newIPs := make([]string, 0, len(ips))
	for _, ip := range ips {
		if ip != ipToRemove {
			newIPs = append(newIPs, ip)
		}
	}

	// Update the ipsMap
	ipsMap.Store(ipType, newIPs)
}

// GetIPsFromMap retrieves the IP list of the specified type from the ipsMap
func GetIPsFromMap(ipType string) ([]string, error) {
	value, ok := ipsMap.Load(ipType)
	if !ok {
		return nil, fmt.Errorf("IP type does not exist: %s", ipType)
	}

	ips, ok := value.([]string)
	if !ok {
		return nil, fmt.Errorf("IP list format is incorrect: %s", ipType)
	}

	return ips, nil
}

func isOtherLocalIPv4(ip net.IP) bool {
	if ip.To4() == nil {
		return false
	}

	ipv4 := ip.To4()
	if ipv4[0] >= 224 && ipv4[0] <= 239 {
		return false
	}

	if ipv4[0] == 169 && ipv4[1] == 254 {
		return false
	}

	return true
}
