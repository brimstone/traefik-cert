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
	"github.com/brimstone/traefik-cert/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := server.NewServer(server.ServerOptions{
			Address:  viper.GetString("address"),
			Key:      viper.GetString("public"),
			AcmeFile: viper.GetString("acme"),
		})
		if err != nil {
			return err
		}
		return s.Serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringP("address", "l", ":80", "Address on which to listen [$ADDRESS]")
	viper.BindPFlag("address", serveCmd.Flags().Lookup("address"))
	viper.BindEnv("address")

	serveCmd.Flags().StringP("public", "k", "public.key", "Key to use to validate requests [$PUBLIC]")
	viper.BindPFlag("public", serveCmd.Flags().Lookup("public"))
	viper.BindEnv("public")

	serveCmd.Flags().StringP("acme", "f", "/acme/acme.json", "Path to ACME JSON [$ACME]")
	viper.BindPFlag("acme", serveCmd.Flags().Lookup("acme"))
	viper.BindEnv("acme")
}
