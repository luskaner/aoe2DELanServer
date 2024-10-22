package models

import (
	"encoding/json"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const resourceFolder = "resources"

var configFolder = filepath.Join(resourceFolder, "config")
var responsesFolder = filepath.Join(resourceFolder, "responses")
var CloudFolder = filepath.Join(responsesFolder, "cloud")

type MainResources struct {
	keyedFilenames  mapset.Set[string]
	Login           []i.A
	ArrayFiles      map[string]i.A
	KeyedFiles      map[string][]byte
	nameToSignature map[string]string
	CloudFiles      CloudFiles
}

func (r *MainResources) Initialize(gameId string, keyedFilenames mapset.Set[string]) {
	r.ArrayFiles = make(map[string]i.A)
	r.KeyedFiles = make(map[string][]byte)
	r.nameToSignature = make(map[string]string)
	r.keyedFilenames = keyedFilenames
	r.initializeLogin(gameId)
	r.initializeResponses(gameId)
	r.initializeCloud(gameId)
}

func (r *MainResources) initializeLogin(gameId string) {
	data, err := os.ReadFile(filepath.Join(configFolder, gameId, "login.json"))
	if err != nil {
		panic(err)
	}
	var login = orderedmap.New[string, any]()
	err = json.Unmarshal(data, login)
	if err != nil {
		panic(err)
	}
	r.Login = make([]i.A, login.Len())
	j := 0
	for el := login.Oldest(); el != nil; el = el.Next() {
		r.Login[j] = i.A{el.Key, el.Value}
		j++
	}
}

func (r *MainResources) initializeResponses(gameId string) {
	dirEntries, _ := os.ReadDir(filepath.Join(responsesFolder, gameId))
	for _, entry := range dirEntries {
		data, err := os.ReadFile(filepath.Join(responsesFolder, gameId, entry.Name()))
		if err != nil {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		if r.keyedFilenames.ContainsOne(name) {
			var result = orderedmap.New[string, any]()
			err = json.Unmarshal(data, result)
			if err == nil {
				rawSignature, _ := result.Get("dataSignature")
				serverSignature := rawSignature.(string)
				r.KeyedFiles[name] = data
				r.nameToSignature[name] = serverSignature
			}
		} else {
			var result i.A
			err = json.Unmarshal(data, &result)
			if err == nil {
				r.ArrayFiles[name] = result
			}
		}
	}
}

func (r *MainResources) initializeCloud(gameId string) {
	r.CloudFiles = *BuildCloudfilesIndex(filepath.Join(configFolder, gameId), filepath.Join(CloudFolder, gameId))
}

func (r *MainResources) ReturnSignedAsset(name string, w *http.ResponseWriter, req *http.Request, keyedResponse bool) {
	var serverSignature string
	var response any
	if keyedResponse {
		response = r.KeyedFiles[name]
		serverSignature = r.nameToSignature[name]
	} else {
		response = r.ArrayFiles[name]
		arrayResponse := response.(i.A)
		serverSignature = arrayResponse[len(arrayResponse)-1].(string)
	}
	if req.URL.Query().Get("signature") != serverSignature {
		if keyedResponse {
			i.RawJSON(w, response.([]byte))
		} else {
			i.JSON(w, response)
		}
		return
	}
	if keyedResponse {
		i.RawJSON(w, []byte(fmt.Sprintf(`{"result":0,"dataSignature":"%s"}`, serverSignature)))
	} else {
		emptyArrays := make(i.A, len(response.(i.A))-2)
		ret := i.A{0}
		ret = append(ret, emptyArrays...)
		ret = append(ret, serverSignature)
		i.JSON(w, ret)
	}
}
