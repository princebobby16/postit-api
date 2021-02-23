package mediaupload

import (
	"encoding/json"
	"github.com/disintegration/imaging"
	"github.com/twinj/uuid"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"postit-backend-api/pkg"
	"postit-backend-api/pkg/logs"
	"strings"
	"time"
)

const (
	_ int = iota
	_ = 1 << (10 * iota)
	MB
)

var imageExt string

func HandleMediaUpload(w http.ResponseWriter, r *http.Request) {

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
	logs.Logger.Info("Headers => TraceId: " + traceId + ", TenantNamespace: "+tenantNamespace)

	imageId := r.URL.Query().Get("image_id")

	err = r.ParseMultipartForm(10 * MB)
	if err != nil {
		logs.Logger.Error(err)
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	file, handler, err := r.FormFile("media_file")
	if err != nil {
		logs.Logger.Info("Error Retrieving the File")
		logs.Logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	logs.Logger.Info("Uploaded File: ", handler.Filename)
	logs.Logger.Info("File Size: ", handler.Size)
	logs.Logger.Info("MIME Header: ", handler.Header)

	extension := strings.Split(handler.Filename, ".")[1]

	logs.Logger.Info("Image extension ", extension)
	imageExt = extension

	f := make(chan multipart.File)
	go parseMultipartToFile(f, tenantNamespace, imageId, extension)
	f <- file
	close(f)

	_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
		Data: pkg.Data{
			Id:        "",
			UiMessage: "FILE UPLOADING",
		},
		Meta: pkg.Meta{
			Timestamp:     time.Now(),
			TransactionId: transactionId.String(),
			TraceId:       traceId,
			Status:        "SUCCESS",
		},
	})
}

func parseMultipartToFile(fileChannel <-chan multipart.File, nmsp string, id string, extension string) {
	// create a temp file
	for file := range fileChannel {

		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			logs.Logger.Error(err)
			return
		}

		wd, err := os.Getwd()
		if err != nil {
			logs.Logger.Error(err)
			return
		}

		join := filepath.Join(wd, "pkg/"+nmsp)

		err = os.Mkdir(join, 0755)
		if err != nil {
			if os.IsExist(err) {
				logs.Logger.Warn(err)
			} else {
				logs.Logger.Error(err)
				return
			}
		}

		tempFile, err := ioutil.TempFile(join, "upload_*."+extension)
		if err != nil {
			logs.Logger.Error(err)
			return
		}

		_, err = tempFile.Write(fileBytes)
		if err != nil {
			logs.Logger.Error(err)
			return
		}

		img, err := imaging.Open(tempFile.Name())
		if err != nil {
			logs.Logger.Error(err)
			return
		}

		imb := imaging.AdjustBrightness(img, -5)
		src := imaging.Resize(imb, 500, 0, imaging.Lanczos)
		err = imaging.Save(src, tempFile.Name())
		if err != nil {
			logs.Logger.Error(err)
			return
		}
		tempFile.Close()

		err = os.Rename(tempFile.Name(), path.Join(join, id+"."+extension))
		if err != nil {
			logs.Logger.Error(err)
			return
		}

		logs.Logger.Info("Successfully resized image...")
	}
}

func HandleCancelMediaUpload(w http.ResponseWriter, r *http.Request) {
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

	imageId := r.URL.Query().Get("image_id")
	logs.Logger.Info(imageId)

	workingDir, err := os.Getwd()
	if err != nil {
		logs.Logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logs.Logger.Info(workingDir)

	imageStoragePath := path.Join(workingDir, "/pkg/" + tenantNamespace)
	logs.Logger.Info(imageStoragePath)

	if imageStoragePath == "" {
		logs.Logger.Warn("No image has been uploaded to server")
		w.WriteHeader(http.StatusGone)
		return
	}

	err = os.Remove(imageStoragePath + "/" + imageId + "." + imageExt)
	if err != nil {
		logs.Logger.Error(err)
		return
	}

	_ = json.NewEncoder(w).Encode(pkg.StandardResponse{
		Data: pkg.Data{
			Id:        "",
			UiMessage: "FILE DELETED",
		},
		Meta: pkg.Meta{
			Timestamp:     time.Now(),
			TransactionId: transactionId.String(),
			TraceId:       traceId,
			Status:        "SUCCESS",
		},
	})

}