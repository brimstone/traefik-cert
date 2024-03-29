/*

Copyright © 2024 Matt Robinson <brimstone@the.narro.ws>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package cmd

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/brimstone/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func mkCert(certfile string) {
	template := &x509.Certificate{
		IsCA:                  true,
		BasicConstraintsValid: true,
		SubjectKeyId:          []byte{1, 2, 3},
		SerialNumber:          big.NewInt(1234),
		Subject: pkix.Name{
			Country:      []string{"Earth"},
			Organization: []string{"Mother Nature"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Second * 5),
		// see http://golang.org/pkg/crypto/x509/#KeyUsage
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	// generate private key
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		panic(err)
	}

	publickey := &privatekey.PublicKey

	// create a self-signed certificate. template = parent
	var parent = template
	cert, err := x509.CreateCertificate(rand.Reader, template, parent, publickey, privatekey)
	if err != nil {
		panic(err)
	}

	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	}

	cert = pem.EncodeToMemory(block)

	err = ioutil.WriteFile(certfile, cert, 0777)

	if err != nil {
		panic(err)
	}
}

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a command, after fetching certs",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New()
		cert, err := getValidCert(0, cmd)
		if err != nil {
			return err
		}
		// Create command
		var child *exec.Cmd

		restartChild := false

		go func() {
			for {
				killChild := true
				certSig := cert.Signature
				delay := cert.NotAfter.Sub(time.Now())
				log.Println("Cert expires in", delay)
				if delay > time.Hour*24 {
					delay = time.Hour * 24
					killChild = false
				}
				log.Println("Recheck in", delay)
				time.Sleep(delay)
				log.Println("Cert expired, renewing")
				cert, err = getValidCert(0, cmd)
				if err != nil {
					return
				}
				if bytes.Equal(certSig, cert.Signature) {
					killChild = false
				}
				if killChild {
					restartChild = true
					child.Process.Signal(syscall.SIGTERM)
				}
			}
		}()
		time.Sleep(time.Second)
		for {
			restartChild = false
			child = exec.Command(args[0], args[1:]...)
			for _, e := range os.Environ() {
				if strings.HasPrefix(e, "JWT=") {
					continue
				}
				child.Env = append(child.Env, e)
			}
			child.Stdin = os.Stdin
			child.Stdout = os.Stdout
			child.Stderr = os.Stderr
			log.Println("Starting child")
			err = child.Run()
			exitcode := child.ProcessState.ExitCode()
			log.Printf("Command finished with error(%d): %s", exitcode, err)
			if !restartChild {
				break
			}
		}

		return err
	},
}

func getValidCert(tries int, cmd *cobra.Command) (*x509.Certificate, error) {
	certfile := viper.GetString("cert")
	if tries > 2 {
		return nil, errors.New("Too many tries")
	} else if tries == 0 {
		//mkCert(certfile)
		err := getcertFunc(cmd, []string{})
		if err != nil {
			return nil, err
		}
	}
	cert, err := parseCertFile(certfile)
	if err != nil {
		log.Println(err)
		return getValidCert(tries+1, cmd)
	}
	if time.Now().Before(cert.NotBefore) {
		return nil, errors.New("Cert too new")
	}
	if time.Now().After(cert.NotAfter) {
		log.Println("Cert too old")
		return getValidCert(tries+1, cmd)
	}

	return cert, nil
}

func parseCertFile(certfile string) (*x509.Certificate, error) {
	certPEM, err := ioutil.ReadFile(certfile)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.New("failed to parse certificate: " + err.Error())
	}
	return cert, nil
}

func init() {
	rootCmd.AddCommand(execCmd)
	initgetcertFlags()
	execCmd.Flags().AddFlagSet(getcertFlags)
}
