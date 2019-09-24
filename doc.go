/*
Package gourmet is an extendable network analysis and intrusion detection system.

gourmet contains example analyzers as sub-packages, along with examples of how to use gourmet. If
you plan to implement your own analyzer or third-party analyzers.

Usage With No Analyzers

By default, gourmet analyzes Ethernet packets and logs basic information about the connections. This
information is contained in a Connection type. This Connection type is marshalled into a JSON object
and appended to the log file.

For UDP connections, each packet is transformed into a Connection object. However, for TCP
connections, the stream is first reassembled and then turned into a Connection object. 
 */
package gourmet