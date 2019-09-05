package main

import (
    "github.com/kvasirlabs/gourmet"
    "fmt"
    "log"
)


func main() {
    opt := &gourmet.SensorOptions{
        InterfaceName: "wlan0",
        InterfaceType: gourmet.AfpacketType,
        IsPromiscuous: true,
    }
    src, err := gourmet.NewSensor(opt)
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
}