package router

import (
	"github.com/gorilla/mux"
	"gitlab.com/pbobby001/postit-api/app/controllers"
	"gitlab.com/pbobby001/postit-api/app/controllers/emojiList"
	"gitlab.com/pbobby001/postit-api/app/controllers/mediaupload"
	"gitlab.com/pbobby001/postit-api/app/controllers/posts"
	"gitlab.com/pbobby001/postit-api/app/controllers/social"
	"gitlab.com/pbobby001/postit-api/app/controllers/websockets"
	"net/http"
)

//Route Create a single route object
type Route struct {
	Name    string
	Path    string
	Method  string
	Handler http.HandlerFunc
}

//Routes Create an object of different routes
type Routes []Route

// InitRoutes Set up routes
func InitRoutes() *mux.Router {
	router := mux.NewRouter()

	routes := Routes{
		// health check
		Route{
			Name:    "Health Check",
			Path:    "/",
			Method:  http.MethodGet,
			Handler: controllers.HealthCheckHandler,
		},
		Route{
			Name:    "Email Notification Service",
			Path:    "/send-email",
			Method:  http.MethodPost,
			Handler: controllers.EmailNotificationService,
		},
		Route{
			Name:    "Delete Uploaded Files",
			Path:    "/delete/all",
			Method:  http.MethodDelete,
			Handler: mediaupload.DeleteUploadedFiles,
		},
		// posts
		Route{
			Name:    "Create Post",
			Path:    "/posts",
			Method:  http.MethodPost,
			Handler: posts.HandleCreatePost,
		},
		Route{
			Name:    "Fetch Posts",
			Path:    "/posts",
			Method:  http.MethodGet,
			Handler: posts.HandleFetchPosts,
		},
		Route{
			Name:    "Delete Post",
			Path:    "/posts",
			Method:  http.MethodDelete,
			Handler: posts.HandleDeletePost,
		},
		Route{
			Name:    "Update Post",
			Path:    "/posts",
			Method:  http.MethodPut,
			Handler: posts.HandleUpdatePost,
		},
		Route{
			Name:    "Batch delete",
			Path:    "/batch-delete",
			Method:  http.MethodPost,
			Handler: posts.HandleBatchDelete,
		},
		Route{
			Name:    "Batch Post",
			Path:    "/batch-post",
			Method:  http.MethodPost,
			Handler: posts.HandleBatchPost,
		},
		// schedule
		Route{
			Name:    "Create Post Schedule",
			Path:    "/schedule-post",
			Method:  http.MethodPost,
			Handler: posts.HandleCreatePostSchedule,
		},
		Route{
			Name:    "Fetch Post Schedule",
			Path:    "/schedule-post",
			Method:  http.MethodGet,
			Handler: posts.HandleFetchPostSchedule,
		},
		Route{
			Name:    "Update Post Schedule",
			Path:    "/schedule-post",
			Method:  http.MethodPut,
			Handler: posts.HandleUpdatePostSchedule,
		},
		Route{
			Name:    "Delete Post Schedule",
			Path:    "/schedule-post",
			Method:  http.MethodDelete,
			Handler: posts.HandleDeletePostSchedule,
		},
		// emojiList
		Route{
			Name:    "Get Emoji",
			Path:    "/emoji",
			Method:  http.MethodGet,
			Handler: emojiList.HandleGetEmoji,
		},
		//Get Facebook Code
		Route{
			Name:    "Get Facebook Code",
			Path:    "/fb/code",
			Method:  http.MethodPost,
			Handler: social.HandleFacebookCode,
		},
		Route{
			Name:    "Delete Facebook Code",
			Path:    "/fb/code",
			Method:  http.MethodDelete,
			Handler: social.HandleDeleteFacebookCode,
		},
		Route{
			Name:    "Fetch Facebook Code",
			Path:    "/all/code",
			Method:  http.MethodGet,
			Handler: social.AllAccounts,
		},

		// websockets
		Route{
			Path:    "/pws/schedule-status",
			Method:  http.MethodGet,
			Handler: websockets.ScheduleStatus,
		},
		//Route{
		//	Name: "Count Post",
		//	Path:    "/count/post",
		//	Method:  http.MethodGet,
		//	Handler: posts.CountPosts,
		//},
		Route{
			Name: "Count Schedule",
			Path:    "/count/data",
			Method:  http.MethodGet,
			Handler: posts.CountSchedule,
		},

		Route {
			Name: "Media Upload",
			Path: "/file/upload",
			Method: http.MethodPost,
			Handler: mediaupload.HandleMediaUpload,
		},

		Route{
			Name:    "Reverse Upload",
			Path:    "/file/upload",
			Method:  http.MethodDelete,
			Handler: mediaupload.HandleCancelMediaUpload,
		},

		//	Test
		Route{
			Name:    "Test Endpoint",
			Path:    "/tests",
			Method:  http.MethodGet,
			Handler: controllers.EndPointForTests,
		},
	}

	for _, route := range routes {
		router.Name(route.Name).
			Methods(route.Method).
			Path(route.Path).
			Handler(route.Handler)
	}

	return router
}
