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
