package game

import (
    "fmt"
    "log"
    "strconv"
    "strings"
    "regexp"
    "encoding/json"
    "io/ioutil"
)

var valueRegexp = regexp.MustCompile("^[$]?([0-9,]*)$|^[Nn][Oo][Nn][Ee]$|^$")

type Round int

const (
    UNKNOWN Round = iota
    ICHIBAN
    NIBAN
    OWARI
    TIEBREAKER
)

type Questions struct {}

type Question struct {
    Category string
    Value int
    Question string
    Answer string
    Round Round
    Showing int
}

// LoadQuestsions reads the contents of the file at path as JSON and
// tries to interpret it as question data.
func LoadQuestions(path string) (*Questions, error) {
    contents, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var data interface{}
    if err = json.Unmarshal(contents, &data); err != nil {
        return nil, fmt.Errorf("error unmarshaling questions JSON: %v", err)
    }

    questions, err := decodeQuestions(&data)
    if err != nil {
        return nil, fmt.Errorf("error decoding question data: %v", err)
    }
    _ = questions
    return nil, nil
}


func decodeQuestions(i *interface{}) ([]*Question, error) {
    var retVal []*Question
    switch (*i).(type) {
    case []interface{}:
    default:
        return nil, fmt.Errorf("expecting a list of questions as a JSON array")
    }
    qs := (*i).([]interface{})
    for _, v := range qs {
        switch v.(type) {
        case map[string]interface{}:
        default:
            log.Printf("questions should be a dictionary, skipping bad question, found: %+v", v)
            continue
        }
        q := v.(map[string]interface{})

        category, err := parseCategory(q)
        if err != nil {
            log.Printf("Discarding question without category, err: %v: %+v", err, q)
            continue
        }

        value, err := parseValue(q)
        if err != nil {
            log.Printf("Discarding question bad value err: %v: %v", err, q)
            continue
        }

        prompt, err := parseQuestion(q)
        if err != nil {
            log.Printf("Discarding question without question, err: %v: %+v", err, q)
            continue
        }

        answer, err := parseAnswer(q)
        if err != nil {
            log.Printf("Discarding question without answer, err: %v: %+v", err, q)
            continue
        }

        round, err := parseRound(q)
        if err != nil {
            log.Printf("Discarding question with bad round, err: %v: %+v", err, q)
            continue
        }

        showing, err := parseShowing(q)
        if err != nil {
            log.Printf("Discarding question with bad showing, err: %v: %+v", err, q)
            continue
        }

        question := &Question{
            Category: category,
            Value: value,
            Question: prompt,
            Answer: answer,
            Round: round,
            Showing: showing,
        }

        retVal = append(retVal, question)
    }
    return retVal, nil
}

func parseCategory(q map[string]interface{}) (string, error) {
    if _, ok := q["category"]; !ok {
        return "", fmt.Errorf("question has no cateory")
    }
    cat := q["category"]
    switch t := cat.(type) {
    case string:
        return cat.(string), nil
    case int:
        return strconv.Itoa(cat.(int)), nil
    case float64:
        return strconv.FormatFloat(cat.(float64), 'f', 6, 64), nil
    default:
        return "", fmt.Errorf("bad type parsing category, got %v, expected string", t)
    }
}

func parseValue(q map[string]interface{}) (int, error) {
    // If the question has no value, just set it to zero.
    if _, ok := q["value"]; !ok {
        return 0, nil
    }
    val := q["value"]
    switch t := val.(type) {
    case string:
        str := val.(string)
        if !valueRegexp.MatchString(str) {
            return -1, fmt.Errorf("failed to parse value, should be a number with dollar sign, empty, or 'None': %v", str)
        }
        match := valueRegexp.FindStringSubmatch(str)
        if len(match) == 1 || match[1] == "" || strings.ToLower(match[1]) == "none" {
            return 0, nil
        }
        i, err := strconv.Atoi(strings.Replace(match[1], ",", "", -1))
        if err != nil {
            return -1, fmt.Errorf("failed to parse number despite matching numeric regexp: %v, %v", match[1], err)
        }
        return i, nil
    case int:
        return val.(int), nil
    case float64:
        return int(val.(float64)), nil
    case nil:
        return 0, nil
    default:
        return -1, fmt.Errorf("bad type parsing value, got %T, expected string or int", t)
    }
}

func parseQuestion(q map[string]interface{}) (string, error) {
    if _, ok := q["question"]; !ok {
        return "", fmt.Errorf("question has no prompt")
    }
    ques := q["question"]
    switch t := ques.(type) {
    case string:
        return ques.(string), nil
    case int:
        return strconv.Itoa(ques.(int)), nil
    case float64:
        return strconv.FormatFloat(ques.(float64), 'f', 6, 64), nil
    default:
        return "", fmt.Errorf("bad type parsing question, got %T, expected string", t)
    }
}

func parseAnswer(q map[string]interface{}) (string, error) {
    if _, ok := q["answer"]; !ok {
        return "", fmt.Errorf("question has no answer")
    }
    ans := q["answer"]
    switch t := ans.(type) {
    case string:
        return ans.(string), nil
    case int:
        return strconv.Itoa(ans.(int)), nil
    case float64:
        return strconv.FormatFloat(ans.(float64), 'f', 6, 64), nil
    default:
        return "", fmt.Errorf("bad type parsing answer, got %T, expected string", t)
    }
}

func parseRound(q map[string]interface{}) (Round, error) {
    if _, ok := q["round"]; !ok {
        return UNKNOWN, fmt.Errorf("question has no round")
    }
    round := q["round"]
    switch t := round.(type) {
    case string:
        r := strings.TrimSpace(strings.ToLower(round.(string)))
        if strings.HasPrefix(r, "j") {
            return ICHIBAN, nil
        }
        if strings.HasPrefix(r, "d") {
            return NIBAN, nil
        }
        if strings.HasPrefix(r, "f") {
            return OWARI, nil
        }
        if r == "ichiban" {
            return ICHIBAN, nil
        }
        if r == "niban" {
            return NIBAN, nil
        }
        if r == "owari" {
            return OWARI, nil
        }
        if strings.HasPrefix(r, "tiebreaker") {
            return TIEBREAKER, nil
        }
        return UNKNOWN, nil
    case int:
        return (Round)(round.(int)), nil
    case float64:
        return (Round)(int(round.(float64))), nil
    default:
        return UNKNOWN, fmt.Errorf("bad type reading round, should be a word representing the round, %v", t)
    }
}

func parseShowing(q map[string]interface{}) (int, error) {
    if _, ok := q["show_number"]; !ok {
        return -1, nil
    }
    show := q["show_number"]
    switch t := show.(type) {
    case string:
        return strconv.Atoi(show.(string))
    case int:
        return show.(int), nil
    case float64:
        return int(show.(float64)), nil
    default:
        return -1, fmt.Errorf("bad type parsing showing, got %T, expected number", t)
    }
}
