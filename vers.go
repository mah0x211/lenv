package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var reSemVer = regexp.MustCompile(`^\d+(?:\.\d+){1,2}`)

func IsSemVer(ver string) bool {
	return reSemVer.MatchString(ver)
}

var reVersionNum = regexp.MustCompile(`^\d+$`)

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

type SemVer struct {
	Major int
	Minor int
	Patch int
	Tag   string
}

var reVersionTag = regexp.MustCompile(`^(\d+)[.+-]?(.*)$`)

func NewSemVer(ver string) *SemVer {
	segments := strings.SplitN(ver, ".", 3)
	for i, s := range segments {
		if !reVersionNum.MatchString(s) {
			segments[i] = strings.Join(segments[i:], ".")
			break
		}
	}
	tail := len(segments) - 1
	if tail >= 0 {
		matches := reVersionTag.FindStringSubmatch(segments[tail])
		if len(matches) == 3 {
			segments[tail] = matches[1]
			segments = append(segments, matches[2])
		}
	}

	sv := &SemVer{}
	switch len(segments) {
	case 4:
		sv.Tag = segments[3]
		fallthrough
	case 3:
		sv.Patch = toInt(segments[2])
		fallthrough
	case 2:
		sv.Minor = toInt(segments[1])
		fallthrough
	case 1:
		sv.Major = toInt(segments[0])
	}

	return sv
}

func (a *SemVer) Greater(b *SemVer) bool {
	if a.Major > b.Major {
		return true
	} else if a.Major < b.Major {
		return false
	} else if a.Minor > b.Minor {
		return true
	} else if a.Minor < b.Minor {
		return false
	} else if a.Patch > b.Patch {
		return true
	} else if a.Patch < b.Patch {
		return false
	} else if len(a.Tag) == 0 && len(b.Tag) > 0 {
		return true
	} else if len(b.Tag) == 0 && len(a.Tag) > 0 {
		return false
	}
	return a.Tag > b.Tag
}

func SortVersions(vers []string) {
	sort.Slice(vers, func(i, j int) bool {
		a := NewSemVer(vers[i])
		b := NewSemVer(vers[j])
		return a.Greater(b)
	})
}

type VerItem struct {
	Name   string
	Remote string
	Ver    string
	Sum    string
	Ext    string
	SemVer *SemVer
}

func NewVerItem(name, remote, ver, sum, ext string) *VerItem {
	if remote != "" {
		return &VerItem{
			Name:   name,
			Remote: remote,
			Ver:    name,
		}
	}

	return &VerItem{
		Name:   name,
		Ver:    ver,
		Sum:    sum,
		Ext:    ext,
		SemVer: NewSemVer(ver),
	}
}

type VerItems []*VerItem

func (items VerItems) Sort() {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Remote != "" {
			return items[i].Name > items[j].Name
		}

		if items[i].SemVer == nil {
			items[i].SemVer = NewSemVer(items[i].Ver)
		}
		if items[j].SemVer == nil {
			items[j].SemVer = NewSemVer(items[j].Ver)
		}
		return items[i].SemVer.Greater(items[j].SemVer)
	})
}

type Versions struct {
	dict map[string]*VerItem
}

func NewVersions() *Versions {
	return &Versions{
		dict: make(map[string]*VerItem),
	}
}

func NewVersionsFromFile(filename string) (*Versions, error) {
	vers := NewVersions()
	if err := vers.ReadFile(filename); err != nil {
		return nil, err
	}
	return vers, nil
}

func (vers *Versions) WriteFile(filename string) error {
	b, err := json.MarshalIndent(vers.dict, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(filename, 0, bytes.NewReader(b))
}

func (vers *Versions) ReadFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	dict := make(map[string]*VerItem)
	if err = json.Unmarshal(b, &dict); err != nil {
		return err
	}
	vers.dict = dict

	return nil
}

func (vers *Versions) Add(name, ver, sum, ext string) bool {
	if IsSemVer(ver) {
		vers.dict[ver] = NewVerItem(name, "", ver, sum, ext)
		return true
	}
	return false
}

func (vers *Versions) AddBranch(name, remote string) {
	vers.dict[name] = NewVerItem(name, remote, "", "", "")
}

func (vers *Versions) GetItem(ver string) *VerItem {
	return vers.dict[ver]
}

var reSemVerWithoutTag = regexp.MustCompile(`^\d+[.]\d+([.]\d+)?$`)

func (vers *Versions) PickItem(ver string) *VerItem {
	if item := vers.GetItem(ver); item != nil {
		return item
	}

	items := VerItems{}
	for _, item := range vers.dict {
		if reSemVerWithoutTag.Match([]byte(item.Ver)) {
			items = append(items, item)
		}
	}
	items.Sort()

	// pick a latest version
	if ver == "latest" {
		return items[0]
	}

	// pick a latest version that matches the prefix
	for _, item := range items {
		if strings.HasPrefix(item.Ver, ver) {
			return item
		}
	}

	return nil
}

func (vers *Versions) GetList() (VerItems, int) {
	vitems := VerItems{}
	maxlen := 0
	for ver, item := range vers.dict {
		vitems = append(vitems, item)
		if len(ver) > maxlen {
			maxlen = len(ver)
		}
	}
	vitems.Sort()
	return vitems, maxlen
}

func ListTargetVersions(cfg *TargetConfig) string {
	var b strings.Builder

	fmt.Fprintf(&b, "list available %q versions:\n", cfg.Name)
	vers, err := NewVersionsFromFile(cfg.VersionFile)
	if err != nil {
		fmt.Fprintf(&b, "failed to read version file %q: %v\n", cfg.VersionFile, err)
	} else if items, maxlen := vers.GetList(); len(items) > 0 {
		format := fmt.Sprintf("%%-%ds", maxlen)
		ncols := 0
		arr := []string{}
		for _, item := range items {
			v := fmt.Sprintf(format, item.Ver)
			ncols += maxlen + 2
			if ncols/80 > 0 {
				fmt.Fprintf(&b, "%s\n", strings.Join(arr, "  "))
				arr = arr[:0]
				ncols = maxlen + 2
			}
			arr = append(arr, v)
		}
		fmt.Fprintf(&b, "%s\n", strings.Join(arr, "  "))
	}

	return b.String()
}

func PickTargetVersionItem(cfg *TargetConfig, ver string) *VerItem {
	print("check %s version %q definition ... ", cfg.Name, ver)
	vers, err := NewVersionsFromFile(cfg.VersionFile)
	if err != nil {
		fatalf("failed to read version file %q: %v", cfg.VersionFile, err)
	}

	item := vers.PickItem(ver)
	if item == nil {
		printf("not found")
		fatalf("%s version %q does not defined in %q\n%s", cfg.Name, ver, cfg.VersionFile, ListTargetVersions(cfg))
	} else if item.SemVer != nil && item.Ext != ".tar.gz" {
		fatalf("unsupported media-type %q", item.Name)
	}
	printf("found %q", item.Ver)

	return item
}

func CmdVers() {
	for _, cfg := range []*TargetConfig{
		LuaCfg, LuaJitCfg, LuaRocksCfg,
	} {
		print(ListTargetVersions(cfg))
		print("\n")
	}
}
