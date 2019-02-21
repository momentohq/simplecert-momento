//
//  simplecert
//
//  Created by Philipp Mieden
//  Contact: dreadl0ck@protonmail.ch
//  Copyright © 2018 bestbytes. All rights reserved.
//

package simplecert

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/lextoumbourou/goodhosts"
)

// updateHosts is used in local mode
// to add all host entries for the domains
func updateHosts() {
	hosts, err := goodhosts.NewHosts()
	if err != nil {
		log.Fatal("[FATAL] could not open hostsfile: ", err)
	}

	for _, d := range c.Domains {
		if !hosts.Has(localhost, d) {
			hosts.Add(localhost, d)
		}
	}

	if err := hosts.Flush(); err != nil {
		log.Fatal("[FATAL] could not update /etc/hosts: ", err)
	}
}

// createLocalCert first creates a local root CA for mkcert
// and then generates a trusted certificate for the domains specified in the configuration
func createLocalCert(certFilePath, keyFilePath string) {

	log.Println("[INFO] no cached cert found. Creating a new one for local development...")

	// run mkcert to create root CA
	runCommand("mkcert", "-install")

	// run mkcert to generate the certificate
	runCommand("mkcert", c.Domains...)

	var (
		newCertFile string
		newKeyFile  string
	)
	if len(c.Domains) > 1 {
		newCertFile = c.Domains[0] + "+" + strconv.Itoa(len(c.Domains)-1) + ".pem"
		newKeyFile = c.Domains[0] + "+" + strconv.Itoa(len(c.Domains)-1) + "-key.pem"
	} else {
		newCertFile = c.Domains[0] + ".pem"
		newKeyFile = c.Domains[0] + "-key.pem"
	}

	// rename certificate file
	log.Println("[INFO] renaming", newCertFile, "to", certFilePath)
	err := os.Rename(newCertFile, certFilePath)
	if err != nil {
		log.Fatal("[FATAL] failed to rename cert file: ", err)
	}

	// rename key file
	log.Println("[INFO] renaming", newKeyFile, "to", keyFilePath)
	err = os.Rename(newKeyFile, keyFilePath)
	if err != nil {
		log.Fatal("[FATAL] failed to rename key file: ", err)
	}
}

// domainsChanged check the stored domains when running in local mode
// if they dont match the domains from the configuration
// this function returns true
func domainsChanged(certFilePath, keyFilePath string) bool {

	// read certificate data from disk
	certData, err := ioutil.ReadFile(certFilePath)
	if err != nil {
		log.Fatal("[FATAL] simplecert could not load X509 key pair: ", err)
	}

	// PEM decode
	block, _ := pem.Decode(certData)
	if block == nil {
		panic("failed to parse certificate PEM")
	}

	// parse certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Fatal("[FATAL] simplecert could not load X509 key pair: ", err)
	}

	// if the number of entries is not equal, bail out.
	if len(cert.DNSNames) != len(c.Domains) {
		return true
	}

	// compare each entry
	for i, d := range cert.DNSNames {
		if d != c.Domains[i] {
			return true
		}
	}

	// identical
	return false
}