package test

import (
	"fmt"
	"testing"

	"github.com/supergodk/go-utils/v1/timeutil"
)

func Test(t *testing.T) {
	a, err := timeutil.GetDurationSeconds("01:30:45")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(a)
}
