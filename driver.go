package pwd

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
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
	err := setupCerts(d)
	if err != nil {
		return fmt.Errorf("Error configuring PWD certs: %v ", err)
	}
	type instance struct {
		Name string
		IP   string
	}

	resp, err := http.Post(fmt.Sprintf("http://%s:%s/sessions/%s/instances", d.Hostname, d.Port, d.SessionId), "", nil)

	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Could not create instance %v %v", err, resp)
	}

	defer resp.Body.Close()

	i := &instance{}

	json.NewDecoder(resp.Body).Decode(i)
	d.IPAddress = i.IP
	d.InstanceName = i.Name
	d.URL = fmt.Sprintf("tcp://ip%s-2375.%s:%s", strings.Replace(d.IPAddress, ".", "_", -1), d.Hostname, d.SSLPort)
	d.Created = true
	return nil

}

func setupCerts(d *Driver) error {
	resp, err := http.Get(fmt.Sprintf("http://%s:%s/keys", d.Hostname, d.Port))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println(err, resp)
		return errors.New("Error fetching keys to setup certs")
	}
	defer resp.Body.Close()

	tr := tar.NewReader(resp.Body)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		// TODO handle errors
		storageCert, _ := os.Create(d.ResolveStorePath(hdr.Name))
		machineCert, _ := os.Create(filepath.Join(d.StorePath, "certs", hdr.Name))
		defer storageCert.Close()
		defer machineCert.Close()
		w := io.MultiWriter(storageCert, machineCert)

		if _, err := io.Copy(w, tr); err != nil {
			log.Println(err)
			return errors.New("Error copying certs")
		}
	}
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
	if d.StorePath == filepath.Join(mcnutils.GetHomeDir(), ".docker", "machine") {
		return errors.New("Default storage path is discouraged when using PWD driver. Use -s flag or MACHINE_STORAGE_PATH env variable to set one")
	}
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
