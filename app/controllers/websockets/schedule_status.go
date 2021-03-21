package websockets

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/lib/pq"
	"gitlab.com/pbobby001/postit-api/app/middlewares"
	"gitlab.com/pbobby001/postit-api/db"
	"gitlab.com/pbobby001/postit-api/pkg"
	"gitlab.com/pbobby001/postit-api/pkg/logs"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func Writer(conn *websocket.Conn, tenantNamespace string, connection *sql.DB) {
	for {
		ticker := time.NewTicker(5 * time.Second)

		for t := range ticker.C {
			logs.Logger.Info("Updating status: ", t)
			statuses, err := FetchStatuses(connection, tenantNamespace)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}

			jsonBytes, err := json.Marshal(statuses)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}

			err = conn.WriteMessage(websocket.TextMessage, jsonBytes)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}

		}
	}
}

func FetchStatuses(connection *sql.DB, tenantNamespace string) ([]pkg.ScheduleStatus, error) {

	//  Prepare the query
	query := fmt.Sprintf("SELECT schedule_id, schedule_title, schedule_from, schedule_to, post_ids FROM %s.schedule WHERE is_due = $1", tenantNamespace)

	// run the query
	rows, err := connection.Query(query, true)
	if err != nil {
		return nil, err
	}

	var schedules []pkg.PostSchedule
	for rows.Next() {
		var scheduleData pkg.PostSchedule
		err = rows.Scan(
			&scheduleData.ScheduleId,
			&scheduleData.ScheduleTitle,
			&scheduleData.From,
			&scheduleData.To,
			pq.Array(&scheduleData.PostIds),
		)
		if err != nil {
			return nil, err
		}

		schedules = append(schedules, scheduleData)
	}

	var scheduleStatuses []pkg.ScheduleStatus
	var posts []pkg.ScheduledPost
	if schedules != nil {

		query = fmt.Sprintf("SELECT post_id, post_message, hash_tags, post_image, image_paths, post_status FROM %s.scheduled_post WHERE scheduled_post_id = $1 AND post_status = $2", tenantNamespace)

		for _, i := range schedules {
			rows, err = connection.Query(query, i.ScheduleId, true)
			if err != nil {
				return nil, err
			}

			for rows.Next() {
				var post pkg.ScheduledPost
				err = rows.Scan(
					&post.PostId,
					&post.PostMessage,
					pq.Array(&post.HashTags),
					pq.Array(&post.PostImages),
					pq.Array(&post.ImagePaths),
					&post.PostStatus,
				)
				if err != nil {
					return nil, err
				}

				posts = append(posts, post)
			}
			scheduleStatus := pkg.ScheduleStatus{
				ScheduleId:    i.ScheduleId,
				ScheduleTitle: i.ScheduleTitle,
				From:  		   i.From,
				To:   		   i.To,
				TotalPost: 	   len(i.PostIds),
				Posts:         posts,
				PostCount:     len(posts),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			scheduleStatuses = append(scheduleStatuses, scheduleStatus)
		}
	}

	return scheduleStatuses, nil
}

func ScheduleStatus(w http.ResponseWriter, r *http.Request) {

	logs.Logger.Info("connecting to websocket")

	ws, err := upgrade(w, r)
	if err != nil {
		_ = logs.Logger.Error(err)
		return
	}

	var webSocketHandshake pkg.WebSocketHandShakeData
	err = ws.ReadJSON(&webSocketHandshake)
	if err != nil {
		_ = logs.Logger.Error(err)
		return
	}
	logs.Logger.Info(webSocketHandshake)

	// validate token
	err = pkg.WebSocketTokenValidateToken(webSocketHandshake.AuthToken, middlewares.PrivateKey, webSocketHandshake.TenantNamespace)
	if err != nil {
		logs.Logger.Error(err)
		ws.Close()
		return
	}
	//q := `UPDATE postit.scheduled_post SET post_status = true WHERE post_id=$1`
	//_, err = db.Connection.Exec(q, "298bccf8-c103-474b-b708-a8797860feb0")
	//if err != nil {
	//	logs.Logger.Error(err)
	//	return
	//}

	logs.Logger.Info("connection upgraded")
	go Writer(ws, webSocketHandshake.TenantNamespace, db.Connection)
}
