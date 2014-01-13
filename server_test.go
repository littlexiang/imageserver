package imageserver

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
)

var UploadUrl = "http://localhost:9998/test/test.jpg"
var GetUrl = "http://localhost:9999/test/test.jpg"

func Test_Get(t *testing.T) {

}

func Test_Post(t *testing.T) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	filename := "test.jpg"

	fileWriter, err := bodyWriter.CreateFormFile("photo", filename)
	if err != nil {
		t.Error("error writing to buffer")
	}

	fh, err := os.Open(filename)
	if err != nil {
		t.Error("error opening file")
	}

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		t.Error("io copy error")
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(UploadUrl, contentType, bodyBuf)

	t.Logf("%#v", resp)

	if err != nil || resp.StatusCode != 200 {
		t.Error("resp error")
	}
	defer resp.Body.Close()
}

func Test_Delete(t *testing.T) {

	req, err := http.NewRequest("DELETE", UploadUrl, nil)

	resp, err := http.DefaultClient.Do(req)

	t.Logf("%#v", resp)

	if err != nil || resp.StatusCode != 200 {

		t.Error()
	}
	// do something with resp
}
