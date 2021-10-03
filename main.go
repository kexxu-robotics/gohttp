package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/yaml.v2"
)

// example dev certs:
//mkdir -p certs
//openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
//	-keyout certs/localhost.key -out certs/localhost.crt \
//	-subj "/C=NL/ST=Limburg/L=Geleen/O=Marco Franssen/OU=Development/CN=localhost/emailAddress=marco.franssen@gmail.com"

type justFilesFilesystem struct {
	fs http.FileSystem
}

func (fs justFilesFilesystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return neuteredReaddirFile{f}, nil
}

type neuteredReaddirFile struct {
	http.File
}

func (f neuteredReaddirFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

type Conf struct {
	Port       int
	Domain     string
	StaticPath string
}

var conf Conf

func main() {
	// load conf
	data, err := ioutil.ReadFile("conf.yaml")
	if err != nil {
		panic(err)
	}
	conf = Conf{}
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		panic(err)
	}
	fmt.Println("loaded conf", conf)
	if conf.StaticPath == "" {
		conf.StaticPath = "www"
	}

	mux := http.NewServeMux()
	mux.Handle("/",
		http.FileServer(
			justFilesFilesystem{http.Dir(conf.StaticPath)}),
	)

	var httpsSrv *http.Server

	if conf.Port == 443 {
		fmt.Println("using SSL protection")
		conf.Port = 80 // also serve port 80 when serving https

		dataDir := "." // where the certs are saved
		hostPolicy := func(ctx context.Context, host string) error {
			// Note: change to your real domain
			allowedHost := conf.Domain
			if host == allowedHost {
				return nil
			}
			return fmt.Errorf("acme/autocert: only %s host is allowed", allowedHost)
		}

		httpsSrv = &http.Server{
			Addr:         ":443",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  120 * time.Second,
			Handler:      mux,
		}
		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: hostPolicy,
			Cache:      autocert.DirCache(dataDir),
		}
		httpsSrv.Addr = ":443"
		httpsSrv.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}

		go func() {
			err := httpsSrv.ListenAndServeTLS("", "")
			if err != nil {
				log.Fatalf("httpsSrv.ListendAndServeTLS() failed with %s", err)
			}
		}()
		httpSrv := &http.Server{
			Addr:         ":80",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  120 * time.Second,
			Handler:      m.HTTPHandler(mux),
		}
		if err := httpSrv.ListenAndServe(); err != nil {
			fmt.Println(err)
		}

	} else {
		httpSrv := &http.Server{
			Addr:         fmt.Sprint(":", conf.Port),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  120 * time.Second,
			Handler:      mux,
		}
		if err := httpSrv.ListenAndServe(); err != nil {
			fmt.Println(err)
		}
	}

}
