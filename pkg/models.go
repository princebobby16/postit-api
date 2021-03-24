package pkg

import (
	"github.com/gorilla/websocket"
	"time"
)

type (

	FacebookPostitUserData struct {
		Username string 			`json:"username"`
		UserId string 				`json:"user_id"`
		AccessToken string 			`json:"access_token"`
	}

	WebSocketHandShakeData struct {
		TenantNamespace string		`json:"tenant_namespace"`
		AuthToken string 			`json:"auth_token"`
	}

	CountResponse struct {
		PostCount int 			`json:"post_count"`
		ScheduleCount int 		`json:"schedule_count"`
		AccountCount int 		`json:"account_count"`
		Meta Meta 				`json:"meta"`
	}

	ScheduleStatus struct {
		ScheduleId string 		`json:"schedule_id"`
		ScheduleTitle string 	`json:"schedule_title"`
		From time.Time 			`json:"from"`
		To time.Time 			`json:"to"`
		TotalPost int 			`json:"total_post"`
		Posts []ScheduledPost    `json:"posts"`
		PostCount int          `json:"post_count"`
		CreatedAt time.Time    `json:"created_at"`
		UpdatedAt time.Time    `json:"updated_at"`
	}

	Client struct {
		Id string
		// Message controller
		Hub *MessageController
		// Websocket connection
		Conn *websocket.Conn
		// channel to send outgoing messages
		Send chan []byte
	}

	MessageController struct {
		// Registered Clients
		Clients map[*Client]bool
		// Channel for receiving incoming messages from client
		Broadcast chan []byte
		// Channel for registering requests from clients
		Register chan *Client
		// Channel for unregistering requests from clients
		Unregister chan *Client
	}

	EmailRequest struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phone_number"`
		Message     string `json:"message"`
	}

	LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	AuthResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	FacebookUserData struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}

	ScheduledPost struct {
		PostId       string    `json:"post_id"`
		PostMessage  string    `json:"post_message"`
		PostImages   [][]byte  `json:"post_image"`
		ImagePaths 	 []string  `json:"image_paths"`
		HashTags     []string  `json:"hash_tags"`
		PostStatus   bool      `json:"post_status"`
		PostPriority bool      `json:"post_priority"`
		CreatedOn    time.Time `json:"created_on"`
		UpdatedOn    time.Time `json:"updated_on"`
	}

	Post struct {
		PostId       string    `json:"post_id"`
		PostMessage  string    `json:"post_message"`
		PostImage    []byte    `json:"post_image"`
		HashTags     []string  `json:"hash_tags"`
		PostStatus   bool      `json:"post_status"`
		PostPriority bool      `json:"post_priority"`
		CreatedOn    time.Time `json:"created_on"`
		UpdatedOn    time.Time `json:"updated_on"`
	}

	SocialMediaProfiles struct {
		Facebook []string 			`json:"facebook"`
		Twitter []string 			`json:"twitter"`
		LinkedIn []string 			`json:"linked_in"`
	}

	PostSchedule struct {
		ScheduleId    string    `json:"schedule_id"`
		ScheduleTitle string    `json:"schedule_title"`
		PostToFeed    bool      `json:"post_to_feed"`
		From          time.Time `json:"from"`
		To            time.Time `json:"to"`
		PostIds       []string  `json:"post_ids"`
		Duration      float64   `json:"duration"`
		IsDue         bool      `json:"is_due"`
		Profiles SocialMediaProfiles `json:"profiles"`
		CreatedOn     time.Time `json:"created_on"`
		UpdatedOn     time.Time `json:"updated_on"`
	}

	FetchPostResponse struct {
		Data []DbPost `json:"data"`
		Meta Meta     `json:"meta"`
	}

	FetchSchedulePostResponse struct {
		Data []PostSchedule `json:"data"`
		Meta Meta           `json:"meta"`
	}

	StandardResponse struct {
		Data Data `json:"data"`
		Meta Meta `json:"meta"`
	}

	DbPost struct {
		PostId         string    `json:"post_id"`
		FacebookPostId string	 `json:"facebook_post_id"`
		PostMessage    string    `json:"post_message"`
		PostImages 	   [][]byte `json:"post_images"`
		ImagePaths     []string  `json:"image_paths"`
		HashTags       []string  `json:"hash_tags"`
		Scheduled  	   bool		`json:"scheduled"`
		PostStatus     bool      `json:"post_status"`
		PostPriority   bool      `json:"post_priority"`
		CreatedOn      time.Time `json:"created_on"`
		UpdatedOn      time.Time `json:"updated_on"`
	}

	FileTooBigResponse struct {
		Message string `json:"message"`
		Status  string `json:"status_code"`
		Meta    *Meta  `json:"meta"`
	}

	TokenResponse struct {
		Token string `json:"token"`
		Meta  Meta   `json:"meta"`
	}

	FacebookCode struct {
		Code string `json:"code"`
	}

	Data struct {
		Id        string `json:"id"`
		UiMessage string `json:"ui_message"`
	}

	Meta struct {
		Timestamp     time.Time `json:"timestamp"`
		TransactionId string    `json:"transaction_id"`
		TraceId       string    `json:"trace_id"`
		Status        string    `json:"status"`
	}

	BatchDeletePostRequest struct {
		PostIds []string `json:"post_ids"`
	}
)
