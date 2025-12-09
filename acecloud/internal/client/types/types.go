package types

import "encoding/json"

type VMCreateRequest struct {
	Name                string          `json:"name"`
	Flavor              string          `json:"flavor"`
	BootUUID            string          `json:"boot_uuid"`
	DeleteOnTermination bool            `json:"delete_on_termination"`
	Networks            []string        `json:"network,omitempty"`
	SecurityGroups      []string        `json:"security_group,omitempty"`
	SourceType          string          `json:"source_type"`
	Key                 string          `json:"key"`
	AvailabilityZone    string          `json:"availability_zone"`
	BillingType         string          `json:"billing_type"`
	Volumes             []VolumeRequest `json:"volumes,omitempty"`
	Count               int             `json:"count"`
}

type VolumeRequest struct {
	Boot        bool   `json:"boot"`
	VolumeType  string `json:"volume_type"`
	Size        int    `json:"size"`
	BillingType string `json:"billing_type"`
}

type VMCreateResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    struct {
		ID string `json:"id"`
	} `json:"data"`
}

type ErrorResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

type AvailabilityZone struct {
	Name string `json:"name"`
}

type VMGetResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    struct {
		Key              string `json:"key"`
		ID               string `json:"id"`
		Name             string `json:"name"`
		Status           string `json:"status"`
		AvailabilityZone string `json:"availability_zone"`
		// Addresses: public/private
		Addresses struct {
			Public []struct {
				Version int    `json:"version"`
				Addr    string `json:"addr"`
				MacAddr string `json:"mac_addr"`
				Name    string `json:"name"`
				Type    string `json:"type"`
			} `json:"public"`
			Private []interface{} `json:"private"`
		} `json:"addresses"`
	} `json:"data"`
}

type DeleteResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type VMUpdateRequest struct {
	Name         string `json:"name"`
	CustomUpdate string `json:"custom_update,omitempty"`
}

type VMUpdateResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    json.RawMessage
}
