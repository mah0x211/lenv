package main

import (
	"fmt"
	"sort"
	"strings"
)

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
			sort.Sort(sort.Reverse(sort.StringSlice(vers)))

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
