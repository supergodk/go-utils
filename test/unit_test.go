package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/supergodk/go-utils/v1/timeutil"
)

func Test(t *testing.T) {
	first, last := timeutil.GetLastMonthTime(time.Now().Unix())
	fmt.Println(first, last)
}
