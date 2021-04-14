package social

import (
	"encoding/json"
	"fmt"
	"github.com/twinj/uuid"
	"gitlab.com/pbobby001/postit-api/db"
	"gitlab.com/pbobby001/postit-api/pkg"
	"gitlab.com/pbobby001/postit-api/pkg/logs"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func HandleFacebookCode(w http.ResponseWriter, r *http.Request) {
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
	logs.Logger.Info("Headers => TraceId: " + traceId + ", TenantNamespace: " + tenantNamespace)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}
	logs.Logger.Info(string(body))

	var requestBody pkg.FacebookCode

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}
	logs.Logger.Info(requestBody)

	appUUid := uuid.NewV4()
	appId := os.Getenv("FACEBOOK_APP_ID")
	appSecret := os.Getenv("FACEBOOK_APP_SECRET")
	appUrl := os.Getenv("FACEBOOK_APP_URL")

	// use code to get access token
	var client http.Client
	url := fmt.Sprintf("https://graph.facebook.com/oauth/access_token?client_id=%s&redirect_uri=%s&client_secret=%s&code=%s",
		appId,
		appUrl,
		appSecret,
		requestBody.Code,
	)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}

	response, err := client.Do(request)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}

	b, _ := ioutil.ReadAll(response.Body)
	logs.Logger.Info(string(b))

	var shortLivedFbAccessToken pkg.AuthResponse
	err = json.Unmarshal(b, &shortLivedFbAccessToken)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}
	logs.Logger.Info(response.Header)
	logs.Logger.Info(response.Status)
	logs.Logger.Info(response.StatusCode)
	logs.Logger.Info(shortLivedFbAccessToken.AccessToken)
	logs.Logger.Info(shortLivedFbAccessToken.ExpiresIn)
	logs.Logger.Info(shortLivedFbAccessToken.TokenType)

	// use short lived access token to get a long lived one
	url = fmt.Sprintf("https://graph.facebook.com/oauth/access_token?grant_type=fb_exchange_token&client_id=%s&client_secret=%s&fb_exchange_token=%s",
		appId,
		appSecret,
		shortLivedFbAccessToken.AccessToken,
	)
	response, err = http.Get(url)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}
	logs.Logger.Info(string(body))

	var longLivedFbAccessToken pkg.AuthResponse
	err = json.Unmarshal(body, &longLivedFbAccessToken)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}

	url = fmt.Sprintf("https://graph.facebook.com/me?access_token=%s", longLivedFbAccessToken.AccessToken)
	response, err = http.Get(url)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}
	logs.Logger.Info(string(body))

	var fbUser pkg.FacebookUserData
	err = json.Unmarshal(body, &fbUser)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}

	//Store inside the db
	stmt := fmt.Sprintf("INSERT INTO %s.application_info(application_uuid, application_name, application_id, application_secret, application_url, user_access_token, expires_in, user_name, user_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)", tenantNamespace)
	logs.Logger.Info("query", stmt)
	lId, err := db.Connection.Exec(stmt,
		&appUUid,
		"facebook",
		&appId,
		&appSecret,
		&appUrl,
		&longLivedFbAccessToken.AccessToken,
		&longLivedFbAccessToken.ExpiresIn,
		&fbUser.Name,
		&fbUser.Id,
	)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusInternalServerError)
		return
	}
	i, _ := lId.LastInsertId()
	logs.Logger.Info("Last insert Id: ", i)

	_ = json.NewEncoder(w).Encode(struct {
		Message string   `json:"message"`
		Meta    pkg.Meta `json:"meta"`
	}{
		Message: "Stored Code",
		Meta: pkg.Meta{
			Timestamp:     time.Now(),
			TransactionId: transactionId.String(),
			TraceId:       traceId,
			Status:        "SUCCESS",
		},
	})
}

func HandleDeleteFacebookCode(w http.ResponseWriter, r *http.Request) {

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
	logs.Logger.Info("Headers => TraceId: " + traceId + ", TenantNamespace: " + tenantNamespace)

	appUuid := r.URL.Query().Get("app_id")
	if appUuid == "" {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}

	stmt := fmt.Sprintf("DELETE FROM %s.application_info WHERE user_id = '%s'", tenantNamespace, appUuid)
	logs.Logger.Info(stmt)
	_, err = db.Connection.Exec(stmt)
	if err != nil {
		_ = logs.Logger.Error(err)
		return
	}

	_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
		Data: pkg.Data{
			Id:        "",
			UiMessage: "Code Deleted",
		},
		Meta: pkg.Meta{
			Timestamp:     time.Now(),
			TransactionId: transactionId.String(),
			TraceId:       traceId,
			Status:        "SUCCESS",
		},
	})
}

func FetchFacebookPosts(w http.ResponseWriter, r *http.Request) {
	logs.Logger.Info("===========================================")
	logs.Logger.Info("Handling Fetch Facebook Posts ...")
	logs.Logger.Info("===========================================")
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
	logs.Logger.Info("Headers => TraceId: " + traceId + ", TenantNamespace: " + tenantNamespace)

	query := fmt.Sprintf("SELECT post_id, facebook_post_id, facebook_user_id, post_message FROM %s.post WHERE post_fb_status = $1", tenantNamespace)
	rows, err := db.Connection.Query(query, true)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, "", err, http.StatusInternalServerError)
		return
	}

	var dbP pkg.FacebookPostData
	var dbPosts []pkg.FacebookPostData
	var comment pkg.Comment
	for rows.Next() {
		err := rows.Scan(
			&dbP.PostId,
			&dbP.FacebookPostId,
			&dbP.FacebookUserId,
			&dbP.PostMessage,
		)
		if err != nil {
			pkg.SendErrorResponse(w, transactionId, "", err, http.StatusInternalServerError)
			return
		}

		// get accessToken from db
		var accessToken string
		query := fmt.Sprintf("SELECT user_access_token FROM %s.application_info WHERE user_id = $1", tenantNamespace)
		err = db.Connection.QueryRow(query, dbP.FacebookUserId).Scan(&accessToken)
		if err != nil {
			pkg.SendErrorResponse(w, transactionId, "", err, http.StatusInternalServerError)
			return
		}
		// prepare request to facebook for the comments
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodGet, os.Getenv("FACEBOOK_COMMENTS_URL")+"/v10.0/"+dbP.FacebookPostId+"/comments?access_token="+accessToken, nil)
		if err != nil {
			logs.Logger.Info(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		logs.Logger.Info(req.URL.String())

		// make the request
		//req.URL.Query().Set("access_token", accessToken)
		logs.Logger.Info(dbP)
		logs.Logger.Info("fetching comments from facebook")
		resp, err := client.Do(req)
		if err != nil {
			logs.Logger.Info(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logs.Logger.Info(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		logs.Logger.Info(string(body))
		logs.Logger.Info(resp.Status)

		if resp.StatusCode != 200 {
			logs.Logger.Info(resp.StatusCode)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = json.Unmarshal(body, &comment)
		if err != nil {
			logs.Logger.Info(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logs.Logger.Info(comment)

		dbP.Comments = comment

		dbPosts = append(dbPosts, dbP)
	}

	if dbPosts == nil {
		dbPosts = []pkg.FacebookPostData{}
	} else {
		logs.Logger.Info(dbPosts[0].PostMessage)
	}

	//	If everything goes right build the response
	response := pkg.FetchFacebookPostResponse{
		Data: dbPosts,
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
