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
	scmd.Usage = func() {
		fmt.Println(` Usage: sudo etchosts <cmd> [options]")
cmd:
	create
	read
	update
	delete
`)
		scmd.PrintDefaults()
	}
}

func main() {
	if len(os.Args) < 2 {
		scmd.Usage()
		return
	}
	ehosts, err := etchosts.New("")
	if err != nil {
		panic(err)
	}
	defer ehosts.Close()
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
			fmt.Println("err creating entry:", err)
			return
		}
		ehosts.Flush()
	case "read":
		if entry, err := ehosts.Read(hostname); err != nil {
			fmt.Println("err reading entry:", err)
		} else {
			fmt.Println(entry)
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
			fmt.Println("err updating entry:", err)
		}
		ehosts.Flush()
	case "delete":
		if err := ehosts.Delete(hostname); err != nil {
			fmt.Println("err deleting entry:", err)
		}
		ehosts.Flush()
	default:
		panic(fmt.Sprintln("invalid command: ", cmd))
	}
}
