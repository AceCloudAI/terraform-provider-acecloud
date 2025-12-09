package types

//*what we are sending in the request to create a key-pair
//*"name":"test-key"
type KeyPairCreateRequest struct {
	Name string `json:"name,omitempty"`
}

type KeyPairDeleteRequest struct {
	Key    string   `json:"key,omitempty"`
	Values []string `json:"values,omitempty"`
}

func KeyPairDeleteRequestFromIDs() *KeyPairDeleteRequest {
	return &KeyPairDeleteRequest{
		Key:    "id",
		Values: []string{},
	}
}

//* response of our API
type KeyPairData struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PublicKey   string `json:"publicKey,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	PrivateKey  string `json:"privateKey,omitempty"`
	Type        string `json:"type,omitempty"`
}

//*strucutred response for when we get list of key-pairs
type KeyPairListResponse struct {
	Error   bool          `json:"error"`
	Message string        `json:"message"`
	Data    []KeyPairData `json:"data"`
}

//*KeyPairResponse represents the response structure for a single key pair
type KeyPairResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    KeyPairData `json:"data"`
}
