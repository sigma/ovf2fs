package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/vmware/vmw-guestinfo/rpcvmx"

	"github.com/vmware/vmw-guestinfo/vmcheck"
	ovf "github.com/vmware/vmw-ovflib"
)

type ovfWrapper struct {
	env *ovf.OvfEnvironment
}

func (ovf ovfWrapper) readConfig(key string) (string, error) {
	return ovf.env.Properties["guestinfo."+key], nil
}

func isAvailable() bool {
	res, err := vmcheck.IsVirtualWorld()
	return res && (err == nil)
}

func readConfig(key string) (string, error) {
	data, err := rpcvmx.NewConfig().String(key, "")
	if err == nil {
		log.Printf("Read from %q: %q\n", key, data)
	} else {
		log.Printf("Failed to read from %q: %v\n", key, err)
	}
	return data, err
}

func genEnv(env ovf.OvfEnvironment, filepath string) {
	os.MkdirAll(path.Dir(filepath), os.ModePerm)
	f, err := os.Create(filepath)
	if err != nil {
		return
	}
	defer f.Close()

	for k, v := range env.Properties {
		k = strings.ToUpper(strings.Replace(k, ".", "_", -1))
		f.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
}

func main() {
	if !isAvailable() {
		os.Exit(1)
	}

	ovfEnvPtr := flag.String("ovfenv", "/media/ovfenv/ovf-env.xml", "path to ovf environment file")
	envFilePtr := flag.String("env", "/etc/environment.d/ovf", "path to generated env file")

	flag.Parse()

	var ovfEnv []byte

	if _, err := os.Stat(*ovfEnvPtr); os.IsNotExist(err) {
		data, err := readConfig("ovfenv")
		if err != nil {
			ovfEnv = make([]byte, 0)
		} else {
			ovfEnv = []byte(data)
		}
	} else {
		ovfEnv, err = ioutil.ReadFile(*ovfEnvPtr)
		if err != nil {
			ovfEnv = make([]byte, 0)
		}
	}

	env, err := ovf.ReadEnvironment(ovfEnv)
	if err != nil {
		os.Exit(1)
	}

	genEnv(env, *envFilePtr)
}
