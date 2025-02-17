package posts

import (
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"github.com/twinj/uuid"
	"gitlab.com/pbobby001/postit-api/db"
	"gitlab.com/pbobby001/postit-api/pkg"
	"gitlab.com/pbobby001/postit-api/pkg/logs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func HandleCreatePost(w http.ResponseWriter, r *http.Request) {
	logs.Logger.Info("===========================================")
	logs.Logger.Info("Handling create post ...")
	logs.Logger.Info("===========================================")

	/* TODO: store transaction info
	Create an id for this transaction */
	transactionId := uuid.NewV4()

	headers, err := pkg.ValidateHeaders(r)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, "", err, http.StatusBadRequest)
		return
	}

	//Get the relevant headers
	traceId := headers["trace-id"]
	tenantNamespace := headers["tenant-namespace"]

	// Logging the headers
	logs.Logger.Info("Headers => TraceId: %s, TenantNamespace: %s", traceId, tenantNamespace)

	// generate uuid for post
	id := uuid.NewV4()
	logs.Logger.Info(id)

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}

	logs.Logger.Info("Request Object: ", string(requestBody))

	// Create Post instance to decode request object into
	var post *pkg.Post

	// Decode request body into the Post struct
	err = json.Unmarshal(requestBody, &post)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		_ = logs.Logger.Error(err)
		return
	}

	path := filepath.Join(wd, "pkg/"+tenantNamespace)
	logs.Logger.Info(path)

	fileInfo, err := ioutil.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			_ = logs.Logger.Warn(err)
		} else {
			_ = logs.Logger.Error(err)
			return
		}
	}

	var imagePaths []string
	var images [][]byte
	var imageBytes []byte

	if fileInfo != nil {
		for _, file := range fileInfo {
			logs.Logger.Info(file.Name())

			if file.Name() == "f.json" {
				continue
			}

			fileLocation := filepath.Join(path, file.Name())

			openImage, err := os.Open(fileLocation)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}

			imageBytes, err = ioutil.ReadAll(openImage)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}
			err = openImage.Close()
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}
			images = append(images, imageBytes)

			jsonFile, err := os.Open(filepath.Join(path, "f.json"))
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}

			readFile, err := ioutil.ReadAll(jsonFile)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}
			err = jsonFile.Close()
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}

			fileData := make(map[string]string)
			err = json.Unmarshal(readFile, &fileData)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}

			imagePath := fileData[file.Name()]
			logs.Logger.Info(imagePath)
			imagePaths = append(imagePaths, imagePath)
		}
	}

	// imagePaths
	logs.Logger.Info(imagePaths)

	// Generate hashtag list
	hashTagList := pkg.GenerateHashTags(post.HashTags)
	logs.Logger.Info("HashTags: ", hashTagList)

	/* Replace post arrays with the new array list
	Totally unnecessary but I did it anyway */
	post.HashTags = hashTagList

	// TODO: Build and use a crud service
	//build query
	query := fmt.Sprintf("INSERT INTO %s.post (post_id, facebook_post_id, post_message, post_images, image_paths, hash_tags, post_fb_status, post_tw_status, post_li_status, post_priority, scheduled) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)", tenantNamespace)
	logs.Logger.Info("Db Query: ", query)

	result, err := db.Connection.Exec(query, id.String(), "", post.PostMessage, pq.Array(images), pq.Array(imagePaths), pq.Array(post.HashTags), false, false, false, post.PostPriority, false)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusInternalServerError)
		return
	}

	go func() {
		err = os.RemoveAll(path)
		if err != nil {
			_ = logs.Logger.Error(err)
			return
		}
	}()

	//Just to be sure data was inserted
	insertId, _ := result.LastInsertId()
	logs.Logger.Info("Last Insert Id: ", insertId)

	// Build response
	response := pkg.StandardResponse{
		Data: pkg.Data{Id: id.String(), UiMessage: "Post Created!"},
		Meta: pkg.Meta{Timestamp: time.Now(), TransactionId: transactionId.String(), TraceId: traceId, Status: "SUCCESS"},
	}

	// Send response message
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusInternalServerError)
		return
	}

}

func HandleFetchPosts(w http.ResponseWriter, r *http.Request) {
	logs.Logger.Info("===========================================")
	logs.Logger.Info("Handling Fetch post ...")
	logs.Logger.Info("===========================================")

	transactionId := uuid.NewV4()

	// TODO: Properly Validate headers
	headers, err := pkg.ValidateHeaders(r)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, "", err, http.StatusBadRequest)
		return
	}

	//Get the relevant headers
	traceId := headers["trace-id"]
	tenantNamespace := headers["tenant-namespace"]

	// Logging the headers
	logs.Logger.Infof("Headers => TraceId: %s, TenantNamespace: %s", traceId, tenantNamespace)

	// TODO: refactor fetch post to send images as well
	// Build the sql query
	query := fmt.Sprintf("SELECT post_id, facebook_post_id, post_message, post_images, image_paths, hash_tags, post_fb_status, post_tw_status, post_li_status, post_priority, scheduled, created_at, updated_at FROM %s.post ORDER BY updated_at DESC LIMIT 1000", tenantNamespace)
	logs.Logger.Info(query)

	// Run the query on the db using that particular db connection
	rows, err := db.Connection.Query(query)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusInternalServerError)
		return
	}

	// Create a post instance
	var postList []pkg.DbPost
	//loop through the posts
	//if rows.Next() {
	for rows.Next() {
		var post pkg.DbPost
		// Set the db values to the post values
		err = rows.Scan(
			&post.PostId,
			&post.FacebookPostId,
			&post.PostMessage,
			pq.Array(&post.PostImages),
			pq.Array(&post.ImagePaths),
			pq.Array(&post.HashTags),
			&post.PostFbStatus,
			&post.PostTwStatus,
			&post.PostLiStatus,
			&post.PostPriority,
			&post.Scheduled,
			&post.CreatedOn,
			&post.UpdatedOn,
		)
		if err != nil {
			pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusInternalServerError)
			return
		}

		if post.ImagePaths == nil || len(post.ImagePaths) == 0 {
			post.ImagePaths = []string{}
		}

		if post.PostImages == nil || len(post.PostImages) == 0 {
			post.PostImages = [][]byte{}
		}

		if post.HashTags == nil || len(post.HashTags) == 0 {
			post.HashTags = []string{}
		}
		//	Build the post data list
		postList = append(postList, post)
	}
	// Generate an id for this particular transaction
	if postList != nil {
		logs.Logger.Info(postList[0].PostMessage)
	}

	if postList == nil || len(postList) == 0 {
		postList = []pkg.DbPost{}
	}

	//	If everything goes right build the response
	response := pkg.FetchPostResponse{
		Data: postList,
		Meta: pkg.Meta{
			Timestamp:     time.Now(),
			TransactionId: transactionId.String(),
			TraceId:       traceId,
			Status:        "SUCCESS",
		},
	}

	w.WriteHeader(http.StatusFound)
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusInternalServerError)
		return
	}
}

func HandleUpdatePost(w http.ResponseWriter, r *http.Request) {
	logs.Logger.Info("===========================================")
	logs.Logger.Info("Handling Update Post ...")
	logs.Logger.Info("===========================================")

	// Generate an id for this particular transaction
	transactionId := uuid.NewV4()

	// TODO: Properly Validate headers
	headers, err := pkg.ValidateHeaders(r)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, "", err, http.StatusBadRequest)
		return
	}

	//Get the relevant headers
	traceId := headers["trace-id"]
	tenantNamespace := headers["tenant-namespace"]

	// Logging the headers
	logs.Logger.Info("Headers => TraceId: %s, TenantNamespace: %s", traceId, tenantNamespace)

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}

	logs.Logger.Info(string(requestBody))

	// Create Post instance to decode request object into
	var post *pkg.Post

	// Decode request body into the Post struct
	err = json.Unmarshal(requestBody, &post)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}

	// Generate hash tag list
	hashTagList := pkg.GenerateHashTags(post.HashTags)
	logs.Logger.Info(hashTagList)

	/* Replace post arrays with the new array list
	Totally unnecessary but I did it anyway */
	post.HashTags = hashTagList

	//	Get url param
	uPostId, err := uuid.Parse(r.URL.Query().Get("post_id"))
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}
	logs.Logger.Info(uPostId)

	go func() {
		if err = pkg.Update(tenantNamespace, err, uPostId, post); err != nil {
			_ = logs.Logger.Error(err)
			return
		}
	}()

	response := &pkg.StandardResponse{
		Data: pkg.Data{
			Id:        uPostId.String(),
			UiMessage: "Post Updated!",
		},
		Meta: pkg.Meta{
			Timestamp:     time.Now(),
			TransactionId: transactionId.String(),
			TraceId:       traceId,
			Status:        "SUCCESS",
		},
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func HandleDeletePost(w http.ResponseWriter, r *http.Request) {
	logs.Logger.Info("===========================================")
	logs.Logger.Info("Handling Delete Post ...")
	logs.Logger.Info("===========================================")

	// Generate an id for this particular transaction
	transactionId := uuid.NewV4()

	// TODO: Properly Validate headers
	headers, err := pkg.ValidateHeaders(r)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, "", err, http.StatusBadRequest)
		return
	}

	//Get the relevant headers
	traceId := headers["trace-id"]
	tenantNamespace := headers["tenant-namespace"]

	// Logging the headers
	logs.Logger.Info("Headers => TraceId: %s, TenantNamespace: %s", traceId, tenantNamespace)

	//	Get url param
	postId := r.URL.Query().Get("post_id")
	logs.Logger.Info(postId)

	// TODO: Fetch query from post
	query := fmt.Sprintf("DELETE FROM %s.post WHERE post_id = $1", tenantNamespace)
	logs.Logger.Info(query)

	val, err := db.Connection.Query(query, postId)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}

	arr, _ := val.Columns()
	logs.Logger.Info(arr)

	response := pkg.StandardResponse{
		Data: pkg.Data{
			Id:        postId,
			UiMessage: "Post Deleted!",
		},
		Meta: pkg.Meta{
			Timestamp:     time.Now(),
			TransactionId: transactionId.String(),
			TraceId:       traceId,
			Status:        "SUCCESS",
		},
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&response)
}

func HandleBatchDelete(w http.ResponseWriter, r *http.Request) {
	logs.Logger.Info("===========================================")
	logs.Logger.Info("Handling Batch Delete Post ...")
	logs.Logger.Info("===========================================")

	// Generate an id for this particular transaction
	transactionId := uuid.NewV4()

	// TODO: Properly Validate headers
	headers, err := pkg.ValidateHeaders(r)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, "", err, http.StatusBadRequest)
		return
	}

	//Get the relevant headers
	traceId := headers["trace-id"]
	tenantNamespace := headers["tenant-namespace"]

	// Logging the headers
	logs.Logger.Info("Headers => TraceId: %s, TenantNamespace: %s", traceId, tenantNamespace)

	// Read the request body
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}
	logs.Logger.Info(string(b))

	var request pkg.BatchDeletePostRequest
	err = json.Unmarshal(b, &request)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}

	logs.Logger.Info(request)

	// build the query
	query := fmt.Sprintf("DELETE FROM %s.post WHERE post_id = $1", tenantNamespace)
	//iterate over the post ids from the request
	for _, i := range request.PostIds {
		_, err = db.Connection.Exec(query, i)
		if err != nil {
			pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusInternalServerError)
			return
		}
	}

	response := pkg.StandardResponse{
		Data: pkg.Data{
			Id:        "",
			UiMessage: "Posts Deleted!",
		},
		Meta: pkg.Meta{
			Timestamp:     time.Now(),
			TransactionId: transactionId.String(),
			TraceId:       traceId,
			Status:        "SUCCESS",
		},
	}

	_ = json.NewEncoder(w).Encode(response)
}
