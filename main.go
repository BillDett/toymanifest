package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"toymanifest/model"
)

var port string
var storagepath string
var layerpath string

var layermediatype = "application/vnd.oci.image.layer.v1.tar+gzip"
var configmediatype = "application/vnd.oci.image.config.v1+json"

func pathsFromSum(sum string) (string, string) {
	dp := layerpath + string(os.PathSeparator) + sum[0:2]
	lp := dp + string(os.PathSeparator) + sum
	return dp, lp
}

// Manage manifests
func manifest(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	manifestId := vars["manifest_id"]
	if req.Method == http.MethodGet {
		log.Printf("GET manifest %s\n", manifestId)

	} else if req.Method == http.MethodPost {
		log.Printf("POST manifest %s\n", manifestId)
		// TODO: Should we check a Content-type here?
		manifest := model.Manifest{}
		buf, _ := ioutil.ReadAll(req.Body)
		if err := json.Unmarshal(buf, &manifest); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Unable to unmarshal manifest JSON\n%s", err)
			return
		}
		//log.Println(manifest)
		// TODO: Save the manifest structures in a database
		manifest.Save(manifestId)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed\n")
	}
}

// Deliver layer blob
func layer(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		vars := mux.Vars(req)
		log.Printf("GET layer %s\n", vars["layer_id"])
		w.Header().Set("Content-Type", layermediatype)
		_, layerfilepath := pathsFromSum(vars["layer_id"])
		f, err := os.Open(layerfilepath)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Unable to find layer %s\n%s", vars["layer_id"], err)
			return
		}
		if _, err := io.Copy(w, f); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Unable to write layer %s\n%s", vars["layer_id"], err)
			return
		}

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed\n")
	}
}

// Accept layer blobs
//
func upload(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		fmt.Fprintf(w, "POST layer\n")
		contentType := req.Header.Get("Content-type")
		if contentType != layermediatype || contentType != configmediatype {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			fmt.Fprintf(w, "Bad content type in layer\n")
			return
		}
		// TODO: Make this work for multi-part uploads (big blobs)
		//   We should have a separate /uploads directory
		buf, _ := ioutil.ReadAll(req.Body)
		sum := fmt.Sprintf("%x", sha256.Sum256(buf))

		dirpath, layerfilepath := pathsFromSum(sum)

		_, err := os.Stat(layerfilepath)
		if os.IsNotExist(err) {
			err = os.MkdirAll(dirpath, 0700)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("Unable to create directory %s\n%s", dirpath, err)
				return
			}

			fmt.Printf("Sum is %s, layerfilepath is %s\n", sum, layerfilepath)

			err = ioutil.WriteFile(layerfilepath, buf, 0700)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("Unable to write layer %s\n%s\n", sum, err)
				return
			}
		} else {
			log.Printf("Skipping existing layer %s\n", sum)
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed\n")
	}
}

func main() {

	db, err := model.StartDatabase()
	if err != nil {
		log.Printf("Unable to start database\n%s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	r := mux.NewRouter()

	r.HandleFunc("/manifest/{manifest_id}", manifest)
	r.HandleFunc("/layer/{layer_id}", layer)
	r.HandleFunc("/upload", upload)

	storagepath = "/home/bdettelb/registrydata"
	layerpath = storagepath + string(os.PathSeparator) + "sha256"
	port = "8080"
	log.Printf("Started listening on port %s\n", port)
	http.ListenAndServe(":"+port, r)

}
