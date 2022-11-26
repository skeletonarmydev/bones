package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hashicorp/terraform-exec/tfexec"
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

	workingDir := os.Getenv("TERRAFORM_FILES_DIR")
	execPath := os.Getenv("TERRAFORM_EXEC_DIR")

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background())
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	pass, err := tf.Plan(context.Background(), tfexec.Out(workingDir+"/out.plan"))
	if err != nil {
		log.Fatalf("error running Plan: %s", err)
	}

	if pass {
		plan, err := tf.ShowPlanFile(context.Background(), workingDir+"/out.plan")
		if err != nil {
			log.Fatalf("error running fetch plan: %s", err)
		}

		for _, s := range plan.ResourceChanges {
			fmt.Printf("Change: %s %s\n", s.Change.Actions, s.Name)
		}

		fmt.Println("Applying changes")
		err2 := tf.Apply(context.Background())

		if err2 != nil {
			log.Fatalf("error running apply: %s", err2)
		}

	} else {
		fmt.Println("Destroying changes")
		err := tf.Destroy(context.Background())
		if err != nil {
			log.Fatalf("error running destroy: %s", err)
		}
	}

	//handleRequests()
}
