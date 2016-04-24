/*
	etchosts create -hostname <hostname> -ipaddr <ipaddr> -alias <alias>
	etchosts read -hostname <hostname>
	etchosts update -hostname -ipaddr <ipaddr> -alias <alias>
	etchosts delete -hostname <hostname>
*/
package main

import (
	"flag"
	"fmt"
	"github.com/lafolle/etchosts"
	"net"
	"os"
)

var (
	hostname string
	ipaddr   string
	alias    string

	scmd *flag.FlagSet
)

func init() {
	scmd = flag.NewFlagSet("cmd", flag.ExitOnError)
	scmd.StringVar(&hostname, "hostname", "", "hostname of machine")
	scmd.StringVar(&ipaddr, "ip", "", "ipaddr of machine")
	scmd.StringVar(&alias, "alias", "", "alias of machine")
}

func main() {
	ehosts, err := etchosts.New("")
	if err != nil {
		panic(err)
	}
	scmd.Parse(os.Args[2:])
	if !scmd.Parsed() {
		panic("failed to parse flags")
	}
	if len(hostname) == 0 {
		panic("hostname cannot be empty")
	}
	cmd := os.Args[1]
	switch cmd {
	case "create":
		if len(ipaddr) == 0 {
			panic(ipaddr + " cannot be empty")
		}
		ip := net.ParseIP(ipaddr)
		if ip == nil {
			panic("create: failed to parse ip")
		}
		if err := ehosts.Create(etchosts.Entry{
			Hostname: hostname,
			Ipaddr:   ip,
			Alias:    alias,
		}); err != nil {
			panic(err)
		}
		ehosts.Flush()
	case "read":
		if entry, err := ehosts.Read(hostname); err != nil {
			fmt.Println(entry)
		} else {
			fmt.Println("No entry present for ", hostname)
		}
	case "update":
		if len(ipaddr) == 0 {
			panic("ipaddr cannot be empty")
		}
		ip := net.ParseIP(ipaddr)
		if ip == nil {
			panic("create: failed to parse ip")
		}
		if err := ehosts.Update(etchosts.Entry{
			Hostname: hostname,
			Ipaddr:   ip,
			Alias:    alias,
		}); err != nil {
			panic(err)
		}
		ehosts.Flush()
	case "delete":
		if err := ehosts.Delete(hostname); err != nil {
			panic(err)
		}
		ehosts.Flush()
	default:
		panic("invalid command")
	}
}
