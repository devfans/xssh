package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
  "log"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"github.com/devfans/envconf"
)

// `host, hostname, user, bastion, category`
var (
	// etcd store
	pRoot     *string = flag.String("root", "/ssh", "etcd root path")
	pEndpoint *string = flag.String("url", "", "store endpoint")

	// host operations
	pAdd      *bool   = flag.Bool("add", false, "add host into store")
	pGet      *bool   = flag.Bool("get", false, "get host details from store")
	pList     *bool   = flag.Bool("list", false, "list hosts from store")
	pChange   *bool   = flag.Bool("change", false, "change host parameters")
	pDelete   *bool   = flag.Bool("del", false, "delete host from store")
	pAwsInput *bool   = flag.Bool("awsInput", false, "import from aws inventory file")
	pSave     *bool   = flag.Bool("save", false, "save ssh config file to disk as ~/.ssh/config")
	pStore    *bool   = flag.Bool("s", false, "save store configuration as ~/.ssh/store")
	pFile     *string = flag.String("file", "", "file path")

	// key operations
	pGetKey   *bool   = flag.Bool("getKey", false, "get key directory from store")
	pListKeys *bool   = flag.Bool("listKeys", false, "list keys from store")
	pDelKey   *bool   = flag.Bool("delKey", false, "delete key from store")
	pPutKey   *bool   = flag.Bool("putKey", false, "put key into store")
	pKey      *string = flag.String("key", "", "dir key in store")
	pValue    *string = flag.String("value", "", "specifiy key's value")
	pAllKeys  *bool   = flag.Bool("allKeys", false, "get all keys from store recursively")

	// host args
	pHost     *string = flag.String("host", "", "host: Host in ssh/config")
	pRename   *string = flag.String("rename", "", "new host name for changing host parameters")
	pHostname *string = flag.String("ip", "", "ip or domain name: HostName in ssh/config")
	pPort     *string = flag.String("port", "22", "port: Port in ssh/config")
	pUser     *string = flag.String("user", "ubuntu", "username: User in ssh/config")
	pBastion  *string = flag.String("bastion", "", "bastion host: used for ProxyComand in ssh/config")
	pCategory *string = flag.String("category", "default", "category")
	pPem      *string = flag.String("pem", "", "private key file name, default will be user.pem under ~/.ssh")
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkFatal(err error, tpl string, args ...interface{}) {
	if err != nil {
		fmt.Printf(tpl, args...)
		log.Fatal(err)
	}
}

// Store interface for key operations
type Store interface {
	init()
	GetKey(key string) (string, error)
	DelKey(key string) error
	ListKeys(recursive bool)
	AllKeys()
	PutKey(key, value string) error
	CollectKeys() [][]byte
}

// etcd store implementation
type EtcdStore struct {
	client client.KeysAPI
}

// main config object
type Config struct {
	store Store
}

// host model
type Host struct {
	Host     string
	Ip       string
	Port     string
	Bastion  string
	User     string
	Category string
	Dir      string
}

func (c *Config) init() {
	// currently only support etcd store
	c.store = &EtcdStore{}
	c.store.init()
}

// create main config
func NewConfig() *Config {
	config := &Config{}
	config.init()
	return config
}

func (es *EtcdStore) init() {
	fmt.Printf("Using etcd store:%s\n", *pEndpoint)
	cfg := client.Config{
		Endpoints:               []string{*pEndpoint},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	checkError(err)
	es.client = client.NewKeysAPI(c)
}

// save host into store
func (h *Host) Save(store Store) bool {
	h.GetDir()
	jsonData, err := json.Marshal(h)
	checkError(err)
	err = store.PutKey(h.Dir, string(jsonData))
	checkError(err)
	return true
}

// delete host from store
func (h *Host) Del(store Store) {
	h.GetDir()
	err := store.DelKey(h.Dir)
	checkError(err)
	fmt.Printf("Deleted %v\n", h.Dir)
}

// dump host config to string
func (h *Host) String() string {
	if len(*pPem) == 0 {
		*pPem = "~/.ssh/" + os.Getenv("USER") + ".pem"
	}

	parts := []string{"Host " + h.Host,
		"HostName " + h.Ip, "Port " + h.Port,
		"User " + h.User, "IdentityFile " + *pPem}

	if len(h.Bastion) > 0 {
		parts = append(parts, "ProxyCommand ssh "+h.Bastion+
			" -W "+"%%"+"h:"+"%%"+"p")
	}

	return fmt.Sprintf(strings.Join(parts, "\n    "))
}

// compose host key path
func (h *Host) GetDir() {
	h.Dir = *pRoot + "/" + h.Category + "/" + h.Host
}

// get host from store
func (h *Host) Get(store Store) {
	h.GetDir()
	resp, err := store.GetKey(h.Dir)
	checkError(err)
	fmt.Printf("Fetching...\n%v", resp)
	err = json.Unmarshal([]byte(resp), h)
	checkError(err)
}

// get key from etcd store
func (es *EtcdStore) GetKey(key string) (string, error) {
	resp, err := es.client.Get(context.Background(), key, nil)
	if err != nil {
		return "", err
	} else {
		return fmt.Sprint(resp.Node.Value), nil
	}
}

// get all keys from etcd store
func (es *EtcdStore) AllKeys() {
	resp, err := es.client.Get(context.Background(), *pKey,
		&client.GetOptions{Recursive: true})
	checkError(err)

	jsonData, err := json.Marshal(resp)
	checkError(err)
	fmt.Println(jsonData)
}

// list keys from etcd store
func (es *EtcdStore) ListKeys(recursive bool) {
	resp, err := es.client.Get(context.Background(), *pKey,
		&client.GetOptions{Recursive: recursive})
	checkError(err)
	jsonData, err := json.Marshal(resp)
	checkError(err)
	fmt.Println(jsonData)
}

// put key into etcd store
func (es *EtcdStore) PutKey(key, value string) error {
	_, err := es.client.Set(context.Background(), key, value, nil)
	if err == nil {
		fmt.Printf("Set is done for key: %v with value %v\n", key, value)
	}
	return err
}

// delete key from etcd store
func (es *EtcdStore) DelKey(key string) error {
	_, err := es.client.Delete(context.Background(), key, nil)
	return err
}

// collect host keys from store
func (es *EtcdStore) CollectKeys() [][]byte {
	keys := make([][]byte, 0)
	*pKey = "/ssh"
	resp, err := es.client.Get(context.Background(), *pKey,
		&client.GetOptions{Recursive: true})

	checkError(err)
	for _, node := range resp.Node.Nodes {
		for _, cnode := range node.Nodes {
			keys = append(keys, []byte(cnode.Value))
		}
	}
	return keys
}

// collect hosts from store
func (c *Config) CollectHosts() []Host {
	var hosts []Host
	keys := c.store.CollectKeys()
	for _, key := range keys {
		host := Host{}
		err := json.Unmarshal(key, &host)
		checkError(err)
		hosts = append(hosts, host)
	}
	return hosts
}

// list all hosts
func (c *Config) ListHosts() {
	hosts := c.CollectHosts()
	for _, h := range hosts {
		fmt.Println(h.String() + "\n")
	}
}

// save as file: ~/.ssh/config
func (c *Config) Save() {
	hosts := c.CollectHosts()
	configDir, err := filepath.Abs(os.Getenv("HOME") + "/.ssh/config")
	checkError(err)

  // make sure dir exists
	dir := filepath.Dir(configDir)
	if _, err = os.Stat(dir); err != nil {
		os.MkdirAll(dir, os.ModePerm)
	}

	// backup old dir
	if _, err = os.Stat(configDir); err == nil {
		err = os.Rename(configDir, configDir+string(time.Now().Format(time.RFC3339)))
		if err != nil {
			log.Fatalf("Failed to rename old config file %v.\n", configDir)
		}
	}

	f, err := os.OpenFile(configDir, os.O_RDWR|os.O_CREATE, 0755)
	checkError(err)
	defer f.Close()

	for _, h := range hosts {
		fmt.Printf("Writing host %v\n", h.Host)
		_, err = f.WriteString(h.String() + "\n")
		checkError(err)
	}

	fmt.Printf("Config file save as %v\n", configDir)
}

// import host from aws inventory file
func (c *Config) AwsImport() {
	if len(*pFile) == 0 {
		log.Fatalln("Please specify the file!")
	}
	fileDir, err := filepath.Abs(*pFile)
	checkError(err)

	fmt.Printf("Loading from: %v\n", fileDir)

	f, err := os.Open(fileDir)
	checkError(err)
	defer f.Close()
	var j interface{}
	jsonParser := json.NewDecoder(f)
	err = jsonParser.Decode(&j)
	checkError(err)
	d := j.(map[string]interface{})
	meta := d["_meta"].(map[string]interface{})
	items := meta["hostvars"].(map[string]interface{})
	for k, v := range items {
		_ = v.(map[string]interface{})
		fmt.Printf("Adding %v\n", k)
		h := Host{Host: k, Ip: k, User: *pUser, Port: *pPort, Bastion: *pBastion, Category: *pCategory}
		h.GetDir()
		jsonData, err := json.Marshal(h)
		checkError(err)
		err = c.store.PutKey(h.Dir, string(jsonData))
		checkError(err)
	}
}

// add host into store
func (c *Config) AddHost() {
	if len(*pHost) == 0 {
		log.Fatalln("HostName is missing!")
	}

	h := Host{Host: *pHost, Ip: *pHostname, Port: *pPort,
		User: *pUser, Bastion: *pBastion, Category: *pCategory}

	if h.Save(c.store) {
		fmt.Printf("Host %v added\n", *pHost)
	} else {
		fmt.Printf("Failed to add host %v\n", *pHost)
	}
}

// get host from store
func (c *Config) GetHost() {
	if len(*pHost) == 0 {
		log.Fatalln("HostName is missing!")
	}
	h := Host{Host: *pHost, Category: *pCategory}
	h.Get(c.store)
	fmt.Println(h.String() + "\n")
}

// change host config
func (c *Config) ChangeHost() {
	if len(*pHost) == 0 {
		log.Fatalln("HostName is missing!")
	}
	h := Host{Host: *pHost, Category: *pCategory}
	h.Get(c.store)

	if len(*pRename) != 0 {
		h.Host = *pRename
	}

	if len(*pUser) != 0 {
		h.User = *pUser
	}

	if len(*pHostname) != 0 {
		h.Ip = *pHostname
	}

	if *pPort != "22" {
		h.Port = *pPort
	}

	if len(*pBastion) != 0 {
		h.Bastion = *pBastion
	} else {
		h.Bastion = ""
	}

	if h.Save(c.store) {
		h.Host = *pHost
		if len(*pRename) != 0 {
			h.Del(c.store)
			fmt.Printf("Host name changed from %v to %v\n", *pHost, *pRename)
		} else {
			fmt.Printf("Host %s is modified\n", *pHost)
		}
		h.Get(c.store)
		fmt.Println(h.String() + "\n")
	}
}

// delete host
func (c *Config) DeleteHost() {
	if len(*pHost) == 0 {
		log.Fatalln("HostName is missing!")
	}
	h := Host{Host: *pHost, Category: *pCategory}
	h.Del(c.store)
	fmt.Printf("Deleted %v\n", h.Dir)
}

func checkKey() {
	if len(*pKey) == 0 {
		log.Fatalln(fmt.Sprintf("Invalid key %v", *pKey))
	}
}

func main() {
	flag.Parse()
	// get url from ~/.ssh/store or env: XSSH_STORE_URL
	conf := envconf.NewConfig("~/.ssh/store")
	conf.Put("store", "etcd")
	conf.Section = "etcd"
	if *pEndpoint == "" {
		*pEndpoint = conf.Get("XSSH_ETCD_URL", "url")
	}

	if *pEndpoint == "" {
		*pEndpoint = "http://localhost:2379"
	}

	conf.Put("url", *pEndpoint)
	config := NewConfig()

	if *pStore {
		conf.Save()
	} else if *pAdd {
		config.AddHost()
	} else if *pDelete {
		config.DeleteHost()
	} else if *pGet {
		config.GetHost()
	} else if *pSave {
		config.Save()
	} else if *pChange {
		config.ChangeHost()
	} else if *pList {
		config.ListHosts()
	} else if *pAllKeys {
		config.store.AllKeys()
	} else if *pGetKey {
		checkKey()
		v, err := config.store.GetKey(*pKey)
		checkError(err)
		fmt.Printf("Value is: %v\n", v)
	} else if *pPutKey {
		checkKey()
		err := config.store.PutKey(*pKey, *pValue)
		checkError(err)
		fmt.Printf("Key %v is now set as %v\n", *pKey, *pValue)
	} else if *pListKeys {
		checkKey()
		config.store.ListKeys(false)
	} else if *pDelKey {
		checkKey()
		err := config.store.DelKey(*pKey)
		checkError(err)
		fmt.Printf("Key %v is removed\n", *pKey)
	} else if *pAwsInput {
		config.AwsImport()
	} else {
		log.Fatalln("Invalid parameters set!")
	}
}
