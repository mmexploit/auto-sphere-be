package server

import (
	"errors"
	"net/http"

	"github.com/Mahider-T/autoSphere/internal/database"
	"github.com/Mahider-T/autoSphere/internal/pkg"
)

func (ser Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	file, err := pkg.UploadFile()
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	err = ser.writeJSON(w, http.StatusOK, envelope{"token": file.Token, "fid": file.Fid, "url": file.URL}, nil)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}
}

func (ser Server) fetchFile(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Fid string `json:"fid"`
	}

	fileURL, err := pkg.FetchFile(input.Fid)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.notFoundResponse(w, r)
			return
		default:
			ser.serverErrorResponse(w, r, err)
		}
		return
	}
	err = ser.writeJSON(w, http.StatusOK, envelope{"file_url": fileURL}, nil)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}

}

func (ser Server) deleteFile(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Fid string `json:"fid"`
	}
	err := pkg.DeleteFile(input.Fid)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	err = ser.writeJSON(w, http.StatusOK, envelope{}, nil)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}

}
