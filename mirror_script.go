package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"gopkg.in/yaml.v2"
	// "runtime"
)

func getConf(file string) (*Config, error) {
    var c Config

	yamlFile, err := os.ReadFile(file)
	if err != nil {
        return nil, err
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
        return nil, err
	}

	return &c, nil
}

func execute(command string, file string) error {
	out, err := exec.Command(command, file).Output()

	if err != nil {
	    return err
	}

	output := string(out[:])
	fmt.Println(output)

    return nil
}

func configVxLAN(config *Config) error {
    f, err := os.Create("vxlan.sh")
    if err != nil {
        return err
    }

    defer f.Close()

    command := fmt.Sprintf("ip link add %s type vxlan id %d dev %s remote %s dstport 4789\n", "vxlan" + strconv.FormatInt(config.VxLANID, 10), config.VxLANID, config.SourceInterface, config.Sensor)
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }

    command = fmt.Sprintf("ip link set %s up", "vxlan" + strconv.FormatInt(config.VxLANID, 10))
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }

    err = execute("sh", "vxlan.sh")
    if err != nil {
        return err
    }

    return nil
}

func configIngress(config *Config) error {
	action := ""
	if config.Filters[0].Action == "deny" {
		action = "drop"
	} else {
		action = "pass"
	}

    f, err := os.Create("ingress.sh")
    if err != nil {
        return err
    }

    defer f.Close()

	command := fmt.Sprintf("tc qdisc add dev %s handle ffff: ingress\n", config.MirrorInterface.Name)
	_, err = f.WriteString(command)
    if err != nil {
        return err
    }
	command = fmt.Sprintf("tc filter add dev %s parent ffff: protocol %s prio 1 u32 match ip src %s action %s\n", config.MirrorInterface.Name, config.Filters[0].Protocol, config.Filters[0].IP, action)
	_, err = f.WriteString(command)
    if err != nil {
        return err
    }
	command = fmt.Sprintf("tc filter add dev %s parent ffff: protocol %s prio 10 u32 match u32 0 0 action mirred egress mirror dev %s\n", config.MirrorInterface.Name, config.Filters[0].Protocol, "vxlan" + strconv.FormatInt(config.VxLANID, 10))
	_, err = f.WriteString(command)
    if err != nil {
        return err
    }

    err = execute("sh", "ingress.sh")
    if err != nil {
        return err
    }

    return nil
}

func configEgressMultiple(config *Config) error {
	action := ""
	if config.Filters[0].Action == "deny" {
		action = "drop"
	} else {
		action = "pass"
	}

    f, err := os.Create("egress_multiple.sh")
    if err != nil {
        return err
    }

    defer f.Close()

    command := fmt.Sprintf("tc qdisc add dev %s handle 1: root prio\n", config.MirrorInterface.Name)
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }
    command = fmt.Sprintf("tc filter add dev %s parent 1: protocol %s prio 1 u32 match ip dst %s action %s\n", config.MirrorInterface.Name, config.Filters[0].Protocol, config.Sensor, action)
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }
    command = fmt.Sprintf("tc filter add dev %s parent 1: protocol %s prio 10 u32 match u32 0 0 action mirred egress mirror dev %s\n", config.MirrorInterface.Name, config.Filters[0].Protocol, "vxlan" + strconv.FormatInt(config.VxLANID, 10))
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }

    err = execute("sh", "egress_multiple.sh")
    if err != nil {
        return err
    }

    return nil
}

func configEgressSingle(config *Config) error {
	action := ""
	if config.Filters[0].Action == "deny" {
		action = "drop"
	} else {
		action = "pass"
	}

    f, err := os.Create("egress_single.sh")
    if err != nil {
        return err
    }

    command := fmt.Sprintf("tc qdisc add dev %s handle 1: root prio\n", config.MirrorInterface.Name)
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }
    command = fmt.Sprintf("tc filter add dev %s parent 1: protocol %s u32 match u32 0 0 action mirred egress mirror dev lo\n", config.MirrorInterface.Name, config.Filters[0].Protocol)
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }
    command = "tc qdisc add dev lo handle ffff: ingress\n"
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }
    command = fmt.Sprintf("tc filter add dev lo parent ffff: protocol %s prio 1 u32 match ip dst %s action %s\n", config.Filters[0].Protocol, config.Sensor, action)
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }
    command = fmt.Sprintf("tc filter add dev lo parent ffff: protocol %s prio 1 u32 match ip src 127.0.0.1 action %s\n", config.Filters[0].Protocol, action)
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }
    command = fmt.Sprintf("tc filter add dev lo parent ffff: protocol %s prio 1 u32 match ip dst 127.0.0.1 action %s\n", config.Filters[0].Protocol, action)
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }
    command = fmt.Sprintf("tc filter add dev lo parent ffff: protocol %s u32 match u32 0 0 action mirred egress mirror dev %s\n", config.Filters[0].Protocol, "vxlan" + strconv.FormatInt(config.VxLANID, 10))
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }

    err = execute("sh", "egress_single.sh")
    if err != nil {
        return err
    }

    return nil
}

func generateRollbackScript(config *Config) error {
    f, err := os.Create("rollback.sh")
    if err != nil {
        return err
    }

    command := fmt.Sprintf("tc qdisc del dev %s root\n", config.MirrorInterface.Name)
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }
    command = fmt.Sprintf("tc qdisc del dev %s ingress\n", config.MirrorInterface.Name)
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }
    _, err = f.WriteString("tc qdisc del dev lo ingress\n")
    if err != nil {
        return err
    }
    command = fmt.Sprintf("ip link del %s\n", "vxlan" + strconv.FormatInt(config.VxLANID, 10))
    _, err = f.WriteString(command)
    if err != nil {
        return err
    }
    return nil
}

func rollBack(config *Config) error {
    err := execute("sh", "rollback.sh")
    if err != nil {
        return err
    }

    return nil
}

func Mirror() {
    fmt.Println("--------------------------------------")
    fmt.Println("Getting config file...")
	c, err := getConf("/usr/bin/mirror_config.yaml")
    if err != nil {
        log.Fatalf("getConf err: %v", err)
    }
    fmt.Println("=> Done")

    generateRollbackScript(c)

    fmt.Println("--------------------------------------")
    fmt.Print("Configuring VxLAN...")
    err = configVxLAN(c)
    if err != nil {
        rollBack(c)
        log.Fatalf("configVxLAN err: %v", err)
    }
    err = execute("rm", "vxlan.sh")
    if err != nil {
        log.Fatal("Fail to delete file!")
    }
    fmt.Println("=> Done")

    fmt.Println("--------------------------------------")
    fmt.Print("Configuring Ingress...")
    if c.MirrorInterface.Ingress {
        err = configIngress(c)
        if err != nil {
            rollBack(c)
            log.Fatalf("configIngress err: %v", err)
        }
    }
    err = execute("rm", "ingress.sh")
    if err != nil {
        log.Fatal("Fail to delete file!")
    }
    fmt.Println("=> Done")

    fmt.Println("--------------------------------------")
    fmt.Print("Configuring Egress...")
    if c.MirrorInterface.Egress {
        if c.MirrorInterface.Name == c.SourceInterface {
            err = configEgressSingle(c)
            if err != nil {
                rollBack(c)
                log.Fatalf("configEgressSingle err: %v", err)
            }
            err = execute("rm", "egress_single.sh")
            if err != nil {
                log.Fatal("Fail to delete file!")
            }
        } else {
            err = configEgressMultiple(c)
            if err != nil {
                rollBack(c)
                log.Fatalf("configEgressMultiple err: %v", err)
            }
            err = execute("rm", "egress_multiple.sh")
            if err != nil {
                log.Fatal("Fail to delete file!")
            }
        }
    }
    fmt.Println("=> Done")

}
