package pretty_test

import (
	"fmt"
	"testing"

	"github.com/avian-digital-forensics/auto-processing/pkg/pretty"
)

func TestHeader(t *testing.T) {
	fmt.Println(
		pretty.Format(
			pretty.Header("Greeting", "Name"),
			pretty.Body([]interface{}{"Hej", "Simon"}),
		),
	)
}
