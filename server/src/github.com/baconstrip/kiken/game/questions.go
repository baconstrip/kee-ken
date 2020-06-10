package game

import (
    "fmt"
    "log"
    "strconv"
    "strings"
    "regexp"
    "encoding/base64"
    "crypto/sha512"
    "encoding/json"
    "io/ioutil"
)

var valueRegexp = regexp.MustCompile("^[$]?([0-9,]*)$|^[Nn][Oo][Nn][Ee]$|^$")

type Question struct {
    Category string
    Value int
    Question string
    Answer string
    Round Round
    Showing int

    // ID contains the base64 encoded SHA512 encoding of the question 
    // text+category, used to uniquely identify the question.
    ID string
}

type Category struct {
    Name string
    Round Round
    Questions []*Question
}

// ByValue implements a type that allows sorting Questions by their value.
type ByValue []*Question

func (b ByValue) Len() int { return len(b) }
func (b ByValue) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ByValue) Less(i, j int) bool { return b[i].Value < b[j].Value }

// LoadQuestsions reads the contents of the file at path as JSON and
// tries to interpret it as question data.
func LoadQuestions(path string) ([]*Question, error) {
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
    return questions, nil
}

// CollateFullCategories groups questions first based on category, then cheks
// that there are exactly 5 questions available. It will ignore any categories
// that don't have 5 questions, after filtering for questions that are played
// in normal play (not tiebreakers nor Owari).
func CollateFullCategories(questions []*Question) ([]*Category, error) {
    var filteredQuestions []*Question
    for _, q := range questions {
        if q.Round == DAIICHI || q.Round == DAINI {
            filteredQuestions = append(filteredQuestions, q)
        }
    }

    categoryGroups := make(map[string][]*Question)
    for _, q := range filteredQuestions {
        categoryGroups[q.Category] = append(categoryGroups[q.Category], q)
    }

    filteredCategories := make(map[string][]*Question)

    for cat, q := range categoryGroups {
        if len(q) == 5 {
            filteredCategories[cat] = q
        }
    }
    log.Printf("Filtered for categories that contain 5 questions, discarded %v categories", len(categoryGroups)-len(filteredCategories))

    var categories []*Category
    for cat, q := range filteredCategories {
        categories = append(categories, &Category{Name: cat, Round: q[0].Round, Questions: q})
    }

    for _, cat := range categories {
        err := InferValues(cat)
        if err != nil {
            log.Printf("failed to infer value for category %v, %v", cat.Name, err)
        }
    }
    return categories, nil
}

// CollateLoneQuestions collects questions for the final rounds of play.
func CollateLoneQuestions(questions []*Question, r Round) []*Category {
    var filteredQuestions []*Question
    for _, q := range questions {
        if q.Round == r {
            filteredQuestions = append(filteredQuestions, q)
        }
    }

    var categories []*Category
    for _, q := range filteredQuestions {
        categories = append(categories, &Category{Name: q.Category, Round: r, Questions: []*Question{q}})
    }

    return categories
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

        hasher := sha512.New()
        hasher.Write([]byte(prompt+category))
        sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

        question := &Question{
            Category: category,
            Value: value,
            Question: prompt,
            Answer: answer,
            Round: round,
            Showing: showing,

            ID: sha,
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
            return DAIICHI, nil
        }
        if strings.HasPrefix(r, "d") {
            return DAINI, nil
        }
        if strings.HasPrefix(r, "f") {
            return OWARI, nil
        }
        if r == "daiichi" {
            return DAIICHI, nil
        }
        if r == "daini" {
            return DAINI, nil
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
