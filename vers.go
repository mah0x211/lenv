package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type version struct {
	major int
	minor int
	patch int
	tag   string
}

var reVersionNum = regexp.MustCompile(`^\d+$`)
var reVersionTag = regexp.MustCompile(`^(\d+)[.+-]?(.*)$`)

func toInt(s string) int {
	if reVersionNum.MatchString(s) {
		n, err := strconv.Atoi(s)
		if err != nil {
			fatalf("failed to strconv.Atoi(%#v): %v", s, err)
		}
		return n
	}
	return 0
}

func newVersion(s string) *version {
	list := strings.SplitN(s, ".", 3)
	for i, s := range list {
		if !reVersionNum.MatchString(s) {
			list[i] = strings.Join(list[i:], ".")
			break
		}
	}
	tail := len(list) - 1
	if tail >= 0 {
		matches := reVersionTag.FindStringSubmatch(list[tail])
		if len(matches) == 3 {
			list[tail] = matches[1]
			list = append(list, matches[2])
		}
	}

	ver := &version{}
	switch len(list) {
	case 4:
		ver.tag = list[3]
		fallthrough
	case 3:
		ver.patch = toInt(list[2])
		fallthrough
	case 2:
		ver.minor = toInt(list[1])
		fallthrough
	case 1:
		ver.major = toInt(list[0])
	}

	return ver
}

func sortVersions(vers []string) {
	sort.Slice(vers, func(i, j int) bool {
		a := newVersion(vers[i])
		b := newVersion(vers[j])
		if a.major > b.major {
			return true
		} else if a.major < b.major {
			return false
		} else if a.minor > b.minor {
			return true
		} else if a.minor < b.minor {
			return false
		} else if a.patch > b.patch {
			return true
		} else if a.patch < b.patch {
			return false
		} else if len(a.tag) == 0 && len(b.tag) > 0 {
			return true
		} else if len(b.tag) == 0 && len(a.tag) > 0 {
			return false
		}
		return a.tag > b.tag
	})
}

func cmdVers() {
	for _, cfg := range []*TargetConfig{
		LuaCfg, LuaJitCfg, LuaRocksCfg,
	} {
		printf("list available %q versions:", cfg.Name)
		if list, err := readVersionFile(cfg.VersionFile); err != nil {
			eprintf("failed to read version file: %v", err)
		} else {
			vers := []string{}
			maxlen := 0
			for v := range list {
				vers = append(vers, v)
				if len(v) > maxlen {
					maxlen = len(v)
				}
			}
			sortVersions(vers)

			format := fmt.Sprintf("%%-%ds", maxlen)
			arr := []string{}
			ncols := 0
			for _, v := range vers {
				v = fmt.Sprintf(format, v)
				ncols += maxlen + 2
				if ncols/80 > 0 {
					printf(strings.Join(arr, "  "))
					arr = arr[:0]
					ncols = maxlen + 2
				}
				arr = append(arr, v)
			}
			if len(arr) > 0 {
				printf(strings.Join(arr, "  "))
			}
		}
		printf("")
	}
}
