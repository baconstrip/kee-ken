package main

import (
    "flag"
    "log"
    "os"

    "github.com/baconstrip/kiken/game"
    "github.com/baconstrip/kiken/server"
)

var (
    flagStaticPath = flag.String("static-path", "", "Path to static content files, must be set")
    flagTemplatesPath = flag.String("template-path", "", "Path to web templates, must be set")
    flagPort = flag.Int("port", 1986, "Port for the server to listen on")
    flagQuestionsList = flag.String("question-list", "", "Path to list of questions, must be set")
)

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

    log.Printf("Starting Kiken server on port %v", *flagPort)
    s := server.New(*flagTemplatesPath, *flagStaticPath, *flagPort)

    log.Fatal(s.ListenAndServe())
}
