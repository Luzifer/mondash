package filters

import (
	"time"

	humanize "github.com/flosch/go-humanize"
	"github.com/flosch/pongo2"
)

func init() {
	pongo2.RegisterFilter("lastNItems", filterLastNItems)
	pongo2.RegisterFilter("naturaltime", filterTimeuntilTimesince)
}

func filterLastNItems(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if !in.CanSlice() {
		return in, nil
	}

	from := in.Len() - param.Integer()
	if from < 0 {
		from = 0
	}

	return in.Slice(from, in.Len()), nil
}

func filterTimeuntilTimesince(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	basetime, isTime := in.Interface().(time.Time)
	if !isTime {
		return nil, &pongo2.Error{
			Sender:   "filter:timeuntil/timesince",
			ErrorMsg: "time-value is not a time.Time-instance",
		}
	}
	var paramtime time.Time
	if !param.IsNil() {
		paramtime, isTime = param.Interface().(time.Time)
		if !isTime {
			return nil, &pongo2.Error{
				Sender:   "filter:timeuntil/timesince",
				ErrorMsg: "time-parameter is not a time.Time-instance",
			}
		}
	} else {
		paramtime = time.Now()
	}

	return pongo2.AsValue(humanize.TimeDuration(basetime.Sub(paramtime))), nil
}
