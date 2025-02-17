package posts

import (
	"bytes"
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
	"time"
)

func CountSchedule(w http.ResponseWriter, r *http.Request) {
	logs.Logger.Info("===========================================")
	logs.Logger.Info("Handling Count Data ...")
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

	query := fmt.Sprintf(
		"SELECT COUNT(post_id) FROM %s.post;",
		tenantNamespace,
	)

	row := db.Connection.QueryRow(query)
	if row.Err() != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, row.Err(), http.StatusBadRequest)
		return
	}
	var postCount int
	err = row.Scan(&postCount)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, row.Err(), http.StatusBadRequest)
		return
	}
	logs.Logger.Info(postCount)


	query = fmt.Sprintf(
		"SELECT COUNT(schedule_id) FROM %s.schedule;",
		tenantNamespace,
	)

	row = db.Connection.QueryRow(query)
	if row.Err() != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, row.Err(), http.StatusBadRequest)
		return
	}

	var scheduleCount int
	err = row.Scan(&scheduleCount)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, row.Err(), http.StatusBadRequest)
		return
	}
	logs.Logger.Info(scheduleCount)


	query = fmt.Sprintf(
		"SELECT COUNT(application_uuid) FROM %s.application_info;",
		tenantNamespace,
	)

	row = db.Connection.QueryRow(query)
	if row.Err() != nil {
		pkg.SendErrorResponse(w, transactionId, traceId, row.Err(), http.StatusBadRequest)
		return
	}

	var accountCount int
	err = row.Scan(&accountCount)
	if err != nil {
		pkg.SendErrorResponse(w, transactionId, "", row.Err(), http.StatusBadRequest)
		return
	}
	logs.Logger.Info(accountCount)

	resp := &pkg.CountResponse {
		PostCount: postCount,
		ScheduleCount: scheduleCount,
		AccountCount: accountCount,
		Meta: pkg.Meta {
			Timestamp:     time.Now(),
			TransactionId: transactionId.String(),
			TraceId:       traceId,
			Status:        "SUCCESS",
		},
	}

	_ = json.NewEncoder(w).Encode(&resp)
}

func HandleCreatePostSchedule(w http.ResponseWriter, r *http.Request) {
	logs.Logger.Info("===========================================")
	logs.Logger.Info("Handling Create Post Schedule ...")
	logs.Logger.Info("===========================================")

	transactionId := uuid.NewV4()
	logs.Logger.Info("Transaction Id: ", transactionId)

	headers, err := pkg.ValidateHeaders(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logs.Logger.Info(err)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        "",
				UiMessage: err.Error(),
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       "",
				Status:        "FAILED",
			},
		})
		return
	}

	//Get the relevant headers
	traceId := headers["trace-id"]
	tenantNamespace := headers["tenant-namespace"]

	// Logging the headers
	logs.Logger.Info("Headers => TraceId: %s, TenantNamespace: %s", traceId, tenantNamespace)

	//	Create an instance of the post schedule
	var postSchedule pkg.PostSchedule

	// Get the request byte slice
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logs.Logger.Info(err)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        "",
				UiMessage: "Something went wrong. Contact admin",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       traceId,
				Status:        "FAILED",
			},
		})
		return
	}

	logs.Logger.Info(string(requestBody))

	// Decode the byte slice into the postSchedule object
	err = json.Unmarshal(requestBody, &postSchedule)
	if err != nil {
		// TODO: send appropriate message
		w.WriteHeader(http.StatusInternalServerError)
		logs.Logger.Info(err)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        "",
				UiMessage: "Something went wrong. Contact admin",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       traceId,
				Status:        "FAILED",
			},
		})
		return
	}
	logs.Logger.Info(postSchedule)

	for _, postId := range postSchedule.PostIds {
		query := fmt.Sprintf("UPDATE %s.post SET scheduled = $1 WHERE post_id = $2", tenantNamespace)
		_, err = db.Connection.Exec(query, true, postId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logs.Logger.Info(err)
			_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
				Data: pkg.Data{
					Id:        "",
					UiMessage: "Something went wrong. Contact admin",
				},
				Meta: pkg.Meta{
					Timestamp:     time.Now(),
					TransactionId: transactionId.String(),
					TraceId:       traceId,
					Status:        "FAILED",
				},
			})
			return
		}
	}

	durationPerPostInSeconds := pkg.GenerateDurationForEachPost(postSchedule)
	logs.Logger.Info(durationPerPostInSeconds)

	// Generate an id for the post schedule
	postScheduleId := uuid.NewV4()
	logs.Logger.Info(postScheduleId)
	postSchedule.ScheduleId = postScheduleId.String()
	logs.Logger.Info("Post Schedule Id: ", postSchedule.ScheduleId)

	postSchedule.Duration = durationPerPostInSeconds
	logs.Logger.Info("Post Duration: ", postSchedule.Duration)

	if postSchedule.Profiles.Facebook == nil {
		postSchedule.Profiles.Facebook = []string{}
	}

	if postSchedule.Profiles.Twitter == nil {
		postSchedule.Profiles.Twitter = []string{}
	}

	if postSchedule.Profiles.LinkedIn == nil {
		postSchedule.Profiles.LinkedIn = []string{}
	}

	// TODO: Build and use a crud service
	//build query
	query := fmt.Sprintf(
		"INSERT INTO %s.schedule (schedule_id, schedule_title, post_to_feed, schedule_from, schedule_to, post_ids, facebook, twitter, linked_in, duration_per_post, is_due) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)", tenantNamespace)

	result, err := db.Connection.Exec(
		query,
		postScheduleId.String(),
		postSchedule.ScheduleTitle,
		postSchedule.PostToFeed,
		postSchedule.From,
		postSchedule.To,
		pq.Array(postSchedule.PostIds),
		pq.Array(postSchedule.Profiles.Facebook),
		pq.Array(postSchedule.Profiles.Twitter),
		pq.Array(postSchedule.Profiles.LinkedIn),
		durationPerPostInSeconds,
		false,
	)
	if err != nil {
		// TODO: Send appropriate error message
		logs.Logger.Info(err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        postScheduleId.String(),
				UiMessage: "Something went wrong. Contact admin",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       traceId,
				Status:        "FAILED",
			},
		})
		return
	}

	// Just to be sure data was inserted
	insertId, _ := result.LastInsertId()
	logs.Logger.Info(insertId)

	// notify the scheduler micro service
	reqBody, _ := json.Marshal(postSchedule)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, os.Getenv("SCHEDULER_URL")+"/schedule", bytes.NewBuffer(reqBody))
	if err != nil {
		logs.Logger.Info(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	req.Header.Add("tenant-namespace", tenantNamespace)
	req.Header.Add("trace-id", traceId)
	resp, err := client.Do(req)
	if err != nil {
		// rollback migrations
		query = fmt.Sprintf("DELETE FROM %s.schedule WHERE schedule_id = $1", tenantNamespace)
		_, err = db.Connection.Exec(query, postScheduleId)
		if err != nil {
			logs.Logger.Info(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, postId := range postSchedule.PostIds {
			query := fmt.Sprintf("UPDATE %s.post SET scheduled = $1 WHERE post_id = $2", tenantNamespace)
			_, err = db.Connection.Exec(query, false, postId)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				logs.Logger.Info(err)
				_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
					Data: pkg.Data{
						Id:        "",
						UiMessage: "Something went wrong. Contact admin",
					},
					Meta: pkg.Meta{
						Timestamp:     time.Now(),
						TransactionId: transactionId.String(),
						TraceId:       traceId,
						Status:        "FAILED",
					},
				})
				return
			}
		}

		logs.Logger.Info(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		logs.Logger.Info(err)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        "",
				UiMessage: "Something went wrong. Contact admin",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       traceId,
				Status:        "FAILED",
			},
		})
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	logs.Logger.Info(string(body))

	if resp.StatusCode != http.StatusOK {
		_ = logs.Logger.Error(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	for i := 0; i < len(postSchedule.PostIds); i++ {
		// change scheduled status of posts
		query = fmt.Sprintf(`UPDATE %s.post SET scheduled = $1 WHERE post_id = $2;`, tenantNamespace)
		_, err = db.Connection.Exec(query, true, postSchedule.PostIds[i])
		if err != nil {
			_ = logs.Logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Build response
	response := pkg.StandardResponse{
		Data: pkg.Data{Id: postScheduleId.String(), UiMessage: "Schedule Created!"},
		Meta: pkg.Meta{Timestamp: time.Now(), TransactionId: transactionId.String(), TraceId: traceId, Status: "SUCCESS"},
	}

	// Send response message
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		logs.Logger.Info(err)
		return
	}

}

func HandleFetchPostSchedule(w http.ResponseWriter, r *http.Request) {
	logs.Logger.Info("===========================================")
	logs.Logger.Info("Handling Fetch Post Schedule...")
	logs.Logger.Info("===========================================")

	// Generate an id for this particular transaction
	transactionId := uuid.NewV4()
	logs.Logger.Info(transactionId)

	headers, err := pkg.ValidateHeaders(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = logs.Logger.Error(err)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        "",
				UiMessage: err.Error(),
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       "",
				Status:        "FAILED",
			},
		})
		return
	}

	//Get the relevant headers
	traceId := headers["trace-id"]
	tenantNamespace := headers["tenant-namespace"]

	// Logging the headers
	logs.Logger.Info(headers)

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		_ = logs.Logger.Error("Invalid method")
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        "",
				UiMessage: "Something went wrong. Contact admin",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       traceId,
				Status:        "FAILED",
			},
		})
		return
	}

	// Build the sql query
	query := fmt.Sprintf("SELECT * FROM %s.schedule ORDER BY updated_at DESC LIMIT 200", tenantNamespace)
	logs.Logger.Info(query)

	// Run the query on the db using that particular db connection
	rows, err := db.Connection.Query(query)
	if err != nil {
		// TODO: send appropriate error message
		w.WriteHeader(http.StatusInternalServerError)
		_ = logs.Logger.Error(err)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        "",
				UiMessage: "Something went wrong. Contact admin",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       traceId,
				Status:        "FAILED",
			},
		})
		return
	}

	// Create a post schedule instance
	var schedules []pkg.PostSchedule
	//loop through the schedules
	for rows.Next() {
		// Set the db values to the post schedules values
		var schedule pkg.PostSchedule
		err = rows.Scan(
			&schedule.ScheduleId,
			&schedule.ScheduleTitle,
			&schedule.PostToFeed,
			&schedule.From,
			&schedule.To,
			pq.Array(&schedule.PostIds),
			&schedule.Duration,
			pq.Array(&schedule.Profiles.Facebook),
			pq.Array(&schedule.Profiles.Twitter),
			pq.Array(&schedule.Profiles.LinkedIn),
			&schedule.IsDue,
			&schedule.CreatedOn,
			&schedule.UpdatedOn,
		)
		if err != nil {
			// TODO: Send an appropriate error message
			w.WriteHeader(http.StatusInternalServerError)
			_ = logs.Logger.Error(err)
			_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
				Data: pkg.Data{
					Id:        "",
					UiMessage: "Something went wrong. Contact admin",
				},
				Meta: pkg.Meta{
					Timestamp:     time.Now(),
					TransactionId: transactionId.String(),
					TraceId:       traceId,
					Status:        "FAILED",
				},
			})
			return
		}
		//	Build the post data list
		schedules = append(schedules, schedule)
	}

	//	If everything goes right build the response
	response := pkg.FetchSchedulePostResponse {
		Data: schedules,
		Meta: pkg.Meta{
			Timestamp:     time.Now(),
			TransactionId: transactionId.String(),
			TraceId:       traceId,
			Status:        "SUCCESS",
		},
	}
	logs.Logger.Info(response)

	w.WriteHeader(http.StatusFound)
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logs.Logger.Info(err)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        "",
				UiMessage: "Something went wrong. Contact admin",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       traceId,
				Status:        "FAILED",
			},
		})
		return
	}
}

func HandleUpdatePostSchedule(w http.ResponseWriter, r *http.Request) {
	logs.Logger.Info("===========================================")
	logs.Logger.Info("Handling Update Post Schedule ...")
	logs.Logger.Info("===========================================")

	// Todo: Create an after update trigger to update the posts in the scheduled_post table

	// Generate an id for this particular transaction
	transactionId := uuid.NewV4()

	headers, err := pkg.ValidateHeaders(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logs.Logger.Info(err)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        "",
				UiMessage: err.Error(),
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       "",
				Status:        "FAILED",
			},
		})
		return
	}

	//Get the relevant headers
	traceId := headers["trace-id"]
	tenantNamespace := headers["tenant-namespace"]

	// Logging the headers
	logs.Logger.Info(headers)

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// TODO: Send appropriate error message
		w.WriteHeader(http.StatusInternalServerError)
		logs.Logger.Info(err)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        "",
				UiMessage: "Something went wrong. Contact admin",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       traceId,
				Status:        "FAILED",
			},
		})
		return
	}

	logs.Logger.Info(string(requestBody))

	// Create Post instance to decode request object into
	var post *pkg.PostSchedule
	// Decode request body into the Post struct
	err = json.Unmarshal(requestBody, &post)
	if err != nil {
		// TODO: Send appropriate error message
		w.WriteHeader(http.StatusInternalServerError)
		logs.Logger.Info(err)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        "",
				UiMessage: "Something went wrong. Contact admin",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       traceId,
				Status:        "FAILED",
			},
		})
		return
	}

	//	Get url param
	postScheduleId := r.URL.Query().Get("schedule_id")
	logs.Logger.Info(postScheduleId)
	uPostId, err := uuid.Parse(postScheduleId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logs.Logger.Info(err)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        "",
				UiMessage: "Something went wrong. Contact admin",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       traceId,
				Status:        "FAILED",
			},
		})
		return
	}

	logs.Logger.Info(uPostId)

	//TODO: Validate post uuid
	query := fmt.Sprintf("UPDATE %s.schedule SET schedule_title = $1, schedule_from = $2, schedule_to = $3, post_ids = $4, post_to_feed = $5, facebook = $6, twitter = $7, linked_in = $8 WHERE schedule_id = $9", tenantNamespace)
	logs.Logger.Info(query)

	if post.Profiles.Facebook == nil {
		post.Profiles.Facebook = []string{}
	}

	if post.Profiles.Twitter == nil {
		post.Profiles.Twitter = []string{}
	}

	if post.Profiles.LinkedIn == nil {
		post.Profiles.LinkedIn = []string{}
	}

	_, err = db.Connection.Exec(
		query,
		post.ScheduleTitle,
		post.From,
		post.To,
		pq.Array(post.PostIds),
		post.PostToFeed,
		pq.Array(post.Profiles.Facebook),
		pq.Array(post.Profiles.Twitter),
		pq.Array(post.Profiles.LinkedIn),
		uPostId,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse {
			Data: pkg.Data{
				Id:        postScheduleId,
				UiMessage: "Unable to Update post schedule",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       traceId,
				Status:        "FAILED",
			},
		})
		logs.Logger.Info(err)
		return
	}

	response := &pkg.StandardResponse{
		Data: pkg.Data{
			Id:        postScheduleId,
			UiMessage: "Post Schedule Updated!",
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

func HandleDeletePostSchedule(w http.ResponseWriter, r *http.Request) {
	logs.Logger.Info("===========================================")
	logs.Logger.Info("Handling delete post schedule...")
	logs.Logger.Info("===========================================")

	// Generate an id for this particular transaction
	transactionId := uuid.NewV4()

	//get headers
	headers, err := pkg.ValidateHeaders(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logs.Logger.Info(err)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        "",
				UiMessage: err.Error(),
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       "",
				Status:        "FAILED",
			},
		})
		return
	}

	//Get the relevant headers
	traceId := headers["trace-id"]
	tenantNamespace := headers["tenant-namespace"]

	// Logging the headers
	logs.Logger.Info("Headers => TraceId: %s, TenantNamespace: %s", traceId, tenantNamespace)

	//	Get url param
	postScheduleId := r.URL.Query().Get("schedule_id")
	logs.Logger.Info(postScheduleId)

	scheduleId, err := uuid.Parse(postScheduleId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(struct {
			Data pkg.Data `json:"data"`
			Meta pkg.Meta `json:"meta"`
		}{
			Data: pkg.Data{
				Id:        "",
				UiMessage: "Bad Request. Try again!",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       traceId,
				Status:        "FAILED",
			},
		})
		logs.Logger.Info(err)
		return
	}

	// TODO: Fetch query from post
	query := fmt.Sprintf("DELETE FROM %s.schedule WHERE schedule_id = $1", tenantNamespace)
	logs.Logger.Info(query)

	val, err := db.Connection.Exec(query, scheduleId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(struct {
			Data pkg.Data `json:"data"`
			Meta pkg.Meta `json:"meta"`
		}{
			Data: pkg.Data{
				Id:        scheduleId.String(),
				UiMessage: "Unable to delete post schedule",
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       traceId,
				Status:        "FAILED",
			},
		})
		logs.Logger.Info(err)
		return
	}

	arr, _ := val.LastInsertId()
	logs.Logger.Info(arr)

	response := pkg.StandardResponse{
		Data: pkg.Data{
			Id:        scheduleId.String(),
			UiMessage: "Post Schedule Deleted!",
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
