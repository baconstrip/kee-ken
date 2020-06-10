package main

import (
    "flag"
    "log"
    "time"
    "os"

    "github.com/kr/pretty"

    "github.com/baconstrip/kiken/game"
    "github.com/baconstrip/kiken/server"
)

var (
    flagStaticPath = flag.String("static-path", "", "Path to static content files, must be set")
    flagTemplatesPath = flag.String("template-path", "", "Path to web templates, must be set")
    flagPort = flag.Int("port", 1986, "Port for the server to listen on")
    flagQuestionsList = flag.String("question-list", "", "Path to list of questions, must be set")
    flagPasscode = flag.String("passcode", "test", "Passcode to use to grant admin privledges")
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

    standardCategories, err := game.CollateFullCategories(q)
    if err != nil {
        log.Printf("Failed to create categories from questions: %v", err)
    }
    owariCategories := game.CollateLoneQuestions(q, game.OWARI)
    tiebreakerCategories := game.CollateLoneQuestions(q, game.TIEBREAKER)

    log.Printf("Loaded %v standard categories, %v Owari, %v Tiebreaker.", len(standardCategories), len(owariCategories), len(tiebreakerCategories))

    // For testing, create a board of the first 5 categories from daiichi/daini,
    // and a question from owari.
    daiichiCount, dainiCount := 0, 0
    var daiichiCats, dainiCats []*game.Category

    for _, c := range standardCategories{
        if daiichiCount == 5 && dainiCount == 5 {
            break
        }

        if daiichiCount < 5  && c.Round == game.DAIICHI {
            daiichiCats = append(daiichiCats, c)
            daiichiCount++
        }
        if dainiCount < 5  && c.Round == game.DAINI {
            dainiCats = append(dainiCats, c)
            dainiCount++
        }
    }

    daiichiBoard := game.NewBoard(game.DAIICHI, daiichiCats...)
    dainiBoard := game.NewBoard(game.DAINI, dainiCats...)
    owariBoard := game.NewBoard(game.OWARI, owariCategories[0])

    g := game.New(daiichiBoard, dainiBoard, owariBoard)
    log.Printf("Created test board: ")
//    pretty.Print(g)
    pretty.Print(owariBoard)
    log.Printf("creating stateful game")

    gState := g.CreateState()
 //   pretty.Print(gState)

    lm := server.NewListenerManager()

    log.Printf("Starting Kiken server on port %v", *flagPort)
    s := server.New(*flagTemplatesPath, *flagStaticPath, *flagPasscode, *flagPort, lm)

    config := game.Configuration{
        ChanceTime: 5*time.Second,
        DisambiguationTime: 200*time.Millisecond,
        AnswerTime: 10*time.Second,
    }

    driver := game.NewGameDriver(s, gState, lm, config)
    go driver.Run()

    log.Fatal(s.ListenAndServe())
}
