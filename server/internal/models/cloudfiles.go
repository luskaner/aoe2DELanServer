package models

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
)

type CloudfilesIndex struct {
	Id       int    `json:"id"`
	Key      string `json:"key"`
	Type     string `json:"type"`
	Version  string `json:"version"`
	ETag     string `json:"etag"`
	Created  string `json:"created"`
	Length   int
	Checksum string
}

type CloudFiles struct {
	baseFolder  string
	Value       map[string]CloudfilesIndex
	Credentials *Credentials
}

func (m *CloudFiles) GetByKey(key string) (string, *CloudfilesIndex, bool) {
	for filename, file := range (*m).Value {
		if file.Key == key {
			return filename, &file, true
		}
	}
	return "", nil, false
}

func (m *CloudFiles) ReadFile(name string) ([]byte, error) {
	_, ok := (*m).Value[name]
	if !ok {
		return nil, os.ErrInvalid
	}
	return os.ReadFile(filepath.Join(m.baseFolder, name))
}

func BuildCloudfilesIndex(configFolder string, baseFolder string) *CloudFiles {
	data, err := os.ReadFile(filepath.Join(configFolder, "cloudfilesIndex.json"))
	if err != nil {
		panic(err)
	}
	index := CloudFiles{baseFolder: baseFolder, Credentials: &Credentials{}}
	err = json.Unmarshal(data, &index.Value)
	if err != nil {
		panic(err)
	}
	index.Credentials.Initialize()
	for i, fileInfo := range index.Value {
		data, err = index.ReadFile(i)
		if err != nil {
			panic(err)
		}
		fileInfo.Length = len(data)
		hash := md5.Sum(data)
		hashSlice := hash[:]
		fileInfo.Checksum = base64.StdEncoding.EncodeToString(hashSlice)
		index.Value[i] = fileInfo
	}
	return &index
}
