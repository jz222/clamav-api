package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

var PORT = "8080"
var SCAN_DIR = "root/uploads"

type response struct {
	IsMalicious   bool   `json:"isMalicious"`
	DetectedVirus string `json:"detectedVirus,omitempty"`
	Result        string `json:"result"`
}

func init() {
	port := os.Getenv("PORT")
	if port != "" {
		PORT = port
	}
}

func respondWithSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	res := map[string]interface{}{
		"success": true,
		"data":    data,
	}

	json.NewEncoder(w).Encode(res)
}

func respondWithError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	res := map[string]interface{}{
		"success": false,
		"message": message,
	}

	json.NewEncoder(w).Encode(res)
}

func scanFile(fileName string) (bool, string, string, error) {
	cmd := exec.Command("clamscan", fileName)

	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut

	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	cmd.Start()

	done := make(chan error)

	go func() {
		done <- cmd.Wait()
	}()

	timeout := time.After(60 * time.Second)

	select {
	case <-timeout:
		cmd.Process.Kill()
		return false, "", "", errors.New("scan exceeced the timeout")

	case err := <-done:
		if err != nil && err.Error() == "exit status 1" {
			exp := regexp.MustCompile(`^.*?: (.*?) FOUND`)
			match := exp.FindStringSubmatch(stdOut.String())

			detectedVirus := ""

			if len(match) > 1 {
				detectedVirus = match[1]
			}

			return true, stdOut.String(), detectedVirus, nil
		}

		if err != nil {
			err = errors.New("failed to scan file with error: " + err.Error() + " " + stdErr.String())
			log.Println(err.Error())
			return false, err.Error(), "", err
		}

		return false, stdOut.String(), "", nil
	}
}

func scanController(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(40 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, "failed to parse file from form with error: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	localFile, err := ioutil.TempFile(SCAN_DIR, "*-"+handler.Filename)
	if err != nil {
		respondWithError(w, "failed to create temporary local file with error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer localFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		respondWithError(w, "failed to read file with error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	localFile.Write(fileBytes)

	isMalicious, result, detectedVirus, err := scanFile(localFile.Name())
	go os.Remove(localFile.Name())
	if err != nil {
		respondWithError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := response{
		IsMalicious:   isMalicious,
		DetectedVirus: detectedVirus,
		Result:        result,
	}

	respondWithSuccess(w, res)
}

func main() {
	h := http.NewServeMux()

	h.HandleFunc("/scan", scanController)

	s := &http.Server{
		Addr:           ":" + PORT,
		Handler:        h,
		ReadTimeout:    5 * time.Minute,
		WriteTimeout:   5 * time.Minute,
		MaxHeaderBytes: 0,
	}

	go func() {
		fmt.Println("Server listening on port " + PORT)
		err := s.ListenAndServe()
		if err != nil {
			log.Println(err.Error())
		}
	}()

	quit := make(chan os.Signal)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.Shutdown(ctx)
	if err != nil {
		log.Fatal("failed to stop server with error: ", err.Error())
	}

	log.Println("server shut down gracefully")
}
