package client

type Spec struct {
	// GitHub App authentication parameters
	AppID          int64  `json:"app_id,omitempty"`
	InstallationID int64  `json:"installation_id,omitempty"`
	PrivateKey     string `json:"private_key,omitempty"`
}
