package posts

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gitlab.com/pbobby001/postit-api/pkg"
	"net/http/httptest"
	"testing"
)

func TestHandleCreatePost(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/posts", HandleCreatePost)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	body, err := json.Marshal(pkg.Post{
		PostMessage:  "something amazing is about to happen",
		HashTags:     []string{"something"},
		PostStatus:   false,
		PostPriority: false,
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(string(body))

	//resp, err := http.Post(testServer.URL + "/posts", "application/json", body)
}
