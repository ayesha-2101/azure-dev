package requesthelper

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func SendRawARMRequest(
	ctx context.Context,
	httpMethod string,
	url string,
	body string,
) (*http.Response, error) {

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		fmt.Printf("failed to create azure credentials: %v\n", err)
		return nil, err
	}
	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		fmt.Printf("failed to get token from credentials: %v\n", err)
		return nil, err
	}

	req, err := http.NewRequest(httpMethod, url, strings.NewReader(body))
	if err != nil {
		fmt.Printf("failed to create http request: %v\n", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("failed to send http request: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	return resp, nil
}
