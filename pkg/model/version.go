package model

import "encoding/json"

type Version int

const (
	V1 Version = iota + 1
	V2
)

func (v Version) Int() int {
	return int(v)
}

func GetVersion(payload []byte) Version {
	type versionStruct struct {
		Version Version `json:"version"`
	}

	var decoded versionStruct
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return V1
	}

	if decoded.Version == V1 || decoded.Version.Int() == 0 {
		return V1
	}

	return decoded.Version
}