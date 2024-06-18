package service

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"xo-packs/core"
)

type FirebaseService interface {
	GetUserByEmail(context.Context, string) (*http.Response, error)
	DeleteUser(context.Context, string, string, UserService) (*http.Response, error)
}

type FirebaseSvcImpl struct{}

func NewFirebaseSvc() FirebaseService {
	return &FirebaseSvcImpl{}
}

type FirebaseError struct {
	message string
}

func (e *FirebaseError) Error() string {
	return e.message
}

func (s *FirebaseSvcImpl) GetUserByEmail(ctx context.Context, email string) (*http.Response, error) {
	client := &http.Client{}

	url, err := url.Parse(os.Getenv("GET_FIREBASE_USER_URL"))
	if err != nil {
		return nil, err
	}

	query := url.Query()
	query.Add("email", email)
	url.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}
	apiKey, err := core.GetSecret(os.Getenv("FIREBASE_API_KEY"))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)

	return client.Do(req)
}

func (s *FirebaseSvcImpl) DeleteUser(ctx context.Context, uid string, username string, userService UserService) (*http.Response, error) {
	client := &http.Client{}

	url, err := url.Parse(os.Getenv("DELETE_FIREBASE_USER_URL"))
	if err != nil {
		return nil, err
	}

	query := url.Query()
	query.Add("uid", uid)
	url.RawQuery = query.Encode()

	req, err := http.NewRequest("DELETE", url.String(), nil)
	if err != nil {
		return nil, err
	}
	apiKey, err := core.GetSecret(os.Getenv("FIREBASE_API_KEY"))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	_, err = userService.DeleteUser(ctx, uid, username)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
