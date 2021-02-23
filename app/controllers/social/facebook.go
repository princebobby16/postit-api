package social

import (
	"encoding/json"
	"fmt"
	"github.com/twinj/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"postit-backend-api/db"
	"postit-backend-api/pkg"
	"postit-backend-api/pkg/logs"
	"time"
)

func HandleFacebookCode (w http.ResponseWriter, r *http.Request) {
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
	logs.Logger.Info("Headers => TraceId: "+ traceId +", TenantNamespace: "+tenantNamespace)

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
		"Facebook_App",
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
		Message string 		`json:"message"`
		Meta pkg.Meta 		`json:"meta"`
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

func HandleFetchFacebookEmailData (w http.ResponseWriter, r *http.Request) {
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
	logs.Logger.Info("Headers => TraceId: "+ traceId +", TenantNamespace: "+tenantNamespace)

	query := fmt.Sprintf("SELECT user_name, user_id, user_access_token FROM %s.application_info ORDER BY updated_at DESC LIMIT 2000", tenantNamespace)
	log.Println(query)
	rows, err := db.Connection.Query(query)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}

	var fbData []pkg.FacebookPostitUserData
	for rows.Next() {
		var fb pkg.FacebookPostitUserData
		err = rows.Scan(&fb.Username, &fb.UserId, &fb.AccessToken)
		if err != nil {
			pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
			return
		}

		fbData = append(fbData, fb)
	}

	if fbData == nil {
		fbData = []pkg.FacebookPostitUserData{}
	}

	resp := struct {
		Data []pkg.FacebookPostitUserData 			`json:"data"`
		Meta pkg.Meta								`json:"meta"`
	}{
		Data: fbData,
		Meta: pkg.Meta{
			Timestamp:     time.Now(),
			TransactionId: transactionId.String(),
			TraceId:       traceId,
			Status:        "SUCCESS",
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func HandleDeleteFacebookCode (w http.ResponseWriter, r *http.Request) {

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
	logs.Logger.Info("Headers => TraceId: "+ traceId +", TenantNamespace: "+tenantNamespace)

	appUuid := r.URL.Query().Get("app_id")
	if appUuid == "" {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}

	stmt := fmt.Sprintf("DELETE FROM %s.application_info WHERE user_id = '%s'", tenantNamespace, appUuid)
	logs.Logger.Info(stmt)
	_, err = db.Connection.Exec(stmt)
	if err != nil {
		logs.Logger.Error(err)
		return
	}

	_ = json.NewEncoder(w).Encode(struct {
		Message string 		`json:"message"`
		Meta pkg.Meta 		`json:"meta"`
	}{
		Message: "Code Deleted",
		Meta: pkg.Meta{
			Timestamp:     time.Now(),
			TransactionId: transactionId.String(),
			TraceId:       traceId,
			Status:        "SUCCESS",
		},
	})
}