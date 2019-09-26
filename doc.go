/*
Package gourmet is an extendable network analysis and intrusion detection system.

Gourmet is designed to be fast, simple, and customized. To customize your Gourmet sensor, you can
implement existing analyzers, or create your own.

The Gourmet Project repository contains analyzers for the following protocols:

  * HTTP (https://github.com/gourmetproject/http_analyzer)
  * DNS (https://github.com/gourmetprojecct/dns_analyzer)

In order to customize Gourmet, you must customize the project's config.yml file. The default
contents of the config.yml file are below.

  type: libpcap
  promiscuous: true
  interface: eth0
  snapshot_length: 262144
  log_file: /log/gourmet.log
  analyzers:
    - "github.com/gourmetproject/simple_analyzer"

Usage With No Analyzers

By default, gourmet analyzes Ethernet packets and logs basic information about the connections. This
information is contained in a Connection type. This Connection type is marshalled into a JSON object
and appended to the log file.

For UDP connections, each packet is transformed into a Connection object. However, for TCP
connections, the stream is first reassembled and then turned into a Connection object.

Usage With Analyzers

If you wish to add an analyzer to Gourmet, you must add the analyzer repo URL to the config.yml
file.

Creating Your Own Analyzer

Analyzers are an implementation of the Analyzer interface. They are written as a Go plugin. More
information about Go plugins can be found here: https://golang.org/pkg/plugin.

Example custom analyzers can be found in the Gourmet Project repository at
https://github.com/gourmetproject. simple_analyzer is the best one to start with.
 */
package gourmet
