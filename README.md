# Go client library for Haven OnDemand

## What is Haven OnDemand?
Haven OnDemand is a set of over 70 APIs for handling all sorts of unstructured data. Here are just some of our APIs' capabilities:
* Speech to text
* OCR
* Text extraction
* Indexing documents
* Smart search
* Language identification
* Concept extraction
* Sentiment analysis
* Web crawlers
* Machine learning

For a full list of all the APIs and to try them out, check out https://www.havenondemand.com/developer/apis

## Installation

    go get github.com/jorgemarsal/hod-go

## Usage

    import hod "github.com/jorgemarsal/hod-go"
    
    client := hod.NewHODClient(<API_KEY>, "v1", nil)

## Making GET requests

    params := &url.Values{}
    params.Add("text", "Dog")
    params.Add("database_match", "wiki_eng")
    rsp, err := client.Get("querytextindex", *params, false)

To make requests asynchronous pass true as the 3rd argument instead:

    rsp, err := client.Get("querytextindex", *params, true)

## Making POST requests

    rsp, err := client.Post("ocrdocument", &hod.PostData{File: "ocrdocument.png"}, false)

To make requests asynchronous pass true as the 3rd argument instead:
    
    rsp, err := client.Post("ocrdocument", &hod.PostData{File: "ocrdocument.png"}, true)

## License
Licensed under the MIT License.
