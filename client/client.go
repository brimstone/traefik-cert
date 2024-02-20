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

package client

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/brimstone/logger"
	"github.com/brimstone/traefik-cert/types"
)

func GetCert(url string, domain string, jwt string) (cert []byte, key []byte, err error) {
	log := logger.New()

	// Check environment
	if url == "" {
		url = os.Getenv("URL")
		if url == "" {
			err = errors.New("URL must not be empty")
			return
		}
	}
	if domain == "" {
		domain = os.Getenv("DOMAIN")
		if domain == "" {
			err = errors.New("DOMAIN must not be empty")
			return
		}
	}
	if jwt == "" {
		jwt = os.Getenv("JWT")
		if jwt == "" {
			err = errors.New("JWT must not be empty")
			return
		}
	}
	// Build baseurl
	baseurl := url
	if os.Getenv("DEV") == "" {
		if strings.HasPrefix(url, "http://") {
			err = errors.New("URL must not be http")
			return
		}
		if !strings.HasPrefix(url, "https://") {
			baseurl = "https://" + baseurl
		}
	} else {
		if !strings.HasPrefix(url, "http://") {
			baseurl = "http://" + baseurl
		}
	}
	log.Debug("baseurl",
		log.Field("baseurl", baseurl),
	)
	// Actually make request
	var req *http.Request
	req, err = http.NewRequest("GET", baseurl+"/cert/"+domain, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+jwt)

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = errors.New(string(body))
		return
	}

	var response types.CertResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return
	}
	cert = response.Cert
	key = response.Key
	return
}
