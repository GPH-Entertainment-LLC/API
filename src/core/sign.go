package core

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudfront/sign"
)

func SignItemContentUrl(itemUrl string, pemKey string, keyPairID string) (*string, error) {
	if itemUrl == "" {
		return nil, nil
	}

	// Parse the PEM block
	pemBlock, _ := pem.Decode([]byte(pemKey))
	if pemBlock == nil || pemBlock.Type != "RSA PRIVATE KEY" {
		return nil, &ErrorResp{Message: "Failed to decode PEM block containing RSA private key"}
	}

	// Parse the RSA private key
	privateKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	signer := sign.NewURLSigner(keyPairID, privateKey)
	signedURL, err := signer.Sign(itemUrl, time.Now().Add(1*time.Hour))
	if err != nil {
		return nil, err
	}
	return &signedURL, nil
}

func SignUrlBatch(urlMap map[int]map[string]*string, pemKey string, keyPairId string) (map[int]map[string]*string, error) {
	for k := range urlMap {
		mainUrl := urlMap[k]["contentMainUrl"]
		thumbUrl := urlMap[k]["contentThumbUrl"]

		signedMainUrl := mainUrl
		signedThumbUrl := thumbUrl
		var err error
		if mainUrl != nil {
			signedMainUrl, err = SignItemContentUrl(*mainUrl, pemKey, keyPairId)
			if err != nil {
				return nil, err
			}
		}

		if signedThumbUrl != nil {
			signedThumbUrl, err = SignItemContentUrl(*thumbUrl, pemKey, keyPairId)
			if err != nil {
				return nil, err
			}
		}

		urlMap[k]["contentMainUrl"] = signedMainUrl
		urlMap[k]["contentThumbUrl"] = signedThumbUrl
	}
	return urlMap, nil
}

func SignUrlBatch2(urlBatch map[int]map[string]*string) (map[int]map[string]*string, error) {
	baseUrl := os.Getenv("CONTENT_SVC_BASE_URL")
	url := baseUrl + "/content/sign/batch"

	// Serialize the map to JSON
	jsonUrlBatch, err := json.Marshal(urlBatch)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return nil, err
	}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonUrlBatch))
	if err != nil {
		return nil, err
	}
	apiKey := os.Getenv("CONTENT_API_KEY")
	request.Header.Set("Authorization", apiKey)

	client := http.Client{}
	maxAttempts := 3
	attempt := 0
	for attempt < maxAttempts {
		resp, err := client.Do(request)
		if err != nil {
			fmt.Println("Error ", err)
			return nil, err
		}

		defer resp.Body.Close()

		// Check the response status
		if resp.StatusCode != http.StatusCreated {
			fmt.Println("Did not receive successful response on sign batch url call, trying again..")
		} else {
			// Deserialize the response body into your struct
			var signedUrlBatch map[int]map[string]*string
			err = json.NewDecoder(resp.Body).Decode(&signedUrlBatch)
			if err != nil {
				fmt.Println("Error decoding JSON:", err)
				return nil, err
			}
			return signedUrlBatch, nil
		}
		fmt.Println("Response Code: ", resp.StatusCode)
		attempt += 1
		time.Sleep(time.Second * 2)
	}
	fmt.Println("After 3 attempts, could not get signed url batch from content-service")
	return nil, &ErrorResp{Message: "Unable to receive signed url batch from content-service"}
}
