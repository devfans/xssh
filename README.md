# xssh

[![Build Status](https://travis-ci.org/devfans/xssh.svg?branch=master)](https://travis-ci.org/devfans/xssh)
[![Go Report Card](https://goreportcard.com/badge/github.com/devfans/xssh)](https://goreportcard.com/report/github.com/devfans/xssh)

Small tool for sharing ssh config with your team members. Using etcd as the store, will support other stores, redis etc...

# Get Started
xssh command line help:
```
  -add
    	add host into store
  -allKeys
    	get all keys from store recursively
  -awsInput
    	import from aws inventory file
  -bastion string
    	bastion host: used for ProxyComand in ssh/config
  -category string
    	category (default "default")
  -change
    	change host parameters
  -del
    	delete host from store
  -delKey
    	delete key from store
  -file string
    	file path
  -get
    	get host details from store
  -getKey
    	get key directory from store
  -host string
    	host: Host in ssh/config
  -ip string
    	ip or domain name: HostName in ssh/config
  -key string
    	dir key in store
  -list
    	list hosts from store
  -listKeys
    	list keys from store
  -pem string
    	private key file name, default will be user.pem under ~/.ssh
  -port string
    	port: Port in ssh/config (default "22")
  -putKey
    	put key into store
  -rename string
    	new host name for changing host parameters
  -root string
    	etcd root path (default "/ssh")
  -s	save store configuration as ~/.ssh/store
  -save
    	save ssh config file to disk as ~/.ssh/config
  -url string
    	store endpoint
  -user string
    	username: User in ssh/config (default "ubuntu")
  -value string
    	specifiy key's value
```

Examples:

+ List all hosts in store
```
xssh -list
```

+ Add host
```
xssh -add -host app01 -ip 10.0.0.10 -user ubuntu -bastion bastion-server
```

+ Delete a host

+ Save as local ssh config file (Automatically backup your old config files)
```
xssh -save
```

+ Save store configurations to ~/.ssh/store
```
xssh -s
```


