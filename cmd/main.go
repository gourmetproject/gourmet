package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/google/gopacket/pcap"
	"github.com/gourmetproject/gourmet"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"strings"
)

type Config struct {
	InterfaceType string `yaml:"type"`
	Interface     string
	Promiscuous   bool
	SnapLen       int    `yaml:"snapshot_length"`
	Bpf           string
	Timeout       int
	LogFile       string `yaml:"log_file"`
	Analyzers     AnalyzerLinks
}

type AnalyzerLinks []string

func (al *AnalyzerLinks) String() string {
	return fmt.Sprintln("[" + strings.Join(*al, ", ") + "]")
}

func (al *AnalyzerLinks) Set(analyzer string) error {
	*al = append(*al, strings.TrimSpace(analyzer))
	return nil
}

var (
	analyzerLinks AnalyzerLinks
	flagInterfaceType = flag.String("y", "libpcap",
		"The packet processing technology for capture (libpcap, afpacket, or pf_ring)")
	flagInterface = flag.String("i", "", "The name of the network interface (required)")
	flagPromiscuous = flag.Bool("p", false, "Promiscuous mode")
	flagSnapLength = flag.Int("s", 262144, "The snapshot length for packet capture")
	flagBpf = flag.String("f", "", "Berkeley packet filter to apply to the capturing interface")
	flagConfig = flag.String("c", "",
		"Gourmet configuration. If this is set, all other command-line flags are ignored")
	flagLogFile = flag.String("l", "gourmet.log", "The output log file")
	flagTimeout = flag.Int("t", 0, "The number of seconds to run the sensor. Zero is no timeout.")
)

func main() {
	var c *Config
	var err error
	flag.Var(&analyzerLinks, "a",
		"Gourmet analyzer to enrich capture. This flag can be set more than once.")
	flag.Parse()
	if *flagConfig != "" {
		c, err = parseConfigFile(*flagConfig)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		c = &Config{
			InterfaceType: *flagInterfaceType,
			Interface:     *flagInterface,
			Promiscuous:   *flagPromiscuous,
			SnapLen:       *flagSnapLength,
			Bpf:           *flagBpf,
			Timeout:       *flagTimeout,
			LogFile:       *flagLogFile,
			Analyzers:     analyzerLinks,
		}
	}
	err = validateConfig(c)
	if err != nil {
		log.Fatal(err)
	}
	ifaceType, err := convertIfaceType(c.InterfaceType)
	if err != nil {
		log.Fatal(err)
	}
	analyzers, err := newAnalyzers(c.Analyzers)
	if err != nil {
		log.Fatal(err)
	}
	opts := &gourmet.SensorOptions{
		InterfaceName: c.Interface,
		InterfaceType: ifaceType,
		IsPromiscuous: c.Promiscuous,
		SnapLen:       uint32(c.SnapLen),
		Bpf:           c.Bpf,
		Timeout:       c.Timeout,
		LogFileName:   c.LogFile,
		Analyzers:     analyzers,
	}
	gourmet.Start(opts)
}

func parseConfigFile(cf string) (c *Config, err error) {
	c = &Config{}
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

func validateConfig(c *Config) (err error) {
	if err = validateInterface(c.Interface); err != nil {
		return err
	}
	if err = validateSnapshotLength(c.SnapLen); err != nil {
		return err
	}
	if err = validateTimeout(c.Timeout); err != nil {
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

func validateTimeout(timeout int) error {
	if timeout < 0 {
		return errors.New("timeout must be a positive integer, or zero for no timeout")
	}
	return nil
}

func convertIfaceType(ifaceType string) (gourmet.InterfaceType, error) {
	if ifaceType == "libpcap" {
		return gourmet.LibpcapType, nil
	} else if ifaceType == "afpacket" {
		return gourmet.AfpacketType, nil
	} else if ifaceType == "pf_ring" {
		return gourmet.PfringType, nil
	} else {
		return 0, errors.New("invalid interface type. Must be libpcap, afpacket, or pf_ring")
	}
}

func newAnalyzers(links []string) (analyzers []gourmet.Analyzer, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	var analyzerFiles []string
	for _, link := range links {
		fmt.Printf("[*] Installing %s\n", link)
		err = exec.Command("go", "get", "-u", "-d", link).Run()
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to install %s: %s", link, err.Error()))
		}
		analyzerFiles = append(analyzerFiles, fmt.Sprintf("%s/go/src/%s/main.go", homeDir, link))
	}
	if len(analyzerFiles) > 0 {
		for _, analyzerFile := range analyzerFiles {
			folderName := filepath.Dir(analyzerFile)
			fmt.Printf("[*] Building %s\n", filepath.Base(filepath.Dir(analyzerFile)))
			out, err := exec.Command("go", "build", "-buildmode=plugin", "-o",
				fmt.Sprintf("%s/main.so", filepath.Dir(analyzerFile)), analyzerFile).CombinedOutput()
			if err != nil {
				return nil, errors.New(
					fmt.Sprintf("failed to build %s: %s", analyzerFile, string(out)))
			}
			p, err := plugin.Open(fmt.Sprintf("%s/main.so", folderName))
			if err != nil {
				return nil, err
			}
			newAnalyzerFunc, err := p.Lookup("NewAnalyzer")
			if err != nil {
				return nil, err
			}
			analyzer := newAnalyzerFunc.(func() gourmet.Analyzer)()
			analyzers = append(analyzers, analyzer)
		}
	}
	return analyzers, nil
}