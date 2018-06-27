package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/brimstone/traefik-cert/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getcertCmd represents the getcert command
var getcertCmd = &cobra.Command{
	Use:   "getcert",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		// Build baseurl
		baseurl := url
		if os.Getenv("DEV") == "" {
			if strings.HasPrefix(url, "http://") {
				return errors.New("URL must not be http")
			}
			if !strings.HasPrefix(url, "https://") {
				baseurl = "https://" + baseurl
			}
		} else {
			if !strings.HasPrefix(url, "http://") {
				baseurl = "http://" + baseurl
			}
		}
		log.Printf("baseurl: %s\n", baseurl)

		// Actually make request
		req, err := http.NewRequest("GET", baseurl+"/cert/"+domain, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+jwt)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return errors.New(string(body))
		}

		var response types.CertResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			return err
		}

		if certfile == "" {
			fmt.Println(string(response.Cert))
		} else {
			err = ioutil.WriteFile(certfile, response.Cert, 0600)
			if err != nil {
				return err
			}
		}

		if keyfile == "" {
			fmt.Println(string(response.Key))
		} else {
			err = ioutil.WriteFile(keyfile, response.Key, 0600)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getcertCmd)

	getcertCmd.Flags().StringP("url", "u", "", "Base URL for companion server [$URL]")
	viper.BindPFlag("url", getcertCmd.Flags().Lookup("url"))
	viper.BindEnv("url")

	getcertCmd.Flags().StringP("domain", "d", "", "Domain of cert to retrieve [$DOMAIN]")
	viper.BindPFlag("domain", getcertCmd.Flags().Lookup("domain"))
	viper.BindEnv("domain")

	getcertCmd.Flags().StringP("jwt", "j", "", "JWT for authentication [$JWT]")
	viper.BindPFlag("jwt", getcertCmd.Flags().Lookup("jwt"))
	viper.BindEnv("jwt")

	getcertCmd.Flags().StringP("cert", "c", "", "Path to save cert file [$CERT]")
	viper.BindPFlag("cert", getcertCmd.Flags().Lookup("cert"))
	viper.BindEnv("cert")

	getcertCmd.Flags().StringP("key", "k", "", "Path to save key file [$KEY]")
	viper.BindPFlag("key", getcertCmd.Flags().Lookup("key"))
	viper.BindEnv("key")
}
