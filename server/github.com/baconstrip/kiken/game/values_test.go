package game

import (
	"reflect"
	"testing"

	"github.com/kr/pretty"
)

var testData = []struct {
	name     string
	category Category
	want     Category
	wantErr  bool
}{
	{
		name: "Doesn't affect questions with already appropriate values",
		category: Category{
			Name:  "TestCategory",
			Round: DAIICHI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    200,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    400,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    600,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    800,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    1000,
					Round:    DAIICHI,
				},
			},
		},
		want: Category{
			Name:  "TestCategory",
			Round: DAIICHI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    200,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    400,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    600,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    800,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    1000,
					Round:    DAIICHI,
				},
			},
		},
	},
	{
		name: "Corrects one incorrect value in DAIICHI",
		category: Category{
			Name:  "TestCategory",
			Round: DAIICHI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    200,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    400,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    2000,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    800,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    1000,
					Round:    DAIICHI,
				},
			},
		},
		want: Category{
			Name:  "TestCategory",
			Round: DAIICHI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    200,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    400,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    600,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    800,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    1000,
					Round:    DAIICHI,
				},
			},
		},
	},
	{
		name: "Corrects duplicates",
		category: Category{
			Name:  "TestCategory",
			Round: DAIICHI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    200,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    400,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    600,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    600,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    1000,
					Round:    DAIICHI,
				},
			},
		},
		want: Category{
			Name:  "TestCategory",
			Round: DAIICHI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    200,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    400,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    600,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    800,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    1000,
					Round:    DAIICHI,
				},
			},
		},
	},
	{
		name: "Corrects one incorrect value",
		category: Category{
			Name:  "TestCategory",
			Round: DAIICHI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    200,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    400,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    600,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    700,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    1000,
					Round:    DAIICHI,
				},
			},
		},
		want: Category{
			Name:  "TestCategory",
			Round: DAIICHI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    200,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    400,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    600,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    800,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    1000,
					Round:    DAIICHI,
				},
			},
		},
	},
	{
		name: "Corrects one incorrect value in DAINI",
		category: Category{
			Name:  "TestCategory",
			Round: DAINI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    400,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    800,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    1200,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    300,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    2000,
					Round:    DAINI,
				},
			},
		},
		want: Category{
			Name:  "TestCategory",
			Round: DAINI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    400,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    800,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    1200,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    1600,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    2000,
					Round:    DAINI,
				},
			},
		},
	},
	{
		name: "Corrects one incorrect value in DAINI with different value range, out of order",
		category: Category{
			Name:  "TestCategory",
			Round: DAINI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    200,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    600,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    800,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    1000,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    1280,
					Round:    DAINI,
				},
			},
		},
		want: Category{
			Name:  "TestCategory",
			Round: DAINI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    200,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    600,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    800,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    1000,
					Round:    DAINI,
				},
				{
					Category: "TestCategory",
					Value:    400,
					Round:    DAINI,
				},
			},
		},
	},
	{
		name: "Corrects one incorrect value in DAIICHI with different value range",
		category: Category{
			Name:  "TestCategory",
			Round: DAIICHI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    100,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    200,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    1000,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    400,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    500,
					Round:    DAIICHI,
				},
			},
		},
		want: Category{
			Name:  "TestCategory",
			Round: DAIICHI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    100,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    200,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    300,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    400,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    500,
					Round:    DAIICHI,
				},
			},
		},
	},
	{
		name: "Errors with too many questions",
		category: Category{
			Name:  "TestCategory",
			Round: DAIICHI,
			Questions: []*Question{
				{
					Category: "TestCategory",
					Value:    100,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    200,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    1000,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    400,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    500,
					Round:    DAIICHI,
				},
				{
					Category: "TestCategory",
					Value:    800,
					Round:    DAIICHI,
				},
			},
		},
		wantErr: true,
	},
}

func TestInferValues(t *testing.T) {
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got := test.category
			err := inferValues(&got)
			if err != nil && !test.wantErr {
				t.Errorf("Error while inferring values: %v", err)
			}
			if test.wantErr && err == nil {
				t.Fatalf("did not get error when expecting one")
			}
			if !test.wantErr && !reflect.DeepEqual(got, test.want) {
				t.Errorf("Infer values did not produce expected output, got:\n\n %v\n\nwant: %v\n\n", pretty.Sprint(got), pretty.Sprint(test.want))
			}
		})
	}
}
