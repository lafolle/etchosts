package etchosts

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"sync"
	"testing"
)

func generateFixture() (string, *os.File, *EtcHosts) {
	// generate file obj
	f, err := ioutil.TempFile("/tmp", "etc_hosts")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	content := `
127.0.0.1 localhost localhost
53.53.68.8 pub sub
`
	if err := ioutil.WriteFile(f.Name(), []byte(content), os.ModeTemporary); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// generate corresponding etchosts obj
	tlist := list.New()
	tlist.PushBack(Entry{
		Ipaddr:   net.ParseIP("127.0.0.1"),
		Hostname: "localhost",
		Alias:    "localhost",
	})
	tlist.PushBack(Entry{
		Ipaddr:   net.ParseIP("53.53.68.8"),
		Hostname: "pub",
		Alias:    "sub",
	})
	ehosts := EtcHosts{
		Entries:   tlist,
		hostsFile: f,
		lock:      nil,
	}

	// generate filepath
	fi, err := f.Stat()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	filepath := "/tmp/" + fi.Name()

	return filepath, f, &ehosts
}

func TestEntryString(t *testing.T) {
	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rIpaddr   net.IP
		rHostname string
		rAlias    string
		// Expected results.
		want string
	}{
		{
			name:      "string formatting",
			rIpaddr:   net.ParseIP("127.0.0.1"),
			rHostname: "localhost",
			rAlias:    "localhost",
			want:      fmt.Sprintf("%s\t%s\t%s\n", "127.0.0.1", "localhost", "localhost"),
		},
	}
	for _, tt := range tests {
		entry := &Entry{
			Ipaddr:   tt.rIpaddr,
			Hostname: tt.rHostname,
			Alias:    tt.rAlias,
		}
		if got := entry.String(); got != tt.want {
			t.Errorf("%q. Entry.String() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestEtcHostsString(t *testing.T) {
	tlist := list.New()
	tlist.PushBack(Entry{
		Ipaddr:   net.ParseIP("127.0.0.1"),
		Hostname: "localhost",
		Alias:    "localhost",
	})
	tlist.PushBack(Entry{
		Ipaddr:   net.ParseIP("89.90.90.2"),
		Hostname: "bloop",
		Alias:    "dup",
	})

	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rEntries   *list.List
		rhostsFile *os.File
		rlock      *sync.RWMutex
		// Expected results.
		want string
	}{
		{
			name:       "description of EtcHosts",
			rEntries:   tlist,
			rhostsFile: nil,
			rlock:      nil,
			want:       fmt.Sprintf("127.0.0.1\tlocalhost\tlocalhost\n89.90.90.2\tbloop\tdup\n"),
		},
	}
	for _, tt := range tests {
		ehosts := &EtcHosts{
			Entries:   tt.rEntries,
			hostsFile: tt.rhostsFile,
			lock:      tt.rlock,
		}
		if got := ehosts.String(); got != tt.want {
			t.Errorf("%q. EtcHosts.String() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestNew(t *testing.T) {
	filepath, tmpHostsFile, ehosts := generateFixture()
	tests := []struct {
		// Test description.
		name string
		// Parameters.
		hostsFilePath string
		// Expected results.
		want    *EtcHosts
		wantErr bool
	}{
		{
			name:          "new etc hosts",
			hostsFilePath: filepath,
			want:          ehosts,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		got, err := New(tt.hostsFilePath)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. New() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got.Entries, tt.want.Entries) {
			t.Errorf("%q. New() = %v, want %v", tt.name, got, tt.want)
		}
	}
	tmpHostsFile.Close()
}

func TestEtcHostsCreate(t *testing.T) {
	_, tmpHostsFile, ehosts := generateFixture()
	testEntry := Entry{
		Ipaddr:   net.ParseIP("70.70.70.70"),
		Hostname: "hallelujah",
		Alias:    "jango",
	}
	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rEntries   *list.List
		rhostsFile *os.File
		rlock      *sync.RWMutex
		// Parameters.
		entry Entry
		// Expected results.
		wantErr bool
	}{
		{
			name:       "create entry",
			rEntries:   ehosts.Entries,
			rhostsFile: tmpHostsFile,
			rlock:      nil,
			entry:      testEntry,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		ehosts := &EtcHosts{
			Entries:   tt.rEntries,
			hostsFile: tt.rhostsFile,
			lock:      tt.rlock,
		}
		if err := ehosts.Create(tt.entry); (err != nil) != tt.wantErr {
			t.Errorf("%q. EtcHosts.Create() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
		if !reflect.DeepEqual(ehosts.Entries.Back().Value.(Entry), testEntry) {
			t.Errorf("%q. EtcHosts.Create() entry = %v, want %v", ehosts.Entries.Back().Value.(Entry), testEntry)
		}
	}
	tmpHostsFile.Close()
}

func TestEtcHostsRead(t *testing.T) {
	_, tmpHostsFile, ehosts := generateFixture()
	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rEntries   *list.List
		rhostsFile *os.File
		rlock      *sync.RWMutex
		// Parameters.
		hostname string
		// Expected results.
		want  Entry
		want1 bool
	}{
		{
			name:       "read hostname",
			rEntries:   ehosts.Entries,
			rhostsFile: tmpHostsFile,
			rlock:      nil,
			hostname:   "pub",
			want: Entry{
				Ipaddr:   net.ParseIP("53.53.68.8"),
				Hostname: "pub",
				Alias:    "sub",
			},
			want1: true,
		},
	}
	for _, tt := range tests {
		ehosts := &EtcHosts{
			Entries:   tt.rEntries,
			hostsFile: tt.rhostsFile,
			lock:      tt.rlock,
		}
		got, got1 := ehosts.Read(tt.hostname)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. EtcHosts.Read() got = %v, want %v", tt.name, got, tt.want)
		}
		if got1 != tt.want1 {
			t.Errorf("%q. EtcHosts.Read() got1 = %v, want %v", tt.name, got1, tt.want1)
		}
	}
	tmpHostsFile.Close()
}

func TestEtcHostsUpdate(t *testing.T) {
	_, tmpHostsFile, ehosts := generateFixture()
	testEntry := Entry{
		Ipaddr:   net.ParseIP("7.7.7.7"),
		Hostname: "localhost",
		Alias:    "dubmash",
	}
	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rEntries   *list.List
		rhostsFile *os.File
		rlock      *sync.RWMutex
		// Parameters.
		updatedEntry Entry
		// Expected results.
		wantErr bool
	}{
		{
			name:         "update entry",
			rEntries:     ehosts.Entries,
			rhostsFile:   tmpHostsFile,
			rlock:        nil,
			updatedEntry: testEntry,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		ehosts := &EtcHosts{
			Entries:   tt.rEntries,
			hostsFile: tt.rhostsFile,
			lock:      tt.rlock,
		}
		if err := ehosts.Update(tt.updatedEntry); (err != nil) != tt.wantErr {
			t.Errorf("%q. EtcHosts.Update() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
		if ehosts.Entries.Front().Value.(Entry).Ipaddr.String() != net.ParseIP("7.7.7.7").String() {
			t.Errorf("EtcHosts.Update() gotEntry = %v, wantEntry %v", ehosts.Entries.Front().Value.(Entry))
		}
	}
	tmpHostsFile.Close()
}

func TestEtcHostsDelete(t *testing.T) {
	_, tmpHostsFile, ehosts := generateFixture()
	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rEntries   *list.List
		rhostsFile *os.File
		rlock      *sync.RWMutex
		// Parameters.
		hostname string
		// Expected results.
		wantErr bool
	}{
		{
			name:       "delete entry",
			rEntries:   ehosts.Entries,
			rhostsFile: tmpHostsFile,
			rlock:      nil,
			hostname:   "pub",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		ehosts := &EtcHosts{
			Entries:   tt.rEntries,
			hostsFile: tt.rhostsFile,
			lock:      tt.rlock,
		}
		if err := ehosts.Delete(tt.hostname); (err != nil) != tt.wantErr {
			t.Errorf("%q. EtcHosts.Delete() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
		if ehosts.Entries.Back().Value.(Entry).Hostname == "pub" {
			t.Errorf("wrong value")
		}
	}
	tmpHostsFile.Close()
}

func TestEtcHostsFlush(t *testing.T) {
	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rEntries   *list.List
		rhostsFile *os.File
		rlock      *sync.RWMutex
		// Expected results.
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		ehosts := &EtcHosts{
			Entries:   tt.rEntries,
			hostsFile: tt.rhostsFile,
			lock:      tt.rlock,
		}
		if err := ehosts.Flush(); (err != nil) != tt.wantErr {
			t.Errorf("%q. EtcHosts.Flush() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestEtcHostsFind(t *testing.T) {
	_, tmpHostsFile, ehosts := generateFixture()
	tests := []struct {
		// Test description.
		name string
		// Receiver fields.
		rEntries   *list.List
		rhostsFile *os.File
		rlock      *sync.RWMutex
		// Parameters.
		hostname string
		// Expected results.
		want *list.Element
	}{
		{
			name:       "find entry",
			rEntries:   ehosts.Entries,
			rhostsFile: tmpHostsFile,
			rlock:      nil,
			hostname:   "pub",
			want:       ehosts.Entries.Back(),
		},
	}
	for _, tt := range tests {
		ehosts := &EtcHosts{
			Entries:   tt.rEntries,
			hostsFile: tt.rhostsFile,
			lock:      tt.rlock,
		}
		if got := ehosts.find(tt.hostname); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. EtcHosts.find() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
