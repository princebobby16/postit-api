package middlewares

import (
	"encoding/json"
	"github.com/cristalhq/jwt"
	"gitlab.com/pbobby001/postit-api/pkg/logs"
	"io/ioutil"
	"net/http"
	"os"
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

			jwtToken, err := jwt.Parse([]byte(tokenString))
			if err != nil {
				logs.Logger.Info(err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Something went wrong! Contact Admin!"))
				return
			}

			var jwtClaims *jwt.StandardClaims
			claims := jwtToken.RawClaims()
			err = json.Unmarshal(claims, &jwtClaims)
			if err != nil {
				logs.Logger.Info(err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Something went wrong! Contact Admin!"))
				return
			}

			//var newToken string
			if jwtClaims.ExpiresAt.Time().Before(time.Now()) {
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

			if jwtClaims.Audience[1] != r.Header.Get("tenant-namespace") && jwtClaims.Audience[0] != "postit-audience" {
				logs.Logger.Info("Jwt Claim Audience", jwtClaims.Audience[0], jwtClaims.Audience[1])
				logs.Logger.Info("Invalid tenant-namespace")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("Wrong org namespace header"))
				return
			}

			validator := jwt.NewValidator(
				jwt.AudienceChecker([]string{"postit-audience", r.Header.Get("tenant-namespace")}),
			)

			err = validator.Validate(jwtClaims)
			if err != nil {
				logs.Logger.Info(err)
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("contact admin"))
				return
			}

			next.ServeHTTP(w, r)
		}
	})
}
