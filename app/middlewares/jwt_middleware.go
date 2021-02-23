package middlewares

import (
	"encoding/json"
	"github.com/cristalhq/jwt"
	"io/ioutil"
	"net/http"
	"os"
	"postit-backend-api/pkg/logs"
	"strings"
	"time"
)

var (
	PrivateKey []byte
)

func init() {
	data, err := ioutil.ReadFile("private.pem")
	if err != nil {
		logs.Logger.Info(err)
		return
	}
	PrivateKey = data
	logs.Logger.Info(string(PrivateKey))
}

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/pws/schedule-status" {
			logs.Logger.Info("Health Check")
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/send-email" || r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}
		header := r.Header.Get("Authorization")
		if header == "" {
			w.WriteHeader(http.StatusUnauthorized)
			logs.Logger.Info("Login required")
			return
		} else {
			authHeader := strings.Split(header, "Bearer ")
			tokenString := authHeader[1]

			logs.Logger.Info(tokenString)

			verifier, err := jwt.NewVerifierHS(jwt.HS512, PrivateKey)
			if err != nil {
				logs.Logger.Info(err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Something went wrong! Contact Admin!"))
				return
			}

			jwtToken, err := jwt.ParseString(tokenString, verifier)
			if err != nil {
				logs.Logger.Info(err)
				logs.Logger.Info("Unable to parse token! Token Malformed")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("Token\rMalformed"))
				return
			}

			var jwtClaims jwt.RegisteredClaims
			claims := jwtToken.RawClaims()
			err = json.Unmarshal(claims, &jwtClaims)
			if err != nil {
				logs.Logger.Info(err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Something went wrong! Contact admin!"))
				return
			}

			//var newToken string
			if jwtClaims.ExpiresAt.Before(time.Now()) {
				logs.Logger.Info("Token expired! getting a new one....")

				client := &http.Client{}
				req, err := http.NewRequest(http.MethodPost, os.Getenv("AUTHENTICATION_SERVER_URL")+"/refresh-token", nil)
				if err != nil {
					logs.Logger.Info(err)
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				req.Header.Set("token", tokenString)
				resp, err := client.Do(req)
				if err != nil {
					logs.Logger.Info(err)
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				logs.Logger.Info("refresh-token: ", resp.Header.Get("refresh-token"))
			}

			if jwtClaims.Audience[0] != r.Header.Get("tenant-namespace") {
				logs.Logger.Info("Jwt Claim Audience", jwtClaims.Audience[0])
				logs.Logger.Info("Invalid tenant-namespace")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("Wrong org namespace header"))
				return
			}

			err = verifier.Verify(jwtToken.Payload(), jwtToken.Signature())
			if err != nil {
				logs.Logger.Info(err)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Invalid Token Signature"))
				return
			}

			next.ServeHTTP(w, r)
		}
	})
}
