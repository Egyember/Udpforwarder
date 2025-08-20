package main

import (
	"errors"
	"fmt"
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
}

type rawConfig struct {
	Rules []rawRule `toml:"rule"`
	Log   bool      `toml:"log"`
}

type rule struct {
	Lisen   *net.UDPAddr
	Forward []*net.UDPAddr
}
type config struct {
	rules []rule
	log   bool
}

func (s config) String() string{
	st := fmt.Sprintf("logging: %v\n", s.log)
	for _, v:= range s.rules{
		st += fmt.Sprintln("lisen:", *v.Lisen)
		for _, w:= range v.Forward{
		st += fmt.Sprintln("\tforward:", *w)
		}
	}
	return st
}

func (s rawConfig) parese() (config, error) {
	var c config
	c.log = s.Log
	c.rules = make([]rule, len(s.Rules))
	for k, v := range s.Rules {
		l, err := v.Lisen.parese()
		if err != nil {
			panic(err)
		}
		c.rules[k].Lisen = &l
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

func lisendAndForward(mtu int, laddr *net.UDPAddr, forward []*net.UDPAddr, id int, report chan int, log bool) {
	defer func(id int, report chan int) {
		if r := recover(); r != nil {
			fmt.Println("lisener failed with error:", r)
			report <- id
		}
	}(id, report)
	con, err := net.ListenUDP("udp", laddr)
	if err != nil {
		panic(err)
	}
	defer con.Close()
	// con.SetDeadline(time.Unix(0, 0)) // disable timeout
	buffer := make([]byte, mtu)
	if log {
		fmt.Println("staring lissener")
	}
	for {
		n, addr, err := con.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}
		if log {
			fmt.Println("got packet")
		}
		for _, v := range forward {
			f, err := net.DialUDP("udp", addr, v)
			if err != nil {
				panic(err)
			}
			f.Write(buffer[:n])
			f.Close()
		}

	}
}

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
	bconf, err := os.ReadFile("config.toml")
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
	crashed := make(chan int)
	fmt.Println(conf)
	for k, v := range conf.rules {
		go lisendAndForward(maxmtu, v.Lisen, v.Forward, k, crashed, conf.log)
	}
	for {
		c := <-crashed
		fmt.Println(c, "th rule crashed restarting...")
		go lisendAndForward(maxmtu, conf.rules[c].Lisen, conf.rules[c].Forward, c, crashed, conf.log)

	}
}
