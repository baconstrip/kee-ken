package game

import (
	"fmt"
	"math"
)

var potentialValueRanges = [][]int{
	{100, 200, 300, 400, 500},
	{200, 400, 600, 800, 1000},
	{400, 800, 1200, 1600, 2000},
}

// FixValues corrects duplicate values, fixes values that sit outside the
// normal procession, and changes the values so that they match the numbers
// used in normal play.
func FixValues(cat *Category) error {
	if err := inferValues(cat); err != nil {
		return err
	}
	return targetRoundValues(cat)
}

// inferValues computes the value of a question to match standard values, in the
// case that there questions with abnormal values. Modifies the questions in
// place. Errors if it is unable to determine values for the category.
func inferValues(cat *Category) error {
	if len(cat.Questions) != 5 {
		return fmt.Errorf("can only infer values when a category has exactly 5 questions, got %v", len(cat.Questions))
	}

	if err := deduplicate(cat); err != nil {
		return err
	}
	return normalize(cat)
}

// estimateValueRange attempts to determine the series of values this Category
// has.
func estimateValueRange(cat *Category) ([]int, error) {
	buckets := make(map[int][]*Question)
	for _, q := range cat.Questions {
		buckets[q.Value] = append(buckets[q.Value], q)
	}

	for _, r := range potentialValueRanges {
		matches := 0
		for _, q := range cat.Questions {
			for _, v := range r {
				if q.Value == v {
					matches = matches + 1
				}
			}
		}
		if matches >= 4 {
			return r, nil
		}
	}
	return nil, fmt.Errorf("could not infer value range for category, too may values that are not standard")
}

// deduplicate attempts to correct categories that have questions with duplicate
// values. Modifies the questions in place.
func deduplicate(cat *Category) error {
	buckets := make(map[int][]*Question)
	for _, q := range cat.Questions {
		buckets[q.Value] = append(buckets[q.Value], q)
	}

	var dupe []*Question
	for _, questions := range buckets {
		if len(questions) > 2 {
			return fmt.Errorf("categories values cannot be inferred, to many questions outlier")
		}
		if len(questions) == 2 && dupe != nil {
			return fmt.Errorf("category has too many duplicates, cannot infer values")
		}
		if len(questions) == 2 {
			dupe = questions
		}
	}

	if dupe == nil {
		return nil
	}

	expectedValues, err := estimateValueRange(cat)
	if err != nil {
		return err
	}

	missingValue := 0
	for _, v := range expectedValues {
		if _, ok := buckets[v]; !ok {
			missingValue = v
		}
	}
	dupe[1].Value = missingValue
	return nil
}

// normalize attempts to correct categories that have questions that aren't in
// line with a standard range.
func normalize(cat *Category) error {
	buckets := make(map[int][]*Question)
	for _, q := range cat.Questions {
		buckets[q.Value] = append(buckets[q.Value], q)
	}

	expectedValues, err := estimateValueRange(cat)
	if err != nil {
		return err
	}

	missingValue := 0
	for _, v := range expectedValues {
		if _, ok := buckets[v]; !ok {
			missingValue = v
		}
	}

	for v, b := range buckets {
		found := false
		for _, val := range expectedValues {
			if v == val {
				found = true
			}
		}
		if found {
			continue
		}
		b[0].Value = missingValue
	}

	return nil
}

func doubleValues(cat *Category) {
	for _, q := range cat.Questions {
		q.Value = q.Value * 2
	}
}

// targetRoundValues attempts to adjust the values of a category to match the
// values for the round its in.
func targetRoundValues(cat *Category) error {
	if len(cat.Questions) != 5 {
		return fmt.Errorf("can only target values when a category has exactly 5 questions, got %v", len(cat.Questions))
	}

	round := cat.Questions[0].Round

	// Since the questions have already been fixed, we can take the lowest value
	// question and use it as a basis.
	min := math.MaxInt32
	for _, q := range cat.Questions {
		if q.Value < min {
			min = q.Value
		}
	}

	// If they're already correct, we're done.
	if round == DAIICHI && min == 200 || round == DAINI && min == 400 {
		return nil
	}

	if round == DAIICHI && min == 100 || round == DAINI && min == 200 {
		doubleValues(cat)
		return nil
	}

	return fmt.Errorf("unable to determine correct range for values in category")
}
