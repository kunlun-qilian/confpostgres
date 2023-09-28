package confpostgres

import (
	"os"
	"time"

	"github.com/go-courier/envconf"
	"golang.org/x/exp/slog"
)

type Retry struct {
	Repeats  int
	Interval envconf.Duration
}

func (r *Retry) SetDefaults() {
	if r.Repeats == 0 {
		r.Repeats = 3
	}
	if r.Interval == 0 {
		r.Interval = envconf.Duration(10 * time.Second)
	}
}

func (r Retry) Do(exec func() error) (err error) {
	if r.Repeats <= 0 {
		err = exec()
		return
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	for i := 0; i < r.Repeats; i++ {
		err = exec()
		if err != nil {
			log.Warn("retry in seconds [%v]", r.Interval)
			time.Sleep(time.Duration(r.Interval))
			continue
		}
		break
	}
	return
}
