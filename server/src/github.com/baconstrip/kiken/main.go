package main

import (
    "flag"
    "log"
    "os"
    "strconv"
    "path/filepath"
    "net/http"
    "html/template"

    "github.com/baconstrip/kiken/game"
)

var (
    flagStaticPath = flag.String("static-path", "", "Path to static content files, must be set")
    flagTemplatesPath = flag.String("template-path", "", "Path to web templates, must be set")
    flagPort = flag.Int("port", 1986, "Port for the server to listen on")
    flagQuestionsList = flag.String("question-list", "", "Path to list of questions, must be set")
)

type templateGroup struct {
    index *template.Template
}

func (t templateGroup) indexHandler(w http.ResponseWriter, r *http.Request) {
    err := t.index.Execute(w, nil)
    if err != nil {
        log.Printf("Error loading index page: %v", err)
    }
}

func main() {
    flag.Parse()

    if *flagStaticPath == "" || *flagTemplatesPath == "" || *flagQuestionsList == "" {
        flag.PrintDefaults()
        os.Exit(1)
    }

    log.Printf("Loading questions...")
    q, err := game.LoadQuestions(*flagQuestionsList)
    if err != nil {
        log.Fatalf("Could not load questions data: %v", err)
    }
    log.Printf("Finished loading questions")
    _ = q

    tg := templateGroup{}
    tg.index = template.Must(template.ParseFiles(filepath.Join(*flagTemplatesPath, "index.html")))

    mux := http.NewServeMux()
    mux.HandleFunc("/", tg.indexHandler)
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(*flagStaticPath))))

    log.Printf("Starting Kiken server on port %v", *flagPort)
    log.Fatal(http.ListenAndServe(":" + strconv.Itoa(*flagPort), mux))
}
