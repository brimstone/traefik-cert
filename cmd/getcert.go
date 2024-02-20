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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/brimstone/traefik-cert/client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// getcertCmd represents the getcert command
var (
	getcertCmd = &cobra.Command{
		Use:   "getcert",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: getcertFunc,
	}
	getcertFlags *pflag.FlagSet
)

func initgetcertFlags() {
	if getcertFlags != nil {
		return
	}
	getcertFlags = pflag.NewFlagSet("getcert", pflag.ContinueOnError)
	getcertFlags.StringP("url", "u", "", "Base URL for companion server [$URL]")
	viper.BindPFlag("url", getcertFlags.Lookup("url"))
	viper.BindEnv("url")

	getcertFlags.StringP("domain", "d", "", "Domain of cert to retrieve [$DOMAIN]")
	viper.BindPFlag("domain", getcertFlags.Lookup("domain"))
	viper.BindEnv("domain")

	getcertFlags.StringP("jwt", "j", "", "JWT for authentication [$JWT]")
	viper.BindPFlag("jwt", getcertFlags.Lookup("jwt"))
	viper.BindEnv("jwt")

	getcertFlags.StringP("cert", "c", "", "Path to save cert file [$CERT]")
	viper.BindPFlag("cert", getcertFlags.Lookup("cert"))
	viper.BindEnv("cert")

	getcertFlags.StringP("key", "k", "", "Path to save key file [$KEY]")
	viper.BindPFlag("key", getcertFlags.Lookup("key"))
	viper.BindEnv("key")

	getcertFlags.StringP("owner", "o", "", "Owner and optional group for files [$KEY]")
	viper.BindPFlag("owner", getcertFlags.Lookup("owner"))
	viper.BindEnv("owner")
}

func init() {
	rootCmd.AddCommand(getcertCmd)
	initgetcertFlags()
	getcertCmd.Flags().AddFlagSet(getcertFlags)
}

func getcertFunc(cmd *cobra.Command, args []string) error {
	domain := viper.GetString("domain")
	if domain == "" {
		return errors.New("must specify domain of cert to retrieve")
	}

	jwt := viper.GetString("jwt")
	if jwt == "" {
		return errors.New("JWT must be specified")
	}

	url := viper.GetString("url")
	if url == "" {
		return errors.New("must specify URL holding certs")
	}

	certfile := viper.GetString("cert")
	keyfile := viper.GetString("key")

	cert, key, err := client.GetCert(url, domain, jwt)

	if err != nil {
		return err
	}

	if cert == "" {
		return errors.New("No cert in response from server")
	}

	if certfile == "" {
		fmt.Println(string(cert))
	} else {
		err = ioutil.WriteFile(certfile, cert, 0600)
		if err != nil {
			return err
		}
	}

	if keyfile == "" {
		fmt.Println(string(key))
	} else {
		err = ioutil.WriteFile(keyfile, key, 0600)
		if err != nil {
			return err
		}
	}

	ownerFlag := viper.GetString("owner")
	if ownerFlag == "" {
		return nil
	}

	var owner *user.User

	ownergroup := strings.Split(ownerFlag, ":")
	if len(ownergroup) > 2 {
		return errors.New("Only one colon allowed in owner flag")
	}

	ownerID, err := strconv.ParseInt(ownergroup[0], 10, 10)
	if err != nil {
		owner, err = user.Lookup(ownergroup[0])
		if err != nil {
			return err
		}
		ownerID, err = strconv.ParseInt(owner.Uid, 10, 10)
		if err != nil {
			return err
		}
	}

	err = os.Chown(certfile, int(ownerID), -1)
	if err != nil {
		return err
	}
	err = os.Chown(keyfile, int(ownerID), -1)
	if err != nil {
		return err
	}

	if len(ownergroup) == 1 {
		// Don't change the group
		return nil
	}

	var ownerGID int64
	// Set the default group for the user
	if ownergroup[1] == "" {
		ownerGID, err = strconv.ParseInt(owner.Gid, 10, 10)
	} else {
		ownerGID, err = strconv.ParseInt(ownergroup[1], 10, 10)
		if err != nil {
			group, err := user.LookupGroup(ownergroup[1])
			if err != nil {
				return err
			}
			ownerGID, err = strconv.ParseInt(group.Gid, 10, 10)
			if err != nil {
				return err
			}
		}
	}

	err = os.Chown(certfile, -1, int(ownerGID))
	if err != nil {
		return err
	}
	err = os.Chown(keyfile, -1, int(ownerGID))
	if err != nil {
		return err
	}

	return nil
}
