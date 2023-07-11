package main

import (
	"go.clever-cloud.dev/client"
)

func NewClient(clevercloudToken, clevercloudSecret string) *client.Client {
	return client.New(client.WithUserOauthConfig(clevercloudToken, clevercloudSecret))
}
