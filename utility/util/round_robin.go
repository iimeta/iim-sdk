package util

type RoundRobin struct {
	CurIndex int
}

func (r *RoundRobin) RoundRobinIndex(lens int) (index int) {

	if r.CurIndex >= lens {
		r.CurIndex = 0
	}

	index = r.CurIndex

	r.CurIndex = (r.CurIndex + 1) % lens

	return
}

func (r *RoundRobin) RoundRobinKey(keys []string) string {
	return keys[r.RoundRobinIndex(len(keys))]
}
