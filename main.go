package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
)

type RequestBase struct {
	Auth string `json:"auth"`
}

type CodeRequest struct {
	Code string `json:"code"`
	RequestBase
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func addCode(code string) {
	f, err := os.OpenFile("codes.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	check(err)

	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%s\n", code))
	check(err)
}

func deleteCode(code string) {
	fpath := "codes.txt"

	f, err := os.Open(fpath)
	check(err)
	defer f.Close()

	var bs []byte
	buf := bytes.NewBuffer(bs)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if scanner.Text() != (code) {
			_, err := buf.Write(scanner.Bytes())
			check(err)
			_, err = buf.WriteString("\n")
			check(err)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(fpath, buf.Bytes(), 0666)
	check(err)

	f, err = os.OpenFile("deleted-codes.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	check(err)

	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%s\n", code))
	check(err)
}

func getCodes() ([]string, error) {
	file, err := os.Open("codes.txt")
	check(err)
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func getPassword() (string, error) {
	file, err := os.ReadFile("password.txt")

	return string(file), err
}

func main() {
	password, err := getPassword()
	check(err)

	fmt.Println("Password: " + password)

	mux := http.NewServeMux()
	mux.HandleFunc("/addCode", func(w http.ResponseWriter, r *http.Request) {
		var addRequest CodeRequest
		err := json.NewDecoder(r.Body).Decode(&addRequest)
		check(err)

		if addRequest.Auth != password {
			return
		}

		addCode(addRequest.Code)
	})
	mux.HandleFunc("/deleteCode", func(w http.ResponseWriter, r *http.Request) {
		var addRequest CodeRequest
		err := json.NewDecoder(r.Body).Decode(&addRequest)
		check(err)

		if addRequest.Auth != password {
			return
		}

		deleteCode(addRequest.Code)
	})
	mux.HandleFunc("/getCodes", func(w http.ResponseWriter, r *http.Request) {
		var request RequestBase
		err := json.NewDecoder(r.Body).Decode(&request)
		check(err)

		if request.Auth != password {
			return
		}

		codes, err := getCodes()
		check(err)

		err = json.NewEncoder(w).Encode(codes)
		check(err)
	})

	handler := cors.Default().Handler(mux)
	err = http.ListenAndServe(":8099", handler)
	check(err)
}
