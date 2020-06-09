package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cihub/seelog"
	"rainbow/storage"
)

var s storage.Storage

type response struct {
	Data    interface{} `json:"data"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
}

func NewServer(storage storage.Storage) error {
	if storage == nil {
		return errors.New("nil storage")
	}
	s = storage

	defer func() {
		if r := recover(); r != nil {
			_ = seelog.Critical(r)
		}
	}()

	http.HandleFunc("/get", getIp)
	http.HandleFunc("/delete", deleteIp)

	err := http.ListenAndServe(":8090", nil)

	return err
}

// http: //localhost:8090/delete?=ip=0.0.0.0
func deleteIp(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Header().Add("content-type", "application/json")
		values := r.URL.Query()
		if _, ok := values["ip"]; ok {
			s.Delete(values["ip"][0])
		}
		response := &response{
			Data:    []string{},
			Code:    200,
			Message: "success",
		}
		b, _ := json.Marshal(response)
		_, _ = w.Write(b)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

//  http://localhost:8090/get
func getIp(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Header().Add("content-type", "application/json")
		if s == nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		v := s.GetRandomOne()
		response := &response{
			Data:    v,
			Code:    200,
			Message: "success",
		}
		b, _ := json.Marshal(response)
		_, _ = w.Write(b)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
