package test

import (
	"fmt"
	"testing"

	"github.com/supergodk/go-utils/v1/timeutil"
)

func Test(t *testing.T) {
	a := timeutil.ClockTickMicroSecondUniq()
	fmt.Println(a)
}
