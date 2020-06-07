package game

import (
    "fmt"
    "log"
)

// inferValue computes the value of a question to match standard values, in the
// case that there questions with abnormal values. Modifies the questions in
// place. Errors if it is unable to determine values for the category. In the 
// case two questions have the same value, a random one is changed.
func inferValue(cat *Category) error {
    // TODO come back to this, it's broken
    if len(cat.Questions) != 5 {
        return fmt.Errorf("can only infer values when a category has exactly 5 questions, got %v", len(cat.Questions))
    }

    buckets := make(map[int][]*Question)
    for _, q := range cat.Questions{
        buckets[q.Value] = append(buckets[q.Value], q)
    }

    var outlier *Question
    var dupe []*Question
    for _, questions := range buckets {
        if len(questions) > 2 {
            return fmt.Errorf("categories values cannot be inferred, to many questions outlier")
        }
        if len(questions) == 2 {
            dupe = questions
        }
    }

    if dupe != nil {
        dedupe := dupe[1]

        buckets[dupe[0].Value] = []*Question{dedupe}
        outlier = dupe[0]
    }

    var possibleValues [][]int
    // TODO support non-standard values via configuration.
    if cat.Questions[0].Round == ICHIBAN {
        possibleValues = [][]int{[]int{200, 400, 600, 800, 1000}, []int{100, 200, 300, 400, 500}}
    } else if cat.Questions[0].Round == NIBAN {
        possibleValues = [][]int{[]int{400, 800, 1200, 1600, 2000}, []int{200, 400, 600, 800, 1000}}
    } else {
        return fmt.Errorf("can only be used on standard questions")
    }

    bad := false
Solve:
    for _, poss := range possibleValues {
        foundValues := []bool{false, false, false, false, false}
        if outlier == nil {
            invalidValue := 0
            for val, _ := range buckets {
                found := false
                for i, v := range poss {
                    if val == v {
                        found = true
                        foundValues[i] = true
                        break
                    }
                }

                if !found {
                    if invalidValue != 0 {
                        if bad {
                            return fmt.Errorf("too many outliers, cannot infer values, buckets: %v, round %v", buckets, buckets[val][0].Round)
                        }
                        bad = true
                        continue Solve

                    }
                    invalidValue = val
                }
            }

            allPresent := true
            for _, found := range foundValues {
                if !found {
                    allPresent = false
                    break
                }
            }
            if allPresent {
                return nil
            }
        }
        log.Printf("Found outlier: %v, possible values: %v", outlier, poss)


        for i, found := range foundValues {
            if !found {
                outlier.Value = poss[i]
                return nil
            }
        }
    }
    return nil
}

