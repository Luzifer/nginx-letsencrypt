package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/Luzifer/rconfig"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/xenolf/lego/acme"
)

var (
	cfg = struct {
		BufferTime     time.Duration `flag:"buffer" env:"BUFFER" default:"360h" description:"How long before expiry to mark the certificate not longer fine"`
		NginxConfig    string        `flag:"nginx-config" env:"NGINX_CONFIG" description:"Config file to collect server names and start nginx from"`
		Email          string        `flag:"email" env:"EMAIL" description:"Email for registration with LetsEncrypt"`
		ListenHTTP     string        `flag:"listen-http" default:":5001" description:"IP/Port to listen on for challenge proxying"`
		LogLevel       string        `flag:"log-level" default:"info" description:"Log level to use (debug, info, warning, error, ...)"`
		ACMEServer     string        `flag:"server" default:"https://acme-v01.api.letsencrypt.org/directory" description:"ACME URL"`
		StorageDir     string        `flag:"storage-dir" env:"STORAGE_DIR" default:"~/.config/nginx-letsencrypt" description:"Directory to cache registration"`
		VersionAndExit bool          `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	version = "dev"

	nginx         *exec.Cmd
	configVersion string
)

func init() {
	if err := rconfig.Parse(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("nginx-letsencrypt %s\n", version)
		os.Exit(0)
	}

	if ep, err := homedir.Expand(cfg.StorageDir); err != nil {
		log.Fatalf("Failed to expand storage dir: %s", err)
	} else {
		cfg.StorageDir = ep
	}

	if lvl, err := log.ParseLevel(cfg.LogLevel); err == nil {
		log.SetLevel(lvl)
	} else {
		log.Fatalf("Failed to parse log level: %s", err)
	}
}

func main() {
	myUser, err := loadOrCreateUser()
	if err != nil {
		log.Fatalf("Unable to load / create user: %s", err)
	}

	client, err := acme.NewClient(cfg.ACMEServer, myUser, acme.RSA2048)
	if err != nil {
		log.Fatal(err)
	}

	if myUser.Registration == nil {
		reg, err := client.Register()
		if err != nil {
			log.Fatalf("Unable to register for LetsEncrypt account: %s", err)
		}
		myUser.Registration = reg

		if err = client.AgreeToTOS(); err != nil {
			log.Fatalf("Failed to accept TOS: %s", err)
		}

		if err = myUser.Save(); err != nil {
			log.Fatalf("Unable to save user file: %s", err)
		}
	}

	client.SetHTTPAddress(cfg.ListenHTTP)
	client.ExcludeChallenges([]acme.Challenge{acme.TLSSNI01, acme.DNS01})

	go func() {
		manageNginxConfig(client)
		for range time.Tick(time.Minute) {
			manageNginxConfig(client)
		}
	}()

	for {
		nginx = exec.Command("nginx", "-c", cfg.NginxConfig, "-g", "daemon off;")
		nginx.Stdout = os.Stdout
		nginx.Stderr = os.Stderr
		log.Errorf("nginx process ended: %s", nginx.Run())
		<-time.After(500 * time.Millisecond)
	}
}

func manageNginxConfig(client *acme.Client) {
	var (
		currentConfigVersion = hashFile(cfg.NginxConfig)
		needsReload          = currentConfigVersion != configVersion
	)

	nameGroups, err := collectServerNameGroups(collectServerNames())
	if err != nil {
		log.Fatalf("Unable to collect server names: %s", err)
	}

	if err := ensureCertFiles(nameGroups); err != nil {
		log.Fatalf("Unable to link initial certificates: %s", err)
	}

	for nginx == nil || nginx.Process == nil {
		// Don't start executing certificate fetch if server is not running, we need it
		<-time.After(100 * time.Millisecond)
	}

	hadErrors := false
	for sld, domains := range nameGroups {
		if newCert, err := createCertificate(client, sld, domains); err != nil {
			log.Errorf("Failed to create certificate for second level domain %q: %s", sld, err)
			hadErrors = true
		} else {
			if newCert {
				needsReload = true
			}
		}
	}

	if hadErrors {
		return
	}

	if needsReload && nginx != nil {
		log.Infof("Reloading nginx to apply config / certificates")
		nginx.Process.Signal(syscall.SIGHUP)
		configVersion = currentConfigVersion
	}
}

func hashFile(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Unable to read nginx config: %s", err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
