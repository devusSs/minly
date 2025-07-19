package yourls

import (
	"errors"
	"net/http"
)

type Client struct {
	endpoint  string
	signature string
	client    *http.Client

	title string
}

func NewClient(endpoint string, signature string) (*Client, error) {
	if signature == "" {
		return nil, errors.New("signature is required")
	}

	return &Client{
		endpoint:  endpoint,
		signature: signature,
		client:    http.DefaultClient,

		title: "Uploaded using minly (github.com/devusSs/minly)",
	}, nil
}
