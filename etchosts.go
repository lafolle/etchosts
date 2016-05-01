package etchosts

import (
	"bufio"
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	defaultEtcHostsPath = "/etc/hosts"
)

type Entry struct {
	Ipaddr   net.IP
	Hostname string
	Alias    string
}

func (entry Entry) String() string {
	return fmt.Sprintf("%s\t%s\t%s", entry.Ipaddr, entry.Hostname, entry.Alias)
}

type EtcHosts struct {
	Entries *list.List
	sync.RWMutex
	hostsFile     *os.File
	hostsFilePath string
}

func (ehosts *EtcHosts) String() string {
	result := bytes.NewBufferString("")
	for e := ehosts.Entries.Front(); e != nil; e = e.Next() {
		fmt.Fprintln(result, e.Value.(Entry))
	}
	return result.String()
}

func hostsFileWritable(hostFilePath string) {
	return true
}

// If hostsFilePath == "" is true, then default hosts file is used.
func New(hostsFilePath string) (*EtcHosts, error) {
	if len(hostsFilePath) == 0 {
		hostsFilePath = defaultEtcHostsPath
	}
	if !hostsFileWritable(hostsFilePath) {
		return errors.New("user not allowed to edit file")
	}
	f, err := os.OpenFile(hostsFilePath, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	hosts := list.New()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		}
		entry := Entry{}
		splitEntries := strings.Fields(line)
		for i := 0; i < len(splitEntries); i++ {
			field := strings.TrimSpace(splitEntries[i])
			if len(field) == 0 {
				continue
			}
			switch i {
			case 0:
				fmt.Println("ip:", field, len(field))
				entry.Ipaddr = net.ParseIP(field)
				if entry.Ipaddr == nil {
					return nil, errors.New("ip for field is nil")
				}
			case 1:
				entry.Hostname = field
			case 2:
				entry.Alias = field
				break
			}
		}
		hosts.PushBack(entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &EtcHosts{
		Entries:       hosts,
		hostsFile:     f,
		hostsFilePath: hostsFilePath,
	}, nil
}

// As `etchosts` package keeps an open file descriptor for its operations,
// application using this pkg will be responsible for calling close on
// associated file.
func (ehosts *EtcHosts) Close() error {
	return ehosts.hostsFile.Close()
}

// Create creates a new entry in hosts file. In case entry is already, present
// err is returned.
func (ehosts *EtcHosts) Create(entry Entry) error {
	if elm := ehosts.find(entry.Hostname); elm != nil {
		return errors.New(fmt.Sprintf("%s hostname already present", entry.Hostname))
	}
	ehosts.Entries.PushBack(entry)
	return nil
}

// Read returns the first entry with hostname matching hostname argument.
func (ehosts *EtcHosts) Read(hostname string) (Entry, error) {
	if elm := ehosts.find(hostname); elm == nil {
		return Entry{}, errors.New(fmt.Sprintf("%s not found", hostname))
	} else {
		return elm.Value.(Entry), nil
	}
}

// Update updates an entry in hosts file, given hostname. If no entry exists
// for given hostname, error occurs. If multiple entries exists for given hostname,
// first entry is updated.
func (ehosts *EtcHosts) Update(updatedEntry Entry) error {
	if elm := ehosts.find(updatedEntry.Hostname); elm != nil {
		ehosts.Entries.InsertAfter(updatedEntry, elm)
		ehosts.Entries.Remove(elm)
		return nil
	}
	return errors.New(fmt.Sprintf("No entry with %s hostname found", updatedEntry.Hostname))
}

// Delete deletes an entry from hosts file. If no entry exists with given hostname
// error is raised. If multiple entries exists for given hostname, first one is deleted.
func (ehosts *EtcHosts) Delete(hostname string) error {
	if elm := ehosts.find(hostname); elm != nil {
		ehosts.Entries.Remove(elm)
	}
	return nil
}

// It could have been possible to do CRUD ops on `etcHostsPath` just after op is called.
// But that might lead too many disk write operations. Hence, once user has done with
// modifying etc hosts file, she should call Flush, which will write to file.
func (ehosts *EtcHosts) Flush() error {
	ehosts.Lock()
	defer ehosts.Unlock()

	// reset contents of hostFile
	if err := ehosts.hostsFile.Truncate(0); err != nil {
		return err
	}
	ehosts.hostsFile.Seek(0, os.SEEK_SET)
	for e := ehosts.Entries.Front(); e != nil; e = e.Next() {
		_, err := ehosts.hostsFile.Write([]byte(e.Value.(Entry).String() + "\n"))
		if err != nil {
			panic(err)
		}
	}
	return nil
}

// find searches an entry matching hostname
func (ehosts *EtcHosts) find(hostname string) *list.Element {
	for e := ehosts.Entries.Front(); e != nil; e = e.Next() {
		if e.Value.(Entry).Hostname == hostname {
			return e
		}
	}
	return nil
}

// Ref
// 1. http://www.tldp.org/LDP/solrhe/Securing-Optimizing-Linux-RH-Edition-v1.3/chap9sec95.html
// 2. Based on CRUD operations
