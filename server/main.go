package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/baconstrip/kiken/editor"
	"github.com/baconstrip/kiken/game"
	"github.com/baconstrip/kiken/question"
	"github.com/baconstrip/kiken/server"
)

var (
	flagStaticPath    = flag.String("static-path", "../webcontent/static", "Path to static content files")
	flagTemplatesPath = flag.String("template-path", "../webcontent/templates", "Path to web templates")
	flagPort          = flag.Int("port", 1986, "Port for the server to listen on")
	// flagQuestionsList  = flag.String("question-list", "", "Path to list of questions, must be set")
	flagQuestionSource = flag.String("question-source", "", "Path to source for questions")
	flagPasscode       = flag.String("passcode", "test", "Passcode to use to grant admin privledges")
	flagDataDir        = flag.String("data-dir", "../data", "Path to location to store shows")
)

// TODO put this somewhere else
// expandPath expands a given path by expanding "~" to the user's home directory
// and resolving "." and ".." to a full absolute path.
func expandPath(p string) (string, error) {
	// Expand "~" to the home directory
	if strings.HasPrefix(p, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(homeDir, p[1:]) // Replace ~ with home directory
	}

	// Resolve "." and ".." and convert the path to absolute
	absPath, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}

	return absPath, nil
}

// TODO put this somewhere else
// countFilesInDir counts the number of files in a given directory
func countFilesInDir(dir string) (int, error) {
	// Read the directory using os.ReadDir (Go 1.16+)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	// Count files (not directories)
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() { // Only count regular files
			count++
		}
	}

	return count, nil
}

func main() {
	flag.Parse()

	log.Printf("Loading questions...")
	q, err := question.LoadQuestions(*flagQuestionSource)
	if err != nil {
		log.Fatalf("Could not load questions data: %v", err)
	}
	log.Printf("Finished loading questions")

	dataDir, err := expandPath(*flagDataDir)
	if err != nil {
		log.Fatalf("Could not open data dir: %v", err)
	}

	// Assign the global dataDir for the editor
	editor.DataDir = dataDir

	dataFileCount, err := countFilesInDir(dataDir)
	if err != nil {
		log.Fatalf("Error counting data files: %v", err)
	}
	log.Printf("Found %v files in the saved data.", dataFileCount)

	log.Printf("Starting Kiken server on port %v", *flagPort)

	gameLm := server.NewListenerManager()
	globalLm := server.NewListenerManager()

	s := server.New(*flagTemplatesPath, *flagStaticPath, *flagPasscode, *flagPort, globalLm, gameLm)

	metagame := game.NewMetaGameDriver(q, s, gameLm, globalLm)
	metagame.Start()

	log.Fatal(s.ListenAndServe())
}
