package main

import (
    "os"
    "io/ioutil"
    /*"encoding/json"*/
    "fmt"
    "regexp"
	"html/template"
	"log"
	"net/http"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	
)

type Entry struct {
    ErrorMsg string
    SearchResult string
	Esp string
	Eng    string
}
/*
type Pronouns struct {
    Prons []string `json:"pronouns"`
}*/


func index(writer http.ResponseWriter, request *http.Request) {
	// message := []byte("Hello, web!")
	tmpl, err := os.Open("template/index.htm")
	check(err)
	
	tmpl_content , err := ioutil.ReadAll(tmpl)
	check(err)
	fmt.Fprint(writer, string(tmpl_content))

	// ins := Ins{author: "Tan Kian-ting", age: 30}
	// err = tmpl.Execute()
	// _, err := writer.Write(message)
	check(err)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func contain(slice []string, str string) bool{
    for _, item := range slice{
        if item == str{
            return true
        }
    }
    return false
}

func get_lemma(str string) string{
    verb_ptn := regexp.MustCompile(`(?P<Stem>[\-a-zA-Z]+?)(?P<Suffix>((ant|int|ont)(an?|ajn?|on?|ojn?|e))|as|is|os|u|i|us)$`)
    noun_ptn := regexp.MustCompile(`(?P<Stem>[\-a-zA-Z]+?)(?P<Suffix>(on?|ojn?))$`)    
    adj_ptn := regexp.MustCompile(`(?P<Stem>[\-a-zA-Z]+?)(?P<Suffix>(an?|ajn?))$`)    

    var lemma string
    verb_match :=  verb_ptn.FindStringSubmatch(str)
    noun_match :=  noun_ptn.FindStringSubmatch(str)
    adj_match :=  adj_ptn.FindStringSubmatch(str)
    
    if verb_match != nil {
    lemma = verb_match[1] + "i"
    } else if noun_match != nil {
    lemma = noun_match[1] + "o"
    } else if adj_match != nil {
    lemma = adj_match[1] + "a"
    } else{
    lemma = str
    }
    
    return lemma

}


func result_esp(result string, writer http.ResponseWriter, request *http.Request){

	db, err := sql.Open("sqlite3", "data/dict.db")
	check(err)
	
    
    var pron_slice []string
    
    res, err := db.Query("SELECT * FROM Pronoun")
    check(err)
    
    for res.Next(){
        var esp_pron string
        err = res.Scan(&esp_pron)
        pron_slice = append(pron_slice, esp_pron)
        
    }
    
    
    var lemma string
    
    if contain(pron_slice, result) != true{
        lemma = get_lemma(result) // estas -> esti
    	//fmt.Println(lemma)
    // if it's a pronoun, let lemma as a result
    } else{
        lemma = result
    }
    	
	
	res, err = db.Query("SELECT * FROM Dict where Esperanto=?", lemma)
	check(err)

    var entry Entry
           
	entry.ErrorMsg = "找不到 " + result + " 的詞條，請重新尋找。" // preset it.
	for res.Next() {
	    var esperanto string
        var english string
        err = res.Scan(&esperanto, &english)
        check(err)
        // total_result := esperanto + "\n" + english
        
        
        
        entry.SearchResult = result
       	entry.Esp = esperanto
	    entry.Eng = english
	    entry.ErrorMsg = ""
	    
    }
    
    	tmpl, err := template.ParseFiles("template/result.htm")
	    check(err)
	    err = tmpl.Execute(writer, entry)
        // _, err = writer.Write([]byte(total_result))
	    check(err)
}


func result(writer http.ResponseWriter, request *http.Request) {
	// message := []byte("Hello, web!")
	result := request.FormValue("word")
	lang := request.FormValue("lang_select")
	
	if lang == "esp"{
	    result_esp(result, writer, request)
	
	}
}

func main() {
	http.HandleFunc("/index.html", index)
	http.HandleFunc("/result.html", result)
	err := http.ListenAndServe("localhost:8080", nil)
	log.Fatal(err)
}
