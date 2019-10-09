package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"

	"github.com/ghodss/yaml"
	"github.com/google/gopacket/pcap"
	"github.com/gourmetproject/gourmet"
)

var (
	flagConfig = flag.String("c", "config.yml", "Gourmet configuration file")
)

func main() {
	var c *gourmet.Config
	var err error
	flag.Parse()
	c, err = parseConfigFile(*flagConfig)
	if err != nil {
		log.Fatal(err)
	}
	if c.MaxCores != 0 && c.MaxCores < runtime.NumCPU() {
		runtime.GOMAXPROCS(c.MaxCores)
	} else if c.MaxCores != 0 {
		fmt.Println(fmt.Errorf("[!] Warning: max_cores argument is invalid. Using %d cores instead", runtime.NumCPU()))
	}

	setDefaults(c)
	err = validateConfig(c)
	if err != nil {
		log.Fatal(err)
	}
	gourmet.Start(c)
}

func parseConfigFile(cf string) (c *gourmet.Config, err error) {
	c = &gourmet.Config{}
	contents, err := ioutil.ReadFile(cf)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(contents, c)
	if err != nil {
		return nil, err
	}
	return c, err
}

func setDefaults(c *gourmet.Config) {
	if c.SnapLen == 0 {
		c.SnapLen = 262144
	}
	if c.LogFile == "" {
		c.LogFile = "gourmet.log"
	}
	if c.InterfaceType == "" {
		c.InterfaceType = "libpcap"
	}
}

func validateConfig(c *gourmet.Config) (err error) {
	if err = validateInterface(c.Interface); err != nil {
		return err
	}
	if err = validateSnapshotLength(c.SnapLen); err != nil {
		return err
	}
	return nil
}

func validateInterface(iface string) error {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}
	for _, device := range devices {
		if device.Name == iface {
			return nil
		}
	}
	return errors.New("specified network interface does not exist")
}

func validateSnapshotLength(snapLen int) error {
	if snapLen < 64 {
		return errors.New("minimum snapshot length is 64")
	}
	if snapLen > 4294967295 {
		return errors.New("snapshot length must be an unsigned 32-bit integer")
	}
	return nil
}
