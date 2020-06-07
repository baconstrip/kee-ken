package game

import (
    "testing"
    "reflect"

    "github.com/kr/pretty"
)

var testData = []struct{
    name string
    category Category
    want Category
    wantErr bool
} {
    {
        name: "Doesn't affect questions with already appropriate values",
        category: Category{
            Name: "TestCategory",
            Round: ICHIBAN,
            Questions: []*Question{
                &Question{
                    Category: "TestCategory",
                    Value: 200,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 400,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 600,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 800,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 1000,
                    Round: ICHIBAN,
                },
            },
        },
        want: Category{
            Name: "TestCategory",
            Round: ICHIBAN,
            Questions: []*Question{
                &Question{
                    Category: "TestCategory",
                    Value: 200,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 400,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 600,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 800,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 1000,
                    Round: ICHIBAN,
                },
            },
        },
    },
    {
        name: "Corrects one incorrect value in ICHIBAN",
        category: Category{
            Name: "TestCategory",
            Round: ICHIBAN,
            Questions: []*Question{
                &Question{
                    Category: "TestCategory",
                    Value: 200,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 400,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 2000,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 800,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 1000,
                    Round: ICHIBAN,
                },
            },
        },
        want: Category{
            Name: "TestCategory",
            Round: ICHIBAN,
            Questions: []*Question{
                &Question{
                    Category: "TestCategory",
                    Value: 200,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 400,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 600,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 800,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 1000,
                    Round: ICHIBAN,
                },
            },
        },
    },
    {
        name: "Corrects duplicates",
        category: Category{
            Name: "TestCategory",
            Round: ICHIBAN,
            Questions: []*Question{
                &Question{
                    Category: "TestCategory",
                    Value: 200,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 400,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 600,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 600,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 1000,
                    Round: ICHIBAN,
                },
            },
        },
        want: Category{
            Name: "TestCategory",
            Round: ICHIBAN,
            Questions: []*Question{
                &Question{
                    Category: "TestCategory",
                    Value: 200,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 400,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 600,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 800,
                    Round: ICHIBAN,
                },
                &Question{
                    Category: "TestCategory",
                    Value: 1000,
                    Round: ICHIBAN,
                },
            },
        },
    },
}

func TestInferValues(t *testing.T) {
    for _, test := range testData {
        t.Run(test.name, func(t *testing.T) {
            got := test.category
            err := inferValue(&got)
            if err != nil {
                t.Errorf("Error while inferring values: %v", err)
            }
            if !reflect.DeepEqual(got, test.want) {
                t.Errorf("Infer values did not produce expected output, got:\n\n %v\n\nwant: %v\n\n", pretty.Sprint(got), pretty.Sprint(test.want))
            }
        })
    }
}
