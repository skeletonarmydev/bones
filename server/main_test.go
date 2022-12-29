package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthCheck(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	returnHealth(w, req)

	res := w.Result()
	defer res.Body.Close()

	var responseStr string
	err := json.NewDecoder(res.Body).Decode(&responseStr)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if responseStr != "Alive" {
		t.Errorf("expected Alive got %v", responseStr)
	}
}

func TestCreateNewTypeHappy(t *testing.T) {

	var projectTypeCreateRequest ProjectTypeCreateRequest
	projectTypeCreateRequest.Name = "go-app"

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(projectTypeCreateRequest)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(buf.String()))
	w := httptest.NewRecorder()
	createNewProjectType(w, req)

	res := w.Result()
	defer res.Body.Close()

	var projectType ProjectType
	err := json.NewDecoder(res.Body).Decode(&projectType)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	if projectType.Name != "go-app" {
		t.Errorf("expected go-app got %v", projectType.Name)
	}
}

func TestCreateNewTypeEmpty(t *testing.T) {

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	createNewProjectType(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected bad status code got %v", res.StatusCode)
	}
}

func TestCreateNewTypeMalformed(t *testing.T) {

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("garbage"))
	w := httptest.NewRecorder()
	createNewProjectType(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected bad status code got %v", res.StatusCode)
	}
}

func TestCreateNewProjectEmpty(t *testing.T) {

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	createNewProject(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected bad status code got %v", res.StatusCode)
	}
}

func TestCreateNewProjectMalformed(t *testing.T) {

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("garbage"))
	w := httptest.NewRecorder()
	createNewProject(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected bad status code got %v", res.StatusCode)
	}
}
