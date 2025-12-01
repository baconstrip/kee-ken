package main

import (
	"flag"
	"log"
	"strings"

	"github.com/baconstrip/kiken/editor"
	"github.com/baconstrip/kiken/game"
	"github.com/baconstrip/kiken/question"
	"github.com/baconstrip/kiken/server"
	"github.com/baconstrip/kiken/util"
)

var (
	flagStaticPath    = flag.String("static-path", "../webcontent/static", "Path to static content files")
	flagTemplatesPath = flag.String("template-path", "../webcontent/templates", "Path to web templates")
	flagPort          = flag.Int("port", 1986, "Port for the server to listen on")
	// flagQuestionsList  = flag.String("question-list", "", "Path to list of questions, must be set")
	flagQuestionSource = flag.String("question-source", "", "Path to source for questions")
	flagPasscode       = flag.String("passcode", "test", "Passcode to use to grant admin privledges")
	flagDataDir        = flag.String("data-dir", "../data", "Path to location to store shows")
	flagStartAt        = flag.String("start-at", "", "If set, the server will start the game at the specified stage, for testing purposes.")
)

var validStartStage map[string]interface{} = map[string]interface{}{
	"owari":   nil,
	"daiichi": nil,
	"daini":   nil,
}

func main() {
	flag.Parse()

	log.Printf("Loading questions...")
	q, err := question.LoadQuestions(*flagQuestionSource)
	if err != nil {
		log.Fatalf("Could not load questions data: %v", err)
	}
	log.Printf("Finished loading questions")

	dataDir, err := util.ExpandPath(*flagDataDir)
	if err != nil {
		log.Fatalf("Could not open data dir: %v", err)
	}

	if *flagStartAt != "" {
		*flagStartAt = strings.ToLower(string([]byte(*flagStartAt)))
		if _, ok := validStartStage[*flagStartAt]; !ok {
			log.Fatalf("Invalid start-at stage specified: %v", *flagStartAt)
		}
		log.Printf("Overriding server start at stage: %v", *flagStartAt)
	} else {
		*flagStartAt = "daiichi"
	}

	// Assign the global dataDir for the editor
	editor.DataDir = dataDir

	dataFileCount, err := util.CountFilesInDir(dataDir)
	if err != nil {
		log.Fatalf("Error counting data files: %v", err)
	}
	log.Printf("Found %v files in the saved data.", dataFileCount)

	log.Printf("Starting Kiken server on port %v", *flagPort)

	gameLm := server.NewListenerManager()
	globalLm := server.NewListenerManager()
	editorLm := server.NewListenerManager()

	s := server.New(*flagTemplatesPath, *flagStaticPath, *flagPasscode, *flagPort, globalLm, gameLm, editorLm)

	metagame := game.NewMetaGameDriver(q, s, *flagStartAt, gameLm, globalLm)
	metagame.Start()

	editor := editor.NewEditorDriver(s, editorLm)
	editor.Start()

	log.Fatal(s.ListenAndServe())
}
