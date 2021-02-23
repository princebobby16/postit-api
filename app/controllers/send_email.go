package controllers

import (
	"encoding/json"
	"github.com/twinj/uuid"
	"io/ioutil"
	"net/http"
	"postit-backend-api/pkg"
	"postit-backend-api/pkg/logs"
	"time"
)

func EmailNotificationService(w http.ResponseWriter, r *http.Request) {
	var emailRequest pkg.EmailRequest

	var transactionId = uuid.NewV4().String()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logs.Logger.Error(err)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        uuid.NewV4().String(),
				UiMessage: "Unable to send email! Contact Admin!",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId,
				TraceId:       "",
				Status:        "FAILED",
			},
		})
		return
	}
	logs.Logger.Info(string(body))

	err = json.Unmarshal(body, &emailRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logs.Logger.Error(err)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        uuid.NewV4().String(),
				UiMessage: "Unable to send email! Contact Admin!",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId,
				TraceId:       "",
				Status:        "FAILED",
			},
		})
		return
	}
	logs.Logger.Info(emailRequest)
	retry, err := pkg.SendEmail(emailRequest)
	if err != nil {
		logs.Logger.Error(err)
		if retry {
			_, err = pkg.SendEmail(emailRequest)
			w.WriteHeader(http.StatusInternalServerError)
			logs.Logger.Error(err)
			_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
				Data: pkg.Data{
					Id:        uuid.NewV4().String(),
					UiMessage: "Unable to send email! Contact Admin!",
				},
				Meta: pkg.Meta{
					Timestamp:     time.Now(),
					TransactionId: transactionId,
					TraceId:       "",
					Status:        "FAILED",
				},
			})
		}
		return
	}

	resp := pkg.StandardResponse{
		Data: pkg.Data{
			Id:        uuid.NewV4().String(),
			UiMessage: "Successfully sent message to shiftr. We'll get back to you soon",
		},
		Meta: pkg.Meta{
			Timestamp:     time.Now(),
			TransactionId: transactionId,
			TraceId:       r.Header.Get("trace-id"),
			Status:        "SUCCESS",
		},
	}

	_ = json.NewEncoder(w).Encode(resp)
}
