package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// NetInfo 表示网卡信息和IP地址
type NetInfo struct {
	InterfaceName string   `json:"interface_name"`
	MACAddress    string   `json:"mac_address"`
	IPv4Addresses []string `json:"ipv4_addresses"`
	IPv6Addresses []string `json:"ipv6_addresses"`
}

// 获取所有网卡的信息
func getNetworkInterfaces() ([]NetInfo, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	var netInfos []NetInfo
	for _, iface := range interfaces {
		netInfo := NetInfo{
			InterfaceName: iface.Name,
			MACAddress:    iface.HardwareAddr.String(),
		}

		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Printf("failed to get addresses for interface %s: %v\n", iface.Name, err)
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP
			if ip.To4() != nil {
				netInfo.IPv4Addresses = append(netInfo.IPv4Addresses, ip.String())
			} else if ip.To16() != nil {
				netInfo.IPv6Addresses = append(netInfo.IPv6Addresses, ip.String())
			}
		}

		netInfos = append(netInfos, netInfo)
	}

	return netInfos, nil
}

// 发送数据到API
func sendToAPI(data []NetInfo) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	resp, err := http.Post("https://setip.eu.org/api/netinfo", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response: %s", resp.Status)
	}

	fmt.Println("Data sent successfully")
	return nil
}

// 比较两个NetInfo切片是否相同
func compareNetInfos(old, new []NetInfo) bool {
	if len(old) != len(new) {
		return false
	}

	for i := range old {
		if old[i].InterfaceName != new[i].InterfaceName ||
			old[i].MACAddress != new[i].MACAddress ||
			!compareStringSlices(old[i].IPv4Addresses, new[i].IPv4Addresses) ||
			!compareStringSlices(old[i].IPv6Addresses, new[i].IPv6Addresses) {
			return false
		}
	}

	return true
}

// 比较两个字符串切片是否相同
func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func main() {
	var previousNetInfos []NetInfo

	for {
		currentNetInfos, err := getNetworkInterfaces()
		if err != nil {
			fmt.Printf("Error getting network interfaces: %v\n", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		if !compareNetInfos(previousNetInfos, currentNetInfos) {
			fmt.Println("Network configuration changed, sending data to API...")
			if err := sendToAPI(currentNetInfos); err != nil {
				fmt.Printf("Error sending data to API: %v\n", err)
			}
			previousNetInfos = currentNetInfos
		} else {
			fmt.Println("No change in network configuration, sending empty data...")
			if err := sendToAPI([]NetInfo{}); err != nil {
				fmt.Printf("Error sending empty data to API: %v\n", err)
			}
		}

		time.Sleep(1 * time.Minute)
	}
}
