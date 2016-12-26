package pwd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	SessionId    string
	Hostname     string
	SSLPort      string
	Port         string
	URL          string
	Created      bool
	InstanceName string
}

type instance struct {
	Name string
	IP   string
}

var notImplemented error = errors.New("Not implemented")

func dump(infos ...interface{}) {
	spew.Fdump(os.Stderr, infos)
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "pwd-session-id",
			Usage:  "PWD session id to create the instance",
			EnvVar: "PWD_SESSION_ID",
		},
		mcnflag.StringFlag{
			Name:   "pwd-hostname",
			Usage:  "PWD hostname create machines from",
			EnvVar: "PWD_HOSTNAME",
			Value:  "play-with-docker.com",
		},
		mcnflag.StringFlag{
			Name:   "pwd-ssl-port",
			Usage:  "pwd ssl port to connect to the daemon",
			EnvVar: "PWD_SSL_PORT",
			Value:  "443",
		},
		mcnflag.StringFlag{
			Name:   "pwd-port",
			Usage:  "pwd port to connect to the API",
			EnvVar: "PWD_PORT",
			Value:  "80",
		},
	}
}

func (d *Driver) Create() error {
	resp, err := http.Post(fmt.Sprintf("http://%s:%s/sessions/%s/instances", d.Hostname, d.Port, d.SessionId), "", nil)

	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Could not create instance %v %v", err, resp)
	}

	defer resp.Body.Close()

	i := &instance{}

	json.NewDecoder(resp.Body).Decode(i)

	err = setupCerts(d, *i)
	if err != nil {
		return fmt.Errorf("Error configuring PWD certs: %v ", err)
	}

	d.IPAddress = i.IP
	d.InstanceName = i.Name
	d.URL = fmt.Sprintf("tcp://ip%s-2375.%s:%s", strings.Replace(d.IPAddress, ".", "_", -1), d.Hostname, d.SSLPort)
	d.Created = true
	return nil

}

func setupCerts(d *Driver, i instance) error {
	hosts := append([]string{}, i.IP, i.Name, "localhost")
	bits := 2048
	machineName := d.GetMachineName()
	org := mcnutils.GetUsername() + "." + machineName
	caPath := filepath.Join(d.StorePath, "certs", "ca.pem")
	caKeyPath := filepath.Join(d.StorePath, "certs", "ca-key.pem")

	serverCertPath := d.ResolveStorePath("server.pem")
	serverKeyPath := d.ResolveStorePath("server-key.pem")

	err := cert.GenerateCert(&cert.Options{
		Hosts:       hosts,
		CertFile:    serverCertPath,
		KeyFile:     serverKeyPath,
		CAFile:      caPath,
		CAKeyFile:   caKeyPath,
		Org:         org,
		Bits:        bits,
		SwarmMaster: false,
	})

	if err != nil {
		return fmt.Errorf("error generating server cert: %s", err)
	}

	type certs struct {
		ServerCert []byte `json:"server_cert"`
		ServerKey  []byte `json:"server_key"`
	}

	c := certs{}
	serverCert, err := ioutil.ReadFile(serverCertPath)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Error reading file: %s", serverCertPath)
	}
	serverKey, err := ioutil.ReadFile(serverKeyPath)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Error reading file: %s", serverCertPath)
	}
	c.ServerCert = serverCert
	c.ServerKey = serverKey
	b, jsonErr := json.Marshal(c)
	if jsonErr != nil {
		log.Println(jsonErr)
		return errors.New("Error encoding json")
	}

	resp, err := http.Post(fmt.Sprintf("http://%s:%s/sessions/%s/instances/%s/keys", d.Hostname, d.Port, d.SessionId, i.Name), "application/json", bytes.NewReader(b))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println(err, resp)
		return errors.New("Error setting up keys on PWD server")
	}
	defer resp.Body.Close()
	return nil
}

var counter int = 0

func (d *Driver) DriverName() string {
	defer func() { counter++ }()
	// Only after creation and when driver is queried for provisioning return "none".
	// This is a hack to avoid SSH provisioning while keeping "pwd" as the machine driver
	if d.Created && counter == 1 {
		return "none"
	}
	return "pwd"
}

func (d *Driver) GetIP() (string, error) {
	return d.IPAddress, nil
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) GetSSHHostname() (string, error) {
	return "", notImplemented
}

func (d *Driver) GetSSHPort() (int, error) {
	return 0, notImplemented
}

func (d *Driver) GetSSHUsername() string {
	return "unsupported"
}

func (d *Driver) GetURL() (string, error) {
	return d.URL, nil
}

func (d *Driver) GetState() (state.State, error) {
	return state.Running, nil
}

func (d *Driver) Kill() error {
	return notImplemented
}

func (d *Driver) PreCreateCheck() error {
	/*
		if d.StorePath == filepath.Join(mcnutils.GetHomeDir(), ".docker", "machine") {
			return errors.New("Default storage path is discouraged when using PWD driver. Use -s flag or MACHINE_STORAGE_PATH env variable to set one")
		}
	*/
	if d.SessionId == "" {
		return errors.New("Session Id must be specified")
	}
	return nil
}

func (d *Driver) Remove() error {
	r, _ := http.NewRequest("DELETE", fmt.Sprintf("http://%s:%s/sessions/%s/instances/%s", d.Hostname, d.Port, d.SessionId, d.InstanceName), nil)
	resp, err := http.DefaultClient.Do(r)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println(err, resp)
		return errors.New("Error removing instance")
	}
	return nil
}

func (d *Driver) Restart() error {
	return notImplemented
}

func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	dump(opts)
	d.SessionId = opts.String("pwd-session-id")
	d.Hostname = opts.String("pwd-hostname")
	d.SSLPort = opts.String("pwd-ssl-port")
	d.Port = opts.String("pwd-port")
	return nil
}

func (d *Driver) Start() error {
	return notImplemented
}

func (d *Driver) Stop() error {
	return notImplemented
}
