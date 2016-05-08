package hod

import (
	"net/url"
	"testing"
)


var hodGetTests = []struct {
        op        string
	params    url.Values
	async     bool
}{
        {"querytextindex", url.Values{"text": []string{"Dog"}, "database_match": []string{"wiki_eng"}}, false},
	{"querytextindex", url.Values{"text": []string{"Dog"}, "database_match": []string{"wiki_eng"}}, true},
}

var hodPostTests = []struct {
        op        string
	postData  *PostData
	async     bool
}{
        {"querytextindex", &PostData{File: "ocrdocument.png"}, false},
	{"querytextindex", &PostData{File: "ocrdocument.png"}, true},
}

var client *HODClient = NewHODClient("d3962fe9-b18b-48b7-8c53-3364afa0cc86", "v1", nil)

func TestGet(t *testing.T) {
        for _, tt := range hodGetTests {
                client.Get(tt.op, tt.params, tt.async)
        }
}

func TestPost(t *testing.T) {
        for _, tt := range hodPostTests {
                client.Post(tt.op, tt.postData, tt.async)
        }
}



