package bfladetector

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func GetUserID(headers map[string]string) (*DetectedUser, error) {
	if xcustomerID, ok := headers["x-customer-id"]; ok {
		return &DetectedUser{Source: DetectedUserSourceXConsumerIDHeader, ID: xcustomerID}, nil
	}
	authz, ok := headers["authorization"]
	if !ok {
		return nil, nil
	}
	if strings.HasPrefix(authz, "Basic ") {
		basic := strings.TrimPrefix(authz, "Basic ")
		usernameAndPassword, err := base64.StdEncoding.DecodeString(basic)
		if err != nil {
			return nil, fmt.Errorf("cannot decode basic authz header: %w", err)
		}
		usernameAndPasswordParts := strings.Split(string(usernameAndPassword), ":")

		// nolint:gomnd
		if len(usernameAndPasswordParts) < 2 {
			return nil, errors.New("broken basic auth header")
		}
		return &DetectedUser{Source: DetectedUserSourceBasic, ID: usernameAndPasswordParts[0]}, nil
	}
	if strings.HasPrefix(authz, "Bearer ") {
		bearer := strings.TrimPrefix(authz, "Bearer ")
		bearerParts := strings.Split(bearer, ".")

		// nolint:gomnd
		if len(bearerParts) == 3 { // is JWT
			s, err := base64.URLEncoding.DecodeString(bearerParts[1])
			if err != nil {
				return nil, fmt.Errorf("unable to decode bearer token: %w", err)
			}
			data := struct {
				Subject string `json:"sub"`
			}{}
			if err := json.Unmarshal(s, &data); err != nil {
				return nil, fmt.Errorf("unable to unmarshal json jwt body: %w", err)
			}
			return &DetectedUser{Source: DetectedUserSourceJWT, ID: data.Subject}, nil
		}

		return nil, ErrUnsupportedAuthScheme
	}
	return nil, ErrUnsupportedAuthScheme
}
