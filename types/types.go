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

package types

type Acme struct {
	Account struct {
		Email        string `json:"Email"`
		Registration struct {
			Body struct {
				Status string `json:"status"`
			} `json:"body"`
			URI string `json:"uri"`
		} `json:"Registration"`
		PrivateKey string `json:"PrivateKey"`
	} `json:"Account"`
	Certificates []struct {
		Domain struct {
			Main string      `json:"Main"`
			SANs interface{} `json:"SANs"`
		} `json:"Domain"`
		Certificate string `json:"Certificate"`
		Key         string `json:"Key"`
	} `json:"Certificates"`
	HTTPChallenges struct {
	} `json:"HTTPChallenges"`
}

type Auth struct {
	Cert struct {
		Domains []string `json:"domains"`
	} `json:"cert"`
}

type CertResponse struct {
	Cert []byte `json:"cert"`
	Key  []byte `json:"key"`
}
