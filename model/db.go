package model

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB

const tableDDL = `
CREATE TABLE IF NOT EXISTS manifest (
	id			NUMERIC PRIMARY KEY,
	config_id 	NUMERIC,
	tag			TEXT NOT NULL,
	version		NUMERIC
);

CREATE TABLE IF NOT EXISTS layer (
	id			NUMERIC PRIMARY KEY,
	digest		TEXT NOT NULL,
	mediaType	TEXT,
	size		TEXT
);

CREATE TABLE IF NOT EXISTS manifestlayer (
	manifest_id	NUMERIC,
	layer_id 	NUMERIC,
	PRIMARY KEY (manifest_id, layer_id),
	FOREIGN KEY (manifest_id)
		REFERENCES manifest (id),
	FOREIGN KEY (layer_id)
		REFERENCES layer (id)
);

CREATE TABLE IF NOT EXISTS annotation (
	manifest_id	NUMERIC,
	key			TEXT,
	value		TEXT,
	PRIMARY KEY (manifest_id, key)
);
`

const dropDDL = `
DROP TABLE manifest;
DROP TABLE layer;
DROP TABLE manifestlayer;
DROP TABLE annotation;
`

func StartDatabase() (*sql.DB, error) {

	database, err := sql.Open("sqlite3", "./toymanifest.db")
	if err != nil {
		return nil, err
	}

	_, err = database.Exec(tableDDL)
	if err != nil {
		return nil, err
	}

	return database, nil
}
