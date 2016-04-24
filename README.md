# etchosts (work in progress)

Package, with command, to do CRUD operations on entries in `/etc/hosts` file.  

## Entry  
```
type Entry struct {
    Ipaddr   net.IP
    Hostname string
    Alias    string
}
```

##Package  

##Command
1. Create entry  
    `etchosts create -h beginAgain -ip 192.168.1.3 -a scarletJohnson`
2. Read entry  
    `etchosts read -h beginAgain`
3. Update entry  
    `etchosts create -h beginAgain -ip 192.168.1.2 -a scarletJohnson`
4. Delete entry  
    `etchosts delete -h beginAgain`
