package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"

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

	return nil
}
