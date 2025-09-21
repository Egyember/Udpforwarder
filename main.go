package main

import (
	"errors"
	"fmt"
	"log"
	"log/syslog"
	"net"
	"os"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type textaddr struct {
	Ip   string `toml:"ip"`
	Port int    `toml:"port"`
}

func (s textaddr) parese() (net.UDPAddr, error) {
	var r net.UDPAddr
	r.IP = net.ParseIP(s.Ip)
	r.Port = int(s.Port)
	return r, nil
}

type rawRule struct {
	Lisen   textaddr   `toml:"lisen"`
	Forward []textaddr `toml:"forward"`
	UseSrc  bool       `toml:"use-src"`
}

type rawConfig struct {
	Rules   []rawRule `toml:"rule"`
	Log     bool      `toml:"syslog"`
	Logaddr string    `toml:"logaddr"`
}

type rule struct {
	Lisen   *net.UDPAddr
	Forward []*net.UDPAddr
	UseSrc  bool
}
type config struct {
	rules   []rule
	syslog  bool
	logaddr string
}

func (s config) String() string {
	st := fmt.Sprintf("logging: %v\n", s.syslog)
	for _, v := range s.rules {
		st += fmt.Sprintln("lisen:", *v.Lisen)
		st += fmt.Sprintln("\tuse-src:", v.UseSrc)
		for _, w := range v.Forward {
			st += fmt.Sprintln("\tforward:", *w)
		}
	}
	return st
}

func (s rawConfig) parese() (config, error) {
	var c config
	c.syslog = s.Log
	c.logaddr = s.Logaddr
	c.rules = make([]rule, len(s.Rules))
	for k, v := range s.Rules {
		l, err := v.Lisen.parese()
		if err != nil {
			panic(err)
		}
		c.rules[k].Lisen = &l
		c.rules[k].UseSrc = v.UseSrc
		c.rules[k].Forward = make([]*net.UDPAddr, len(v.Forward))
		for j, w := range v.Forward {
			l, err := w.parese()
			if err != nil {
				panic(err)
			}
			c.rules[k].Forward[j] = &l
		}
	}
	return c, nil
}

func lisendAndForward(mtu int, rule rule, id int, report chan int, logger *log.Logger) {
	defer func(id int, report chan int, logger *log.Logger) {
		if r := recover(); r != nil {
			logger.Println("lisener failed with error:", r)
			report <- id
		}
	}(id, report, logger)
	con, err := net.ListenUDP("udp", rule.Lisen)
	if err != nil {
		panic(err)
	}
	defer con.Close()
	// con.SetDeadline(time.Unix(0, 0)) // disable timeout
	buffer := make([]byte, mtu)
	logger.Println("staring lissener")
	for {
		n, addr, err := con.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}
		logger.Println("got packet from ", *addr)
		if !rule.UseSrc {
			addr.IP = net.IP{}
		}
		for _, v := range rule.Forward {
			f, err := net.DialUDP("udp", addr, v)
			if err != nil {
				panic(err)
			}
			f.Write(buffer[:n])
			f.Close()
		}

	}
}

var configfile = "config.toml"

func main() {
	inf, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	mtus := make([]int, len(inf))
	for k, v := range inf {
		mtus[k] = v.MTU
	}
	maxmtu := slices.Max(mtus)
	bconf, err := os.ReadFile(configfile)
	if err != nil {
		panic(err)
	}
	var rawConf rawConfig
	r := strings.NewReader(string(bconf))
	d := toml.NewDecoder(r)
	d.DisallowUnknownFields()
	err = d.Decode(&rawConf)
	if err != nil {
		panic(err)
	}

	if len(rawConf.Rules) <= 0 {
		panic(errors.New("not enugh rules"))
	}
	conf, err := rawConf.parese()
	if err != nil {
		panic(err)
	}
	var logger *log.Logger
	if conf.syslog {
		sysLog, err := syslog.Dial("tcp", conf.logaddr,
			syslog.LOG_WARNING|syslog.LOG_DAEMON, "udp_forward")
		if err != nil {
			panic(err)
		}
		logger = log.New(sysLog, "", log.Ldate)

	} else {
		logger = log.Default()
	}
	crashed := make(chan int)
	fmt.Println(conf)
	for k, v := range conf.rules {
		go lisendAndForward(maxmtu, v, k, crashed, logger)
	}
	for {
		c := <-crashed
		logger.Println(c, "th rule crashed restarting...")
		go lisendAndForward(maxmtu, conf.rules[c], c, crashed, logger)

	}
}
