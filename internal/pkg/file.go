package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Mahider-T/autoSphere/internal/database"
)

var (
	masterURL       = os.Getenv("SFS_MASTER_SERVER")
	masterLookupURL = os.Getenv("SFS_MASTER_LOOKUP")
	volumeURL       = os.Getenv("SFS_VOLUME_SERVER")
)

type FileProps struct {
	Fid   string `json:"fid"`
	URL   string `json:"url"`
	Token string `json:"token"`
}

func UploadFile() (FileProps, error) {

	var fileProps FileProps

	res, err := http.Post(masterURL, "application/json", nil)
	if err != nil {
		return FileProps{}, err
	}

	var authHeader = res.Header.Get("Authorization")
	if authHeader == "" {
		return FileProps{}, err
	}
	bearerToken := strings.Split(authHeader, " ")[1]

	bodyBytes, err := io.ReadAll(res.Body)

	if err != nil {
		return FileProps{}, err
	}

	var jsData map[string]interface{}

	if err = json.Unmarshal(bodyBytes, &jsData); err != nil {

		return FileProps{}, err
	}

	fidInterface, ok := jsData["fid"]
	if !ok {
		return FileProps{}, err
	}
	fid, ok := fidInterface.(string) // Type assertion
	if !ok {
		return FileProps{}, err
	}

	urlInterface, ok := jsData["url"]
	if !ok {
		return FileProps{}, err
	}
	url, ok := urlInterface.(string) // Type assertion
	if !ok {
		return FileProps{}, err
	}

	fileProps.Fid = fid
	fileProps.URL = url
	fileProps.Token = bearerToken
	fmt.Print(fid)

	return fileProps, nil
}

func FetchFile(fid string) (string, error) {
	// Extract volume ID from fid
	volumeID := strings.Split(fid, ",")[0]
	lookupURL := fmt.Sprintf("%s/dir/lookup?volumeId=%s", masterURL, volumeID)

	// Request to get the file location
	res, err := http.Get(lookupURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch file location, status: %d", res.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	// Parse JSON response
	var data struct {
		Locations []struct {
			PublicUrl string `json:"publicUrl"`
		} `json:"locations"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	if len(data.Locations) == 0 {
		return "", database.ErrRecordNotFound
	}

	// Construct the file URL
	fileURL := fmt.Sprintf("http://%s/%s", data.Locations[0].PublicUrl, fid)
	return fileURL, nil
}

func DeleteFile(fid string) error {
	// Lookup file location
	lookupURL := fmt.Sprintf("%s?fileId=%s", masterLookupURL, fid)
	res, err := http.Get(lookupURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to lookup file, status: %d", res.StatusCode)
	}

	authHeader := res.Header.Get("Authorization")
	var token string
	if strings.HasPrefix(authHeader, "BEARER ") {
		token = strings.TrimPrefix(authHeader, "BEARER ")
	}

	// Delete request to volume server
	deleteURL := fmt.Sprintf("%s/%s", volumeURL, fid)
	req, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	deleteRes, err := client.Do(req)
	if err != nil {
		return err
	}
	defer deleteRes.Body.Close()

	if deleteRes.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete file, status: %d", deleteRes.StatusCode)
	}

	fmt.Println("Delete successful.")
	return nil
}
