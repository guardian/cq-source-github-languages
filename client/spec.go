package client

type Spec struct {
	Org            string `json:"org,omitempty"`
	AppID          string `json:"app_id,omitempty"`
	InstallationID string `json:"installation_id,omitempty"`
	PrivateKey     string `json:"private_key,omitempty"`
	PrivateKeyPath string `json:"private_key_path,omitempty"`
}
