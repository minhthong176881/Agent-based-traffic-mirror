package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) >=2 && os.Args[1] == "config" {
		reader := bufio.NewReader(os.Stdin)
		configExisted := false
		if _, err := os.Stat("/usr/bin/mirror_config.yaml"); err == nil {
			configExisted = true
		} else if errors.Is(err, os.ErrNotExist) {
			configExisted = false
		}

		if configExisted {
			fmt.Println("Configuration file found!")
			fmt.Print("Do you want to use existed configuration file? (y/N): ")
			agree, _ := reader.ReadString('\n')
			if agree == "n\n" || agree == "N\n" {
				Generate()
				_, err := exec.Command("cp", "mirror_config.yaml", "/usr/bin/").Output()
				if err != nil {
					log.Fatalf("err: %v", err)
				}
			}
		} else {
			Generate()
			_, err := exec.Command("cp", "mirror_config.yaml", "/usr/bin/").Output()
			if err != nil {
				log.Fatalf("err: %v", err)
			}
		}
	}

	Mirror()
}
