package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
)

const token string = "the_secret_token" // read from env variable
const BASE_FILES_PATH string = "files/"
const PRE_COMMIT_CONFIG string = "pre-commit.config"

type ci_struct struct {
	Repo    string `json:"repo"`
	Python  string `json:"python"`
	Golang  string `json:"golang"`
	Node    string `json:"node"`
	Ts      string `json:"ts"`
	Flutter string `json:"flutter"`
	Dart    string `json:"dart"`
	Docker  string `json:"docker"`
	Shell   string `json:"shell"`
}

func main() {
	fmt.Printf("Starting server at port 8080\n")
	http.HandleFunc("/github/ci", CIHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}

func CIHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q\n", html.EscapeString(r.URL.Path))
	header_value := r.Header.Get("Authorization")
	if "Bearer "+token != header_value {
		fmt.Fprint(w, "Unauthorized request\n")
		return
	}
	fmt.Fprintf(w, "Authorization SUCCESS\n")

	if r.Method != "POST" {
		fmt.Fprint(w, "Invalid request method\n")
		return
	}
	r_body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(w, "Error reading body\n")
		return
	}
	fmt.Fprintf(w, "Body: %s\n", string(r_body))
	var ci ci_struct
	err = json.Unmarshal(r_body, &ci)
	if err != nil {
		fmt.Fprint(w, "Error unmarshalling body\n")
		return
	}

	var ci_list []string

	v := reflect.ValueOf(ci)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).String() == "true" {
			ci_list = append(ci_list, v.Type().Field(i).Name)
		}
	}
	fmt.Fprintf(w, "CI list: %s\n", ci_list)

	add_checks(ci_list)
	fmt.Fprintf(w, "Added checks to pre-commit.config file\n")

	// display the contents of the pre-commit.config file
	f, err := os.Open(PRE_COMMIT_CONFIG)
	if err != nil {
		fmt.Printf("Error opening file:%s\n", PRE_COMMIT_CONFIG)
		return
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Printf("Error reading file\n")
		return
	}
	fmt.Fprintf(w, "File contents of %s: \n%s\n", PRE_COMMIT_CONFIG, string(content))
}

func add_checks(ci_list []string) {
	// combine the data of all files in ci_list to a file named pre-commit.config
	// create the file even if it exists
	os.Create(PRE_COMMIT_CONFIG)
	f, err := os.OpenFile(PRE_COMMIT_CONFIG, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("Error opening file\n")
		return
	}

	defer f.Close()
	for i := 0; i < len(ci_list); i++ {

		var file_name string = ci_list[i] + ".ci"
		var file_path string = BASE_FILES_PATH + file_name

		content, err := os.ReadFile(file_path)
		if err != nil {
			fmt.Printf("Error reading file:%s\n", file_name)
			return
		}

		// append the data to the pre-commit.config file
		_, err = fmt.Fprintf(f, string(content)+"\n")
		if err != nil {
			fmt.Printf("Error writing to file:%s\n", PRE_COMMIT_CONFIG)
			return
		}
	}
}
