package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/inquizarus/golbag/models"
	"github.com/inquizarus/golbag/pkg/logging"
	"github.com/inquizarus/golbag/pkg/storages"
	"github.com/inquizarus/gorest"
)

// MakeItemHandler creates handler for item interactions
func MakeItemHandler(s storages.Storage, log logging.Logger) gorest.Handler {
	return &gorest.BaseHandler{
		Name: "bucket",
		Path: "/buckets/{bucket_name}/items/{item_name}",
		Get: func(w http.ResponseWriter, r *http.Request, p map[string]string) {
			var response Response
			var item models.Item
			defer r.Body.Close()
			bucketName := p["bucket_name"]
			itemName := p["item_name"]
			log.Debug(fmt.Sprintf(fmt.Sprintf("trying to retrieve %s from bucket %s", itemName, bucketName)))
			itemBytes, err := s.GetItemFromBucket([]byte(itemName), []byte(bucketName))
			if nil != err {
				log.Error(fmt.Errorf("could not retrieve %s from bucket %s: %v", itemName, bucketName, err))
				response.AddError(err)
				w.WriteHeader(http.StatusConflict)
				writeResponse(w, response)
				return
			}
			json.Unmarshal(itemBytes, &item)
			log.Debug(fmt.Sprintf("successfully retrieved %s from bucket %s", itemName, bucketName))
			w.Write(item.Content)
		},
		Post: func(w http.ResponseWriter, r *http.Request, p map[string]string) {
			var response Response
			var item models.Item
			defer r.Body.Close()
			bucketName := p["bucket_name"]
			itemName := p["item_name"]
			body, err := ioutil.ReadAll(r.Body)
			if nil == err {
				log.Debug(fmt.Sprintf("trying to create item %s in bucket %s", itemName, bucketName))
				item.Content = body
				item.Meta.CreatedAt = time.Now().Unix()
				if itemBytes, err := json.Marshal(item); nil == err {
					err = s.AddItemToBucket([]byte(itemName), []byte(bucketName), itemBytes)
				}

			}
			if nil != err {
				log.Error(fmt.Errorf("could not create item %s in bucket %s: %v", itemName, bucketName, err))
				response.AddError(err)
				w.WriteHeader(http.StatusConflict)
				writeResponse(w, response)
				return
			}
			log.Debug(fmt.Sprintf("successfully created item %s in bucket %s", itemName, bucketName))
			response.Meta.Success = true
			writeResponse(w, response)
		},
		Delete: func(w http.ResponseWriter, r *http.Request, p map[string]string) {
			var response Response
			defer r.Body.Close()
			bucketName := p["bucket_name"]
			itemName := p["item_name"]
			log.Debug(fmt.Sprintf("trying to delete %s from bucket %s", itemName, bucketName))
			err := s.DeleteItemFromBucket([]byte(itemName), []byte(bucketName))
			if nil != err {
				log.Error(fmt.Errorf("could not delete %s from bucket %s: %v", itemName, bucketName, err))
				response.AddError(err)
				w.WriteHeader(http.StatusInternalServerError)
				writeResponse(w, response)
				return
			}
			log.Debug(fmt.Sprintf("successfully delete %s from bucket %s", itemName, bucketName))
			response.Meta.Success = true
			writeResponse(w, response)
		},
	}
}
