package mediaupload

import (
	"encoding/json"
	"github.com/disintegration/imaging"
	"github.com/twinj/uuid"
	"gitlab.com/pbobby001/postit-api/pkg"
	"gitlab.com/pbobby001/postit-api/pkg/logs"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
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

	//extension := strings.Split(handler.Filename, ".")[1]

	//logs.Logger.Info("Image extension ", extension)
	//imageExt = extension

	f := make(chan multipart.File)
	go parseMultipartToFile(f, tenantNamespace, handler.Filename)
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

func parseMultipartToFile(fileChannel <-chan multipart.File, nmsp string, filename string) {
	// Listen on the file channel
	for file := range fileChannel {

		// read the file bytes
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			_ = logs.Logger.Error(err)
			return
		}

		// get the working directory for generating the path for image storage
		wd, err := os.Getwd()
		if err != nil {
			_ = logs.Logger.Error(err)
			return
		}

		// join the working directory path with the path for image storage
		join := filepath.Join(wd, "pkg/"+nmsp)

		// create a new directory for storing the image
		err = os.Mkdir(join, 0755)
		if err != nil {
			if os.IsExist(err) {
				_ = logs.Logger.Warn(err)
			} else {
				_ = logs.Logger.Error(err)
				return
			}
		}

		jsonFileToBeCreated := join + "\\f.json"

		if _, err := os.Stat(jsonFileToBeCreated); os.IsNotExist(err) {
			logs.Logger.Info("File doesnt Exist creating")
			// but create a json file for storing the file info
			jsonFile, err :=  os.Create(jsonFileToBeCreated)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}

			// create a map format for storing the file info
			fileData := make(map[string]string)
			fileData[filename] = join + "\\" + filename

			// encode the file info to json bytes
			jsonFileData, err := json.Marshal(fileData)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}

			// write the json bytes to the json file created
			_, err = jsonFile.Write(jsonFileData)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}
			_ = jsonFile.Close()
		} else {
			logs.Logger.Info("in else")
			jsFile, err := os.Open(jsonFileToBeCreated)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}

			readFile, err := ioutil.ReadAll(jsFile)
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
			logs.Logger.Info(fileData)
			// create a map format for storing the file info
			fileData[filename] = join + "\\" + filename
			logs.Logger.Info(fileData)

			// encode the file info to json bytes
			jsonFileData, err := json.Marshal(fileData)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}

			// write the json bytes to the json file created
			_, err = jsFile.Write(jsonFileData)
			if err != nil {
				_ = logs.Logger.Error(err)
				return
			}
			_ = jsFile.Close()
		}

		tempFile, err := os.Create(join + "/" + filename)
		if err != nil {
			_ = logs.Logger.Error(err)
			return
		}

		_, err = tempFile.Write(fileBytes)
		if err != nil {
			_ = logs.Logger.Error(err)
			return
		}

		img, err := imaging.Open(tempFile.Name())
		if err != nil {
			_ = logs.Logger.Error(err)
			return
		}

		imb := imaging.AdjustBrightness(img, -5)
		src := imaging.Resize(imb, 500, 0, imaging.Lanczos)
		err = imaging.Save(src, tempFile.Name())
		if err != nil {
			_ = logs.Logger.Error(err)
			return
		}
		_ = tempFile.Close()

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