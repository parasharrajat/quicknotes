package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/garyburd/go-oauth/oauth"
	"github.com/gorilla/securecookie"
	"github.com/kjk/u"
	"golang.org/x/oauth2"
)

const (
	cookieAuthKeyHexStr = "513521f0ef43c9446ed7bf359a5a9700ef5fa5a5eb15d0db5eae8e93856d99bd"
	cookieEncrKeyHexStr = "4040ed16d4352320b5a7f51e26443342d55a0f46be2acfe5ba694a123230376a"
	cookieName          = "qnckie" // "quicknotes cookie"
)

var (
	cookieAuthKey []byte
	cookieEncrKey []byte

	secureCookie *securecookie.SecureCookie

	// random string for oauth2 API calls to protect against CSRF
	oauthSecretString = "5576867039"

	githubEndpoint = oauth2.Endpoint{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	}

	oauthGitHubConf = &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		// select level of access you want https://developer.github.com/v3/oauth/#scopes
		Scopes:   []string{"user:email", "repo"},
		Endpoint: githubEndpoint,
	}

	oauthTwitterClient = oauth.Client{
		TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
		TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
		ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authenticate",
		Credentials: oauth.Credentials{
			Secret: "wdDXapzG5zEeQ7ToJnABvBIoGmLFLdvueT7vGMPjCvjUNAU928",
			Token:  "rYmWoMXQ3Wwx69do31TW4DRes",
		},
	}

	secretsMutex sync.Mutex
	secrets      = map[string]string{}
)

// SecureCookieValue is value of the cookie
type SecureCookieValue struct {
	UserID int
}

func initCookieMust() {
	var err error
	cookieAuthKey, err = hex.DecodeString(cookieAuthKeyHexStr)
	u.PanicIfErr(err)
	cookieEncrKey, err = hex.DecodeString(cookieEncrKeyHexStr)
	u.PanicIfErr(err)
	secureCookie = securecookie.New(cookieAuthKey, cookieEncrKey)
	// verify auth/encr keys are correct
	val := map[string]string{
		"foo": "bar",
	}
	_, err = secureCookie.Encode(cookieName, val)
	u.PanicIfErr(err)
}

func setSecureCookie(w http.ResponseWriter, cookieVal *SecureCookieValue) {
	val, err := json.Marshal(cookieVal)
	if err != nil {
		LogErrorf("json.Marshal(%#v) failed with %s\n", cookieVal, err)
		return
	}

	if encoded, err := secureCookie.Encode(cookieName, val); err == nil {
		// TODO: set expiration (Expires    time.Time) long time in the future?
		cookie := &http.Cookie{
			Name:  cookieName,
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
	} else {
		fmt.Printf("setSecureCookie(): error encoding secure cookie %s\n", err)
	}
}

const weekInSeconds = 60 * 60 * 24 * 7

// to delete the cookie value (e.g. for logging out), we need to set an
// invalid value
func deleteSecureCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   cookieName,
		Value:  "deleted",
		MaxAge: weekInSeconds,
		Path:   "/",
	}
	http.SetCookie(w, cookie)
}

func getSecureCookie(r *http.Request, w http.ResponseWriter) *SecureCookieValue {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil
	}
	// detect a deleted cookie
	if "deleted" == cookie.Value {
		return nil
	}
	var ret SecureCookieValue
	if err = secureCookie.Decode(cookieName, cookie.Value, &ret); err != nil {
		// most likely expired cookie, so ignore and delete
		LogErrorf("secureCookie.Decode() failed with %s\n", err)
		deleteSecureCookie(w)
		return nil
	}
	fmt.Printf("Got cookie %#v\n", ret)
	return &ret
}

func getUserFromCookie(r *http.Request, w http.ResponseWriter) *DbUser {
	sc := getSecureCookie(r, w)
	if sc == nil {
		return nil
	}
	user, err := dbGetUserByID(sc.UserID)
	if err != nil {
		LogErrorf("dbGetUserById(%d) failed with %s\n", sc.UserID, err)
		return nil
	}
	return user
}

func putTempCredentials(cred *oauth.Credentials) {
	secretsMutex.Lock()
	defer secretsMutex.Unlock()
	secrets[cred.Token] = cred.Secret
}

func getTempCredentials(token string) *oauth.Credentials {
	secretsMutex.Lock()
	defer secretsMutex.Unlock()
	if secret, ok := secrets[token]; ok {
		return &oauth.Credentials{Token: token, Secret: secret}
	}
	return nil
}

func deleteTempCredentials(token string) {
	secretsMutex.Lock()
	defer secretsMutex.Unlock()
	delete(secrets, token)
}

// getTwitter gets a resource from the Twitter API and decodes the json response to data.
func getTwitter(cred *oauth.Credentials, urlStr string, params url.Values, data interface{}) error {
	if params == nil {
		params = make(url.Values)
	}
	oauthTwitterClient.SignParam(cred, "GET", urlStr, params)
	resp, err := http.Get(urlStr + "?" + params.Encode())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyData, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("GET %s returned status %d, %s", urlStr, resp.StatusCode, bodyData)
	}
	fmt.Printf("getTwitter(): json: %s\n", string(bodyData))
	return json.Unmarshal(bodyData, data)
}

// url: GET /logintwittercb?redirect=$redirect
func handleOauthTwitterCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("handleOauthTwitterCallback() url: '%s'\n", r.URL)
	tempCred := getTempCredentials(r.FormValue("oauth_token"))
	if tempCred == nil {
		http.Error(w, "Unknown oauth_token.", 500)
		return
	}
	deleteTempCredentials(tempCred.Token)
	tokenCred, _, err := oauthTwitterClient.RequestToken(nil, tempCred, r.FormValue("oauth_verifier"))
	if err != nil {
		http.Error(w, "Error getting request token, "+err.Error(), 500)
		return
	}
	putTempCredentials(tokenCred)
	fmt.Printf("tempCred: %#v\n", tempCred)
	fmt.Printf("tokenCred: %#v\n", tokenCred)

	var info map[string]interface{}
	uri := "https://api.twitter.com/1.1/account/verify_credentials.json"
	err = getTwitter(tokenCred, uri, nil, &info)
	if err != nil {
		http.Error(w, "Error getting timeline, "+err.Error(), 500)
		return
	}
	userHandle, okUser := info["screen_name"].(string)
	if !okUser {
		LogErrorf("no 'screen_name' in %#v\n", info)
		// TODO: show error to the user
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	fullName, _ := info["name"].(string)
	// also might be useful:
	// profile_image_url
	// profile_image_url_https
	user, err := dbGetOrCreateUser(userHandle, fullName)
	if err != nil {
		LogErrorf("dbGetOrCreateUser('%s', '%s') failed with '%s'\n", userHandle, fullName, err)
		// TODO: show error to the user
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	cookieVal := &SecureCookieValue{
		UserID: user.ID,
	}
	setSecureCookie(w, cookieVal)
	// TODO: dbUserSetTwitterOauth(user, tokenCredJson)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// url: GET /logintwitter?redirect=$redirect
func handleLoginTwitter(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("handleLoginTwitter() url: '%s'\n", r.URL)

	redirect := strings.TrimSpace(r.FormValue("redirect"))
	if redirect == "" {
		httpErrorf(w, "Missing redirect value for /logintwitter")
		return
	}

	q := url.Values{
		"redirect": {redirect},
	}.Encode()
	cb := "http://" + r.Host + "/logintwittercb?" + q

	tempCred, err := oauthTwitterClient.RequestTemporaryCredentials(nil, cb, nil)
	if err != nil {
		http.Error(w, "Error getting temp cred, "+err.Error(), 500)
		return
	}
	putTempCredentials(tempCred)
	http.Redirect(w, r, oauthTwitterClient.AuthorizationURL(tempCred, nil), 302)
}

// /logingithub?redirect=$redirect
func handleLoginGitHub(w http.ResponseWriter, r *http.Request) {
	redirect := strings.TrimSpace(r.FormValue("redirect"))
	if redirect == "" {
		httpErrorf(w, "Missing redirect value for /logintwitter")
		return
	}
	uri := oauthGitHubConf.AuthCodeURL(oauthSecretString, oauth2.AccessTypeOnline)
	http.Redirect(w, r, uri, http.StatusTemporaryRedirect)
}

// url: GET /logout?redirect=$redirect
func handleLogout(w http.ResponseWriter, r *http.Request) {
	redirect := strings.TrimSpace(r.FormValue("redirect"))
	if redirect == "" {
		httpErrorf(w, "Missing redirect value for /logout")
		return
	}
	deleteSecureCookie(w)
	http.Redirect(w, r, redirect, 302)
}
