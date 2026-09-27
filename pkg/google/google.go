package googlepkg

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Profile struct {
	Sub           string
	Email         string
	Name          string
	Picture       string
	EmailVerified bool
}

// VerifyIDToken validates a Google ID token via Google's tokeninfo endpoint.
// Checks audience == our client ID, expiry, and verified email.
func VerifyIDToken(clientID, idToken string) (*Profile, error) {
	if clientID == "" {
		return nil, errors.New("google login belum dikonfigurasi (GOOGLE_CLIENT_ID kosong)")
	}
	if idToken == "" {
		return nil, errors.New("id_token required")
	}
	req, _ := http.NewRequest("GET", "https://oauth2.googleapis.com/tokeninfo?id_token="+url.QueryEscape(idToken), nil)
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, errors.New("gagal verifikasi token Google")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("token Google tidak valid")
	}
	var v struct {
		Aud           string `json:"aud"`
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Exp           string `json:"exp"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, errors.New("token Google tidak valid")
	}
	if v.Aud != clientID {
		return nil, errors.New("token Google bukan untuk aplikasi ini")
	}
	if exp, _ := strconv.ParseInt(v.Exp, 10, 64); exp > 0 && time.Now().Unix() > exp {
		return nil, errors.New("token Google kedaluwarsa")
	}
	if v.Email == "" || (v.EmailVerified != "true" && v.EmailVerified != "1") {
		return nil, errors.New("email Google belum terverifikasi")
	}
	return &Profile{Sub: v.Sub, Email: v.Email, Name: v.Name, Picture: v.Picture, EmailVerified: true}, nil
}
