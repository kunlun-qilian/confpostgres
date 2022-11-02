package confpostgres

import (
	"github.com/go-courier/envconf"
	"time"
)

type Retry struct {
	Repeats  int
	Interval envconf.Duration
}

func (r Retry) Do(exec func() error) (err error) {
	if r.Repeats <= 0 {
		err = exec()
		return
	}
	for i := 0; i < r.Repeats; i++ {
		err = exec()
		if err != nil {
			time.Sleep(time.Duration(r.Interval))
			continue
		}
		break
	}
	return
}
