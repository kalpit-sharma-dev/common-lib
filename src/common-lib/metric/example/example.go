package main

import (
	"fmt"
	"os"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/metric"
)

func main() {
	err := publish(os.Args[1], os.Args[2], os.Args[3])
	if err != nil {
		fmt.Println(err)
	}
}

func publish(namespace, address, portNumber string) error {
	cfg := metric.New()
	cfg.Namespace = namespace
	cfg.Communication.Address = address
	cfg.Communication.PortNumber = portNumber

	// Publish Counter
	c := metric.CreateCounter("c", "Counter-C", 2000)
	err := metric.Publish(cfg, c)
	if err != nil {
		return err
	}

	// Publish Gauge
	g := metric.CreateGauge("g", "Gauge-G", 2500)
	err = metric.Publish(cfg, g)
	if err != nil {
		return err
	}

	c = metric.CreateCounter("c1", "Counter-C1", 2000)
	g = metric.CreateGauge("g1", "Gauge-G1", 2500)
	e := metric.CreateEvent("E1", "Event-E1")
	err = metric.Publish(cfg, c, g, e)
	if err != nil {
		return err
	}

	// publish event
	event := metric.CreateEvent("E2", "Event-E")
	err = metric.Publish(cfg, event)
	if err != nil {
		return err
	}

	metric.PeriodicPublish(5*time.Second, cfg, func() []metric.Collector {
		c.Inc(1)
		g.Inc(1)
		return []metric.Collector{c, g}
	}, func(err error) {})

	return nil
}
