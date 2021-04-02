package social

import (
	"encoding/json"
	"fmt"
	"github.com/twinj/uuid"
	"gitlab.com/pbobby001/postit-api/db"
	"gitlab.com/pbobby001/postit-api/pkg"
	"gitlab.com/pbobby001/postit-api/pkg/logs"
	"net/http"
	"time"
)

func AllAccounts(w http.ResponseWriter, r *http.Request) {
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
	logs.Logger.Infof("Headers => TraceId: %s TenantNamespace: %s", traceId, tenantNamespace)

	query := fmt.Sprintf("SELECT * FROM %s.application_info ORDER BY updated_at DESC LIMIT 2000", tenantNamespace)
	logs.Logger.Info(query)
	rows, err := db.Connection.Query(query)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
		return
	}

	var fb []pkg.FacebookPostitUserData
	var tw []pkg.TwitterPostitUserData
	var li []pkg.LinkedInPostitUserData
	for rows.Next() {
		var appInfo pkg.ApplicationInfo
		err = rows.Scan(
			&appInfo.ApplicationUuid,
			&appInfo.ApplicationName,
			&appInfo.ApplicationId,
			&appInfo.ApplicationSecret,
			&appInfo.ApplicationUrl,
			&appInfo.UserAccessToken,
			&appInfo.ExpiresIn,
			&appInfo.UserName,
			&appInfo.UserId,
			&appInfo.CreatedAt,
			&appInfo.UpdatedAt,
		)
		if err != nil {
			pkg.SendErrorResponse(w, transactionId, traceId, err, http.StatusBadRequest)
			return
		}

		if appInfo.ApplicationName == "facebook" {
			fb = append(fb, pkg.FacebookPostitUserData {
				Username:    appInfo.UserName,
				UserId:      appInfo.UserName,
				AccessToken: appInfo.UserAccessToken,
			})
		} else if appInfo.ApplicationName == "twitter" {
			tw = append(tw, pkg.TwitterPostitUserData {
				Username:    appInfo.UserName,
				UserId:      appInfo.UserName,
				AccessToken: appInfo.UserAccessToken,
			})
		} else if appInfo.ApplicationName == "linked_in" {
			li = append(li, pkg.LinkedInPostitUserData {
				Username:    appInfo.UserName,
				UserId:      appInfo.UserName,
				AccessToken: appInfo.UserAccessToken,
			})
		}

	}

	_ = json.NewEncoder(w).Encode(struct {
		Data pkg.PostitUserData `json:"data"`
		Meta pkg.Meta			`json:"meta"`
	}{
		Data: pkg.PostitUserData {
			FacebookPostitUserData: fb,
			TwitterPostitUserData:  tw,
			LinkedInPostitUserData: li,
		},
		Meta: pkg.Meta {
			Timestamp:     time.Now(),
			TransactionId: transactionId.String(),
			TraceId:       traceId,
			Status:        "SUCCESS",
		},
	})

}
