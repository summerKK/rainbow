package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cihub/seelog"
	"rainbow/result"
	"rainbow/storage"
)

var storageDb storage.Storage

type response struct {
	Data    result.Result `json:"data"`
	Code    int           `json:"code"`
	Message string        `json:"message"`
}

func NewServer(storage storage.Storage) error {
	if storage == nil {
		return errors.New("nil storage")
	}
	storageDb = storage

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
			storageDb.Delete(values["ip"][0])
		}
		response := &response{
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
		if storageDb == nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		response := &response{
			Code:    200,
			Message: "success",
		}

		v := storageDb.GetRandomOne()
		var res result.Result
		err := json.Unmarshal([]byte(v), &res)
		if err != nil {
			_ = seelog.Errorf("json unmarshal error:%v,json:%s", err, v)
			response.Code = 400
			response.Message = "获取失败请重试!"
		} else {
			response.Data = res
		}

		b, _ := json.Marshal(response)
		_, _ = w.Write(b)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
