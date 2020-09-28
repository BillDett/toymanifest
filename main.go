package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type ManifestLayer struct {
	mediaType string
	size      int
	digest    string
}

type Manifest struct {
	schemaVersion int
	config        ManifestLayer
	layers        []ManifestLayer
	annotations   map[string]string
}

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
	if req.Method == http.MethodGet {
		fmt.Fprintf(w, "GET manifest\n")

	} else if req.Method == http.MethodPost {
		fmt.Fprintf(w, "POST manifest\n")

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
		if contentType != layermediatype {
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

	r := mux.NewRouter()

	r.HandleFunc("/manifest", manifest)
	r.HandleFunc("/layer/{layer_id}", layer)
	r.HandleFunc("/upload", upload)

	storagepath = "/home/bdettelb/registrydata"
	layerpath = storagepath + string(os.PathSeparator) + "sha256"
	port = "8080"
	log.Printf("Started listening on port %s\n", port)
	http.ListenAndServe(":"+port, r)

}
