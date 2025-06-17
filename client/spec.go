package client

type AppAuth struct {
	Org            string `json:"org,omitempty"`
	AppID          string `json:"app_id,omitempty"`
	InstallationID string `json:"installation_id,omitempty"`
	PrivateKey     string `json:"private_key,omitempty"`
	PrivateKeyPath string `json:"private_key_path,omitempty"`
}

type Spec struct {
	// GitHub App authentication parameters
	AppAuth *AppAuth `json:"app_auth,omitempty"`
}
