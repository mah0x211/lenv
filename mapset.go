package main

type StringSet map[string]struct{}

func (ss StringSet) Value() (arr []string) {
	for k := range ss {
		arr = append(arr, k)
	}
	return
}

func (ss StringSet) Set(v string) {
	ss[v] = struct{}{}
}

func (ss StringSet) Delete(v string) {
	delete(ss, v)
}

type MapStringSet map[string]StringSet

func (mss MapStringSet) Set(k, v string) {
	if ss, ok := mss[k]; ok {
		ss.Set(v)
	} else {
		ss = StringSet{}
		ss.Set(v)
		mss[k] = ss
	}
}

func (mss MapStringSet) Delete(k, v string) {
	if ss, ok := mss[k]; ok {
		ss.Delete(v)
		if len(ss) == 0 {
			delete(mss, k)
		}
	}
}
