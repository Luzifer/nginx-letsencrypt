package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Luzifer/rconfig"
	log "github.com/Sirupsen/logrus"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/xenolf/lego/acme"
)

var (
	cfg = struct {
		BufferTime     time.Duration `flag:"buffer" default:"360h" description:"How long before expiry to mark the certificate not longer fine"`
		NginxConfigs   []string      `flag:"nginx-config" description:"Config files to collect server names from"`
		Email          string        `flag:"email" description:"Email for registration with LetsEncrypt"`
		ListenHTTP     string        `flag:"listen-http" default:":5002" description:"IP/Port to listen on for challenge proxying"`
		ACMEServer     string        `flag:"server" default:"https://acme-v01.api.letsencrypt.org/directory" description:"ACME URL"`
		StorageDir     string        `flag:"storage-dir" default:"~/.config/nginx-letsencrypt" description:"Directory to cache registration"`
		VersionAndExit bool          `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	version = "dev"
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

	nameGroups, err := collectServerNameGroups(collectServerNames())
	if err != nil {
		log.Fatalf("Unable to collect server names: %s", err)
	}

	hadErrors := false
	for sld, domains := range nameGroups {
		if err := createCertificate(client, sld, domains); err != nil {
			log.Errorf("Failed to create certificate for second level domain %q: %s", sld, err)
			hadErrors = true
		}
	}

	if hadErrors {
		log.Fatalf("At least one second level domain had errors, failing now.")
	}
}
