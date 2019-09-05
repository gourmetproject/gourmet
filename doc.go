/*
Package gourmet is a Go network security monitoring library.

A simple example below shows how to capture packets on a network interface named "wlan0" using
AF_PACKET in promiscuous mode and iteratively consume ten reassembled TCP streams.

  opt := &SensorOptions{
    InterfaceName: "wlan0",
    InterfaceType: AfpacketType,
    IsPromiscuous: true,
  }
  src, err := NewSensor(opt)
  if err != nil {
    log.Fatal(err)
  }
  counter := 0
  for stream := range src.Streams() {
    fmt.Printf("new stream: %s %s %d\n", stream.NetworkFlow().String(), stream.TransportFlow().String(), len(stream.Payload()))
    counter++
    if counter == 10 {
      break
    }
  }

 */
package gourmet