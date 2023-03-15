package zapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/FreakinRocket/zjson"
)

// ### CONSTANTS ###

// struct contains information that should remain secret and not be included in the online repository
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Host         string `json:"api_URL"`
	Code         string `json:"code"`
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	FilePath     string `json:"-"` //does not save file path into json
}

// make a call to the configured API, stores response into input struct and check for error
func ApiCall(uri string, v any, c *Config) {
	ChkError(json.Unmarshal(tryGet(uri, c), v))
}

// performs a http GET call with oauth2 bearer presentation in header
func HttpGet(host, uri string, bearer string) (respBody []byte, status int) {
	req, err := http.NewRequest("GET", host+uri, nil)
	ChkError(err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	ChkError(err)
	defer resp.Body.Close()

	respBody, err = io.ReadAll(resp.Body)
	ChkError(err)
	status = resp.StatusCode

	return
}

// performs a http POST call with oauth2 bearer presentation in request body
func HttpPost(host, uri string, requestBody []byte) (respBody []byte, status int) {
	req, err := http.NewRequest("POST", host+uri, bytes.NewReader(requestBody))
	ChkError(err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	ChkError(err)
	defer resp.Body.Close()

	respBody, err = io.ReadAll(resp.Body)
	ChkError(err)
	status = resp.StatusCode

	return
}

// gets a new access token using a refresh token
func getTokenFromRefresh(c *Config) (statusCode int) {
	//use a refresh token to get an access token
	requestBody, err := json.Marshal(map[string]string{
		"refresh_token": c.RefreshToken,
		"client_id":     c.ClientID,
		"client_secret": c.ClientSecret,
	})
	ChkError(err)

	//make request
	respBody, statusCode := HttpPost(c.Host, "/token", requestBody)

	//unmarshal response
	err = json.Unmarshal(respBody, &c)
	ChkError(err)

	return
}

// a http GET where if the first one fails it gets a new access code then tries again. This is how expired codes are handled
func tryGet(uri string, c *Config) (respBody []byte) {
	respBody, status := HttpGet(c.Host, uri, c.AccessToken)
	if status != 200 {
		getToken(c)
		respBody, status = HttpGet(c.Host, uri, c.AccessToken)
		if status != 200 {
			log.Fatalln(respBody, status)
		}
	}
	return
}

// Get an initial access token and refresh token from a webpage generated authorization code. This value is entred in the config file before program launch
func getTokensFromCode(c *Config) (statusCode int) {

	//create authorization code request body
	requestBody, err := json.Marshal(map[string]string{
		"code":          c.Code,
		"client_id":     c.ClientID,
		"client_secret": c.ClientSecret,
	})
	ChkError(err)

	//make request
	respBody, statusCode := HttpPost(c.Host, "/token", requestBody)

	//unmarshal response
	json.Unmarshal(respBody, &c)

	return
}

// always attemps to get a new access token using a refresh token first, then if that fails it tries using an authorization code from the config file
func getToken(c *Config) {
	if getTokenFromRefresh(c) != 200 {
		if getTokensFromCode(c) != 200 {
			log.Fatalln("Failed to get token")
		}
	}
	zjson.SaveJSON(c, c.FilePath)
}

// saves programming time and log.fatalLn on an error
func ChkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
