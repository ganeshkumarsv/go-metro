package main

import (
	"testing"
)

const good_cfg = `
init_config:
    snaplen: 512
    idle_ttl: 300
    exp_ttl: 60
    statsd_ip: 127.0.0.1
    statsd_port: 8125
    log_to_file: true
    log_level: debug

config:
- interface: en0
  tags: [mytag]
  ips: []
`

const good_file_cfg = `
init_config:
    snaplen: 512
    idle_ttl: 300
    exp_ttl: 60
    statsd_ip: 127.0.0.1
    statsd_port: 8125
    log_to_file: true
    log_level: debug

config:
- interface: file
  pcap: test_tcp.pcap
  tags: [mytag]
  ips: []
`

const bad_file_cfg = `
init_config:
    snaplen: 512
    idle_ttl: 300
    exp_ttl: 60
    statsd_ip: 127.0.0.1
    statsd_port: 8125
    log_to_file: true
    log_level: debug

config:
- interface: file
  tags: [mytag]
  ips: []
`

const bad_cfg = `
init_config:
    snaplen: 512
    idle_ttl: 300
    exp_ttl: 60
    statsd_ip: 127.0.0.1
    statsd_port: 8125
    log_to_file: true
    log_level: debug

config:
`

const bad_interface_cfg = `
init_config:
    snaplen: 512
    idle_ttl: 300
    exp_ttl: 60
    statsd_ip: 127.0.0.1
    statsd_port: 8125
    log_to_file: true
    log_level: debug

config:
- interface: noifc0
  tags: [mytag]
  ips: []
`

func TestParseConfig(t *testing.T) {
	var cfg RTTConfig
	err := cfg.Parse([]byte(good_cfg))
	if err != nil {
		t.Fatalf("RTTConfig.parse expected == nil, got %q", err)
	}
}

func TestParseBadConfig(t *testing.T) {
	var cfg RTTConfig
	err := cfg.Parse([]byte(bad_cfg))
	if err == nil {
		t.Fatalf("RTTConfig.parse expected error, got %q", err)
	}
}

func TestBadInterfaceSniffer(t *testing.T) {
	var cfg RTTConfig
	err := cfg.Parse([]byte(bad_interface_cfg))
	if err != nil {
		t.Fatalf("RTTConfig.parse expected == %q, got %q", nil, err)
	}

	rttsniffer, err := NewDatadogSniffer(cfg.InitConf, cfg.Configs[0], "tcp")

	//sniff
	err = rttsniffer.Sniff()
	if err == nil {
		t.Fatalf("RTTConfig.parse expected error, but got %q", err)
	}
}

func TestSnifferFromBadFile(t *testing.T) {
	var cfg RTTConfig
	err := cfg.Parse([]byte(bad_file_cfg))
	if err == nil {
		t.Fatalf("RTTConfig.parse expected error, got %v", nil)
	}
}

func TestSnifferFromFile(t *testing.T) {
	var cfg RTTConfig
	err := cfg.Parse([]byte(good_file_cfg))
	if err != nil {
		t.Fatalf("RTTConfig.parse expected == %q, got %q", nil, err)
	}

	rttsniffer, err := NewDatadogSniffer(cfg.InitConf, cfg.Configs[0], "tcp")

	//set artificial host_ip 192.168.1.116 (from pcap)
	rttsniffer.host_ips["192.168.1.116"] = true
	//sniff
	err = rttsniffer.Sniff()
	if err != nil {
		t.Fatalf("Problem running sniffer expected %v, got %v - cfg %v", nil, err, cfg.Configs[0])
	}

	n_flows := 0
	for k := range rttsniffer.flows.FlowMapKeyIterator() {
		n_flows++
		flow, e := rttsniffer.flows.Get(k)
		if flow.Src.String() != "192.168.1.116" {
			t.Fatalf("Bad Source IP in flow.")
		}
		if e && flow.Sampled > 0 {
			t.Fatalf("One way HTTP flow can't be sampled for RTT reliably")
		}
	}

	if n_flows == 0 {
		t.Fatalf("Flow was not detected!")
	}
}
