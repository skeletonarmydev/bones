package main

import (
	github "bones/server/handlers/github"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var banner = `
 ____                        
|  _ \                       
| |_) | ___  _ __   ___  ___ 
|  _ < / _ \| '_ \ / _ \/ __|
| |_) | (_) | | | |  __/\__ \
|____/ \___/|_| |_|\___||___/
                             
Scaffolding service
`

type Article struct {
	Id      string `json:"Id"`
	Title   string `json:"Title"`
	Desc    string `json:"desc"`
	Content string `json:"content"`
}

var Articles []Article

func returnAllArticles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllArticles")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Articles)
}

func createNewArticle(w http.ResponseWriter, r *http.Request) {
	// get the body of our POST request
	// unmarshal this into a new Article struct
	// append this to our Articles array.
	reqBody, _ := ioutil.ReadAll(r.Body)
	var article Article
	json.Unmarshal(reqBody, &article)
	// update our global Articles array to include
	// our new Article
	Articles = append(Articles, article)

	json.NewEncoder(w).Encode(article)
}

func handleRequests() {
	fmt.Printf(banner)

	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/all", returnAllArticles)
	myRouter.HandleFunc("/article", createNewArticle).Methods("POST")
	// finally, instead of passing in nil, we want
	// to pass in our newly created router as the second
	// argument
	fmt.Println("Now online and ready")
	log.Fatal(http.ListenAndServe(":8080", myRouter))

}

func main() {
	Articles = []Article{
		Article{Id: "1", Title: "Hello", Desc: "Article Description", Content: "Article Content"},
		Article{Id: "2", Title: "Hello 2", Desc: "Article Description", Content: "Article Content"},
	}

	var help = flag.Bool("help", false, "Show help")
	var createAction = flag.Bool("create", false, "create")
	var destroyAction = flag.Bool("destroy", false, "destroy")
	var repoFlag = ""
	var appName = ""

	flag.StringVar(&repoFlag, "repo", "", "The repo of the scaffold template")
	flag.StringVar(&appName, "name", "", "The name of your new app")

	// Parse the flag
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *createAction {
		github.CreateRepo(appName, repoFlag)
	} else if *destroyAction {
		github.DestroyRepo(appName)
	}

	//handleRequests()
}
