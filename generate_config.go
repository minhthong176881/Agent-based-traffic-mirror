package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

type MirrorInterface struct {
	Name    string `yaml:"interface"`
	Ingress bool   `yaml:"ingress"`
	Egress  bool   `yaml:"egress"`
}

type Filter struct {
	IP       string `yaml:"ip"`
	Port     int64  `yaml:"port"`
	Protocol string `yaml:"protocol"`
	Priority int64  `yaml:"priority"`
	Action   string `yaml:"action"`
}

type Config struct {
	Version         int64           `yaml:"version"`
	Sensor          string          `yaml:"sensor"`
	MirrorInterface MirrorInterface `yaml:"mirror_interface"`
	SourceInterface string          `yaml:"source_interface"`
	VxLANID         int64           `yaml:"vxlan_id"`
	Filters         []Filter        `yaml:"filters"`
}

func getSensor(reader *bufio.Reader) (string, error) {
	sensor := ""
	ipValid := false
	fmt.Println("--------------------------------------")
	for {
		fmt.Print("Enter sensor IP (e.g: 192.168.74.133): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		input = strings.TrimSuffix(input, "\n")
		ipValid = checkIPAddress(input)
		if ipValid {
			sensor = input
			break
		}
	}

	return sensor, nil
}

func getMirrorInterface(reader *bufio.Reader) (*MirrorInterface, error) {
	var mirror MirrorInterface
	fmt.Println("--------------------------------------")
	for {
		fmt.Print("Enter interface to be mirrored (e.g: eth0): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if input == "\n" {
			fmt.Println("Mirrored interface is required!")
			continue
		}
		if checkInterface(strings.TrimSuffix(input, "\n")) {
			mirror.Name = strings.TrimSuffix(input, "\n")
			break
		}
		fmt.Println("Interface " + strings.TrimSuffix(input, "\n") + " not found!")
	}

	fmt.Print("Mirror ingress traffic? (y/N): ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if input == "y\n" || input == "Y\n" {
		mirror.Ingress = true
	} else {
		mirror.Ingress = false
	}

	fmt.Print("Mirror egress traffic? (y/N): ")
	input, err = reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if input == "y\n" || input == "Y\n" {
		mirror.Egress = true
	} else {
		mirror.Egress = false
	}

	return &mirror, nil
}

func getSourceInterface(reader *bufio.Reader) (string, error) {
	fmt.Println("--------------------------------------")
	var sourceInterface string
	for {
		fmt.Print("Enter interface to copy mirrored traffic (e.g: eth0): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		if input == "\n" {
			fmt.Println("Source interface is required!")
			continue
		}
		if checkInterface(strings.TrimSuffix(input, "\n")) {
			sourceInterface = strings.TrimSuffix(input, "\n")
			break
		}
		fmt.Println("Interface " + strings.TrimSuffix(input, "\n") + " not found!")
	}

	return sourceInterface, nil
}

func getVxLANId() (int64, error) {
	var id int64
	fmt.Println("--------------------------------------")
	for {
		fmt.Print("Enter VxLAN Id (e.g: 108): ")
		_, err := fmt.Scanf("%d", &id)
		if err != nil {
			fmt.Println("VxLAN Id has to be an integer!")
			continue
		} else {
			break
		}
	}

	return id, nil
}

func getFilter(reader *bufio.Reader) (*Filter, error) {
	var filter Filter
	fmt.Println("--------------------------------------")
	ipValid := false
	for {
		fmt.Print("Enter IP address to be filtered (e.g: 192.168.74.133): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		input = strings.TrimSuffix(input, "\n")
		ipValid = checkIPAddress(input)
		if ipValid {
			filter.IP = input
			break
		}
	}

	for {
		var port int64
		fmt.Print("Enter port to be filtered: ")
		_, err := fmt.Scanf("%d", &port)
		if err != nil {
			fmt.Println("VxLAN Id has to be an integer!")
			continue
		} else {
			filter.Port = port
			break
		}
	}

	fmt.Print("Enter protocol to be filtered: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if input == "\n" {
		filter.Protocol = "all"
	} else {
		filter.Protocol = strings.TrimSuffix(input, "\n")
	}

	for {
		var prio int64
		fmt.Print("Enter priority of the filter: ")
		_, err := fmt.Scanf("%d", &prio)
		if err != nil {
			fmt.Println("Priority has to be an integer!")
			continue
		} else {
			filter.Priority = prio
			break
		}
	}

	fmt.Print("Enter action of the filter: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if input != "accept\n" && input != "deny\n" {
		filter.Action = "deny"
	} else {
		filter.Action = strings.TrimSuffix(input, "\n")
	}

	return &filter, nil
}

func getDefaultFilter(config *Config) (*Filter, error) {
	filter := Filter{
		IP:       config.Sensor,
		Port:     4789,
		Protocol: "all",
		Priority: 1,
		Action:   "deny",
	}
	return &filter, nil
}

func checkIPAddress(ip string) bool {
	if net.ParseIP(ip) == nil {
		fmt.Println("Invalid IP address!")
		return false
	} else {
		return true
	}
}

func checkInterface(iface string) bool {
	out, err := exec.Command("ip", "addr").Output()

	if err != nil {
		fmt.Println(err)
		return false
	}

	output := string(out[:])
	if strings.Contains(output, iface) {
		return true
	} else {
		return false
	}
}

func Generate() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("--------------------------------------")
	fmt.Println("Generating configuration file...")
	var config Config

	config.Version = 2

	sensor, err := getSensor(reader)
	if err != nil {
		log.Fatalf("getSensor err: %v", err)
	}
	config.Sensor = sensor

	mirrorIf, err := getMirrorInterface(reader)
	if err != nil {
		log.Fatalf("getMirrorInterface err: %v", err)
	}
	config.MirrorInterface = *mirrorIf

	sourceIf, err := getSourceInterface(reader)
	if err != nil {
		log.Fatalf("getSourceInterface err: %v", err)
	}
	config.SourceInterface = sourceIf

	vxlan, err := getVxLANId()
	if err != nil {
		log.Fatalf("getVxLANId err: %v", err)
	}
	config.VxLANID = vxlan

	// filter, err := getFilter(reader)
	// if err != nil {
	// 	log.Fatalf("getFilter err: %v", err)
	// }
	filter, _ := getDefaultFilter(&config)
	config.Filters = append(config.Filters, *filter)

	data, err := yaml.Marshal(&config)

	if err != nil {

		log.Fatal(err)
	}

	err = os.WriteFile("mirror_config.yaml", data, 0)

	if err != nil {

		log.Fatal(err)
	}

	fmt.Println("--------------------------------------")
	fmt.Println("=> Done")
}
