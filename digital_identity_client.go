package yoti

import (
	"crypto/rsa"
	"os"

	"github.com/getyoti/yoti-go-sdk/v3/cryptoutil"
	"github.com/getyoti/yoti-go-sdk/v3/digitalidentity"
	"github.com/getyoti/yoti-go-sdk/v3/requests"
)

const DefaultURL = "https://api.yoti.com/share/"

// DigitalIdentityClient represents a client that can communicate with yoti and return information about Yoti users.
type DigitalIdentityClient struct {
	// SdkID represents the SDK ID and NOT the App ID. This can be found in the integration section of your
	// application hub at https://hub.yoti.com/
	SdkID string

	// Key should be the security key given to you by yoti (see: security keys section of
	// https://hub.yoti.com) for more information about how to load your key from a file see:
	// https://github.com/getyoti/yoti-go-sdk/blob/master/README.md
	Key *rsa.PrivateKey

	apiURL     string
	HTTPClient requests.HttpClient // Mockable HTTP Client Interface
}

// NewDigitalIdentityClient constructs a Client object
func NewDigitalIdentityClient(sdkID string, key []byte) (*DigitalIdentityClient, error) {
	decodedKey, err := cryptoutil.ParseRSAKey(key)

	if err != nil {
		return nil, err
	}

	return &DigitalIdentityClient{
		SdkID: sdkID,
		Key:   decodedKey,
	}, err
}

// OverrideAPIURL overrides the default API URL for this Yoti Client
func (client *DigitalIdentityClient) OverrideAPIURL(apiURL string) {
	client.apiURL = apiURL
}

func (client *DigitalIdentityClient) getAPIURL() string {
	if client.apiURL != "" {
		return client.apiURL
	}

	if value, exists := os.LookupEnv("YOTI_API_URL"); exists && value != "" {
		return value
	}

	return DefaultURL
}

// GetSdkID gets the Client SDK ID attached to this client instance
func (client *DigitalIdentityClient) GetSdkID() string {
	return client.SdkID
}

// CreateShareSession creates a sharing session to initiate a sharing process based on a policy
func (client *DigitalIdentityClient) CreateShareSession(shareSession *digitalidentity.ShareSessionRequest) (share digitalidentity.ShareSession, err error) {
	return digitalidentity.CreateShareSession(client.HTTPClient, shareSession, client.GetSdkID(), client.getAPIURL(), client.Key)
}

// GetShareSession retrieves the sharing session.
func (client *DigitalIdentityClient) GetSession(sessionID string) (shareSession *digitalidentity.ShareSession, err error) {
	return digitalidentity.GetSession(client.HTTPClient, sessionID, client.GetSdkID(), client.getAPIURL(), client.Key)
}
