package cachingstoragewithqueue_test

import (
	"testing"
	"time"
)

func TestDateTime(t *testing.T) {
	var myTimeExpiry = 300

	currentNow := time.Now()
	customTime := currentNow.Add(time.Duration(myTimeExpiry) * time.Second)

	t.Logf("current time:'%v', custom time:'%v'\n", currentNow, customTime)
}
