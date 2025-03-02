package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func GetCountryByIP(ip string) (string, error) {
	resp, err := http.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	fmt.Println(data)

	country, ok := data["country"].(string)
	if !ok {
		return "", errors.New("failed to parse country from response")
	}

	return country, nil
}

func GetPublicIP() (string, error) {
	resp, err := http.Get("https://api64.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}
