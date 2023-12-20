package main

import (
	"flag"
	"log"
	"os"

	"github.com/baconstrip/kiken/game"
	"github.com/baconstrip/kiken/server"
)

var (
	flagStaticPath    = flag.String("static-path", "", "Path to static content files, must be set")
	flagTemplatesPath = flag.String("template-path", "", "Path to web templates, must be set")
	flagPort          = flag.Int("port", 1986, "Port for the server to listen on")
	flagQuestionsList = flag.String("question-list", "", "Path to list of questions, must be set")
	flagPasscode      = flag.String("passcode", "test", "Passcode to use to grant admin privledges")
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

	log.Printf("creating stateful game")

	log.Printf("Starting Kiken server on port %v", *flagPort)

	gameLm := server.NewListenerManager()
	globalLm := server.NewListenerManager()

	s := server.New(*flagTemplatesPath, *flagStaticPath, *flagPasscode, *flagPort, globalLm, gameLm)

	metagame := game.NewMetaGameDriver(q, s, gameLm, globalLm)
	metagame.Start()

	log.Fatal(s.ListenAndServe())
}
