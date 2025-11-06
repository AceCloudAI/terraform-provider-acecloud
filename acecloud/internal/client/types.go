package client

type VMCreateRequest struct {
	Name                string          `json:"name"`
	Flavor              string          `json:"flavor"`
	BootUUID            string          `json:"boot_uuid"`
	DeleteOnTermination bool            `json:"delete_on_termination"`
	Networks            []string        `json:"network,omitempty"`
	SecurityGroups      []string        `json:"security_group,omitempty"`
	SourceType          string          `json:"source_type"`
	Key                 string 		    `json:"key"`		
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
