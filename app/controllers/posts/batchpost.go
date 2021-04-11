package posts

import (
	"encoding/json"
	"github.com/twinj/uuid"
	"gitlab.com/pbobby001/postit-api/pkg"
	"gitlab.com/pbobby001/postit-api/pkg/logs"
	"io/ioutil"
	"net/http"
	"time"
)

//const (
//	_        = iota
//	KB int64 = 1 << (10 * iota)
//	MB
//	GB
//)

func HandleBatchPost(w http.ResponseWriter, r *http.Request) {
	// Generate transaction id for this transaction
	transactionId := uuid.NewV4()

	// validate the headers
	var headers, err = pkg.ValidateHeaders(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
		_ = logs.Logger.Error(err)
		return
	}

	traceId := headers["trace-id"]
	tenantNamespace := headers["tenant-namespace"]

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
		_ = logs.Logger.Error(err)
		return
	}

	var posts []pkg.Post

	err = json.Unmarshal(body, &posts)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
			Data: pkg.Data{
				Id:        "",
				UiMessage: err.Error(),
			},
			Meta: pkg.Meta{
				Timestamp:     time.Now(),
				TransactionId: transactionId.String(),
				TraceId:       traceId,
				Status:        "FAILED",
			},
		})
		_ = logs.Logger.Error(err)
		return
	}
	// Generate an id for the post
	postId := uuid.NewV4()

	// If the request is received; Iterate over it and store it in the db
	for _, post := range posts {
		logs.Logger.Info("Post Image", post.PostImages)
		hashTagList := pkg.GenerateHashTags(post.HashTags)
		logs.Logger.Info("Hash Tag List: ", hashTagList)

		/* Replace post arrays with the new array list
		Totally unnecessary but I did it anyway */
		post.HashTags = hashTagList
		err = pkg.CreatePost(post, tenantNamespace, postId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
				Data: pkg.Data{
					Id:        "",
					UiMessage: err.Error(),
				},
				Meta: pkg.Meta{
					Timestamp:     time.Now(),
					TransactionId: transactionId.String(),
					TraceId:       traceId,
					Status:        "FAILED",
				},
			})
			_ = logs.Logger.Error(err)
			return
		}
	}

	resp := &pkg.StandardResponse{
		Data: pkg.Data{
			Id:        transactionId.String(),
			UiMessage: "Messages successfully uploaded",
		},
		Meta: pkg.Meta{
			Timestamp:     time.Now(),
			TransactionId: transactionId.String(),
			TraceId:       traceId,
			Status:        "SUCCESS",
		},
	}

	_ = json.NewEncoder(w).Encode(&resp)

}

//func InitPublisher(posts []pkg.Post) error {
//	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
//	if err != nil {
//		return err
//	}
//	defer conn.Close()
//
//	amqpChannel, err := conn.Channel()
//	if err != nil {
//		return err
//	}
//	defer amqpChannel.Close()
//
//	err = amqpChannel.ExchangeDeclare(
//		"msg-distributor", // name
//		"direct",          // type
//		true,              // durable
//		false,             // auto-deleted
//		false,             // internal
//		false,             // no-wait
//		nil,               // arguments
//	)
//	if err != nil {
//		return err
//	}
//
//	body, err := json.Marshal(posts)
//	if err != nil {
//		return err
//	}
//
//	logs.Logger.Error(string(body))
//
//	err = amqpChannel.Publish(
//		"msg-distributor",
//		"",
//		false,
//		false,
//		amqp.Publishing{
//			ContentType:     "application/json",
//			ContentEncoding: "application/json",
//			DeliveryMode:    amqp.Persistent,
//			Priority:        9,
//			Timestamp:       time.Now(),
//			AppId:           "PostIt",
//			Body:            body,
//		})
//
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
