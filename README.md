<h1 align="center">

![Gourmet Gopher](https://raw.githubusercontent.com/gourmetproject/gourmet/master/gourmet.png)

Gourmet

</h1>
<h4 align="center">
	An exquisite network traffic analysis framework
	<br>
	Fast, simple, and customizable
</h4>

# Overview

### Features
- Libpcap support
- AF_PACKET support
- Zero copy packet processing (fast!)
- Automatic TCP stream reassembly
- Berkeley Packet Filter support (currently only for libpcap)
- Easily extendable through Go Plugins (see Analyzers section below)

### Upcoming Features
- BPF support for AF_PACKET
- Binary release w/ command-line configuration

# Usage

Gourmet is not yet finished. But if you would like to give it a test ride, you can do the following:
```
git clone https://github.com/gourmetproject/gourmet
cd gourmet
docker-compose up
```
Make sure you change the `interface` argument in `config.yml` to the network interface on your host
machine that you want capture traffic on. Gourmet will log all captured traffic to `gourmet.log`.

# Design
### Written in Go
Gourmet is designed from the ground up in Go, [the number one language developers want to learn
in 2019](https://jaxenter.com/go-number-one-for-2019-hackerrank-report-155161.html). It utilizes
Google's [gopacket](https://github.com/google/gopacket) library to quickly decode and analyze
large amounts of network traffic. Go makes it fast, easy to maintain, and
[not C/C++](http://trevorjim.com/c-and-c++-are-dead-like-cobol-in-2017/).

### Highly Concurrent
One of Go's shining features is [goroutines](https://golangbot.com/goroutines/). Goroutines are
simply functions that run concurrently with other functions. They are much more lightweight,
flexible, and easy to work with than standard threads. Goroutines communicate with each other using
[channels](https://golangbot.com/channels/). Channels make it extremely simple to synchronize
multithreaded Go programs. 

These two language paradigms dramatically improve the speed, memory efficiency, and simplicity of
concurrently processing thousands, or even millions, of packets per second.

### Easily Customized through Go plugins
Go 1.8, released in February 2017, introduced a new
[plugin build mode](https://golang.org/pkg/plugin/). This build mode allows Go programs (and C
programs, through [cgo](https://golang.org/cmd/cgo/)) to export symbols that are loaded and
resolved by other Go programs at runtime. The Gourmet Project uses plugins as a way to load custom
analyzers passed to the Gourmet sensor at runtime through a YAML configuration file defined by the
user. More information how developers can create their own analyzers as Go plugins can be found
below.

# Analyzers
The Gourmet Project consists of the core Gourmet network sensor and a multitude of common
protocol analyzers implemented as Go plugins. We provide a simple interface for other third-party
developers to create and share their own analyzers as Go plugins.

In order to create your own analyzer, you must implement the Analyzer interface. This interface is
fully documented in the
[Gourmet documentation](https://godoc.org/github.com/gourmetproject/gourmet). A simple
example can be found in the [simple_analyzer](https://github.com/gourmetproject/simple_analyzer)
repository.

In order to implement the interface, you must create a new struct that has a Filter and Analyze
function.

### Filter
The Filter function takes a `*gourmet.Connection` object pointer as a parameter, determines
whether the analyzer should analyze the connection, and returns true or false. The logic contained
within the Filter function should be **as simple as possible to filter out irrelevant packets or
TCP streams**. For example, if you want to write an Analyzer that only looks at DNS traffic, then
your filter function should return true if the source or destination port is 53, and false
otherwise.

### Analyze
The Analyze function takes a gourmet Connection object as a parameter, conducts whatever logic
necessary to analyze that connection, and returns an implementation of the Result interface. A
Result object can be any data structure you like, such as a string, map, array, or struct. The
Result interface only requires you implement the Key function, which returns a string. This string
is used as the key value when we add the Result object to the JSON log for the Connection.

# Analyzer List

- [HTTP Analyzer](https://github.com/gourmetproject/httpanalyzer) - Logs information about HTTP traffic
- [DNS Analyzer](https://github.com/gourmetproject/dnsanalyzer) - Logs information about DNS traffic
- [Simple Analyzer](https://github.com/gourmetproject/simpleanalyzer) - Logs the number of bytes in the connection payload
- [Bedtime Analyzer](https://github.com/gourmetproject/bedtimeanalyzer) - If a specificed domain (such as Netflix) was accessed between certain hours of the day, a Slack bot sends you a message
   - Good example of analyzers depending on other analyzers and using the `init()` function to maintain state.

# Gourmet vs. Zeek (aka Bro)
It is no secret that Zeek is the top choice for network security monitoring.  One of the goals of
this project is to provide an alternative to Zeek. The table below illustrates some key differences
between the two projects.

| Feature          | Gourmet                                                       | Zeek                                                                           |
|------------------|---------------------------------------------------------------|------------------------------------------------------------------------------------|
| Log format       | Single JSON file; each connection is a root-level JSON object | Multiple CSV files; connection data across files is linked through connection UIDs |
| Language         | Pure Go                                                       | Zeek scripting language as a wrapper around C/C++                                   |
| Customization    | Go Plugins                                                    | Zeek scripts                                                                        |
| Production-ready | Not yet, work in progress                                     | Yes                                                                                |
| Open Source      | Yes                                                           | Yes                                                                                |
| Multithreaded    | Yes                                                           | No (see [Zeek Cluster](https://docs.zeek.org/en/stable/cluster/index.html))        |

# Contact Us

<a
href="https://join.slack.com/t/gourmetproject/shared_invite/enQtNzczMjQ4MzgzMTg5LTRjOTllNjc2MzNhMDQyNDdiMWQwZjQ5OTEwZDEyYjhiNWEwZjI3M2Y2MzExMGQ1ZjNkZjlkMjlkYTc3ZDZmN2Y">
	<img
		src="https://cdn.appstorm.net/web.appstorm.net/web/files/2013/10/slack_icon.png"
		alt="Slack icon"
		width="40"
	>
</a>

# Support Us

[![Patreon][patreon-badge]][patreon-link]

[patreon-badge]: https://img.shields.io/endpoint.svg?url=https%3A%2F%2Fshieldsio-patreon.herokuapp.com%2Fkvasirlabs&style=flat-round
[patreon-link]: https://patreon.com/kvasirlabs
