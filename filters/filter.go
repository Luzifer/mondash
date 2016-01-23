package filters

import "github.com/flosch/pongo2"

func init() {
	pongo2.RegisterFilter("lastNItems", filterLastNItems)
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
