package main

import "encoding/xml"

type (
	LifecycleConfiguration struct {
		XMLName xml.Name `xml:"LifecycleConfiguration"`
		Xmlns   string   `xml:"xmlns,attr"`
		Id      string   `xml:"Rule>ID"`
		Prefix  string   `xml:"Rule>Prefix"`
		Status  string   `xml:"Rule>Status"`
		Days    int      `xml:"Rule>Expiration>Days"`
	}
)

func NewLifecycleConfiguration(id, prefix string, status bool, days int) LifecycleConfiguration {
	lc := LifecycleConfiguration{
		Xmlns:  "http://s3.amazonaws.com/doc/2006-03-01/",
		Id:     id,
		Prefix: prefix,
		Status: "Disabled",
		Days:   days,
	}

	if status {
		lc.Status = "Enabled"
	}

	return lc
}
