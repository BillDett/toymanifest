package model

import "log"

type ManifestLayer struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

type Manifest struct {
	SchemaVersion int               `json:"schemaVersion"`
	Config        ManifestLayer     `json:"config"`
	Layers        []ManifestLayer   `json:"layers"`
	Annotations   map[string]string `json:"annotations"`
}

func GetManifest(tag string) (*Manifest, error) {
	// TODO: Reconstruct a manifest from the database tables
	return nil, nil
}

func (m *Manifest) Save(tag string) error {
	log.Printf("Saving manifest to %s with %d layers\n", tag, len(m.Layers))
	// TODO: Save the manifest parts into the database tables
	return nil
}
