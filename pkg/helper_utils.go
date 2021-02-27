package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cristalhq/jwt"
	"github.com/lib/pq"
	"github.com/twinj/uuid"
	"gitlab.com/pbobby001/postit-api/db"
	"gitlab.com/pbobby001/postit-api/pkg/logs"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/* Validate header is a function used to make sure that the required  headers are sent to the API
It takes the http request and extracts the headers from it and returns a map of the needed headers
and an error. Other headers are essentially ignored.*/
func ValidateHeaders(r *http.Request) (map[string]string, error) {
	//Group the headers
	receivedHeaders := make(map[string]string)
	requiredHeaders := []string{"trace-id", "tenant-namespace"}

	for _, header := range requiredHeaders {
		value := r.Header.Get(header)
		if value != "" {
			receivedHeaders[header] = value
		} else if value == "" {
			return nil, errors.New("Required header: " + header + " not found")
		} else {
			return nil, errors.New("No headers received be sure to send some headers")
		}
	}

	return receivedHeaders, nil
}

type smtpServer struct {
	host string
	port string
}

// Address URI to smtp server
func (s *smtpServer) Address() string {
	return s.host + ":" + s.port
}

/* Helper function to send the email to shiftrgh@gmail.com */
func SendEmail(req EmailRequest) (bool, error) {
	from := "princebobby506@gmail.com"
	password := "yoforreal.com"

	to := []string{
		"shiftrgh@gmail.com",
	}
	// smtp server configuration.
	smtpServer := smtpServer{host: "smtp.gmail.com", port: "587"}

	emailBody, err := parseTemplate("index.html", req)
	if err != nil {
		return false, err
	}
	logs.Logger.Info(emailBody)

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + "Message From Shiftr Gh Website" + "!\n"
	headers := []byte(subject + mime + "\n" + string(emailBody))
	var body bytes.Buffer

	body.Write(headers)

	//m := gomail.NewMessage()
	//m.SetHeader("From", from)
	//m.SetHeader("To", to[1])
	//m.SetHeader("Subject", subject+mime)
	//m.SetBody()

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpServer.host)

	retry := false
	// Sending email.
	err = smtp.SendMail(smtpServer.Address(), auth, from, to, body.Bytes())
	if err != nil {
		retry = true
		return retry, err
	}

	return retry, err
}

func parseTemplate(s string, req EmailRequest) ([]byte, error) {

	path, err := filepath.Abs(fmt.Sprintf("cmd/postit/pkg/12/%s", s))
	if err != nil {
		return nil, err
	}
	logs.Logger.Info(path)

	t, err := template.ParseFiles(path)
	if err != nil {
		return nil, nil
	}
	logs.Logger.Info(t.Name())

	buff := new(bytes.Buffer)
	err = t.Execute(buff, req)
	if err != nil {
		return nil, nil
	}

	return buff.Bytes(), nil
}

/* Helper function to append the hash tag symbol['#'] to the various hash tag strings
It takes a slice of strings and returns a new slice with the hash tag symbol['#'] attached to it.*/
func GenerateHashTags(hashT []string) []string {
	var hashTags []string

	for _, b := range hashT {
		trimmedString := strings.TrimLeft(b, "#")
		val := fmt.Sprintf("#%s", trimmedString)
		hashTags = append(hashTags, val)
	}

	return hashTags
}

/* Helper function to create post */
func CreatePost(post Post, tenantNamespace string, postId uuid.UUID) error {
	query := fmt.Sprintf("INSERT INTO %s.post (post_id, post_message, post_image, image_extension, hash_tags, post_priority, post_status) VALUES ($1, $2, $3, $4, $5, $6, $7)", tenantNamespace)
	if post.PostImage == nil {
		_, err := db.Connection.Exec(query, postId.String(), &post.PostMessage, []byte{}, "", pq.Array(&post.HashTags), &post.PostPriority, &post.PostStatus)
		if err != nil {
			return err
		}
	} else if post.PostImage != nil {
		imageExtension := ""
		switch http.DetectContentType(post.PostImage) {
		case "image/webp":
			imageExtension = ".webp"
		case "image/jpeg":
			imageExtension = ".jpeg"
		case "image/png":
			imageExtension = ".png"
		default:
			imageExtension = ".jpg"
		}

		_, err := db.Connection.Exec(query, postId.String(), &post.PostMessage, &post.PostImage, imageExtension, pq.Array(&post.HashTags), &post.PostPriority, &post.PostStatus)
		if err != nil {
			// db error
			return err
		}
	}
	return nil
}

/* Helper function to generate the duration for each post */
func GenerateDurationForEachPost(schedule PostSchedule) float64 {
	totalDuration := schedule.To.Sub(schedule.From)
	log.Println(schedule.To.Sub(schedule.From))

	numberOfPosts := len(schedule.PostIds)

	durationPerPost := totalDuration.Seconds() / float64(numberOfPosts)

	log.Println(durationPerPost)
	return durationPerPost
}

/* Helper func to handle error */
func SendErrorResponse(w http.ResponseWriter, tId uuid.UUID, traceId string, err error, httpStatus int) {
	w.WriteHeader(httpStatus)
	logs.Logger.Error(err)
	_ = json.NewEncoder(w).Encode(StandardResponse {
		Data: Data{
			Id:        "",
			UiMessage: "Something went wrong! Contact Admin!",
		},
		Meta: Meta{
			Timestamp:     time.Now(),
			TransactionId: tId.String(),
			TraceId:       traceId,
			Status:        "FAILED",
		},
	})
	return
}

func WebSocketTokenValidateToken(tokenString string, PrivateKey []byte, tenantNamespace string) error {
	logs.Logger.Info(tokenString)
	logs.Logger.Info(tenantNamespace)

	verifier, err := jwt.NewVerifierHS(jwt.HS512, PrivateKey)
	if err != nil {
		return err
	}

	jwtToken, err := jwt.ParseString(tokenString, verifier)
	if err != nil {
		return err
	}

	var jwtClaims jwt.RegisteredClaims
	claims := jwtToken.RawClaims()
	err = json.Unmarshal(claims, &jwtClaims)
	if err != nil {
		return err
	}

	//var newToken string
	if jwtClaims.ExpiresAt.Before(time.Now()) {
		logs.Logger.Info("Token expired! getting a new one....")

		client := &http.Client{}
		req, err := http.NewRequest(http.MethodPost, os.Getenv("AUTHENTICATION_SERVER_URL")+"/refresh-token", nil)
		if err != nil {
			return err
		}

		req.Header.Set("token", tokenString)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		logs.Logger.Info("refresh-token: ", resp.Header.Get("refresh-token"))
	}

	if jwtClaims.Audience[0] != tenantNamespace {
		return errors.New("invalid tenant namespace")
	}

	err = verifier.Verify(jwtToken.Payload(), jwtToken.Signature())
	if err != nil {
		return err
	}

	return nil
}