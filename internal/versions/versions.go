package versions

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mah0x211/lenv/internal/cli"
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
			cli.Fatalf("failed to strconv.Atoi(%#v): %v", s, err)
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

func newSemVer(ver string) *SemVer {
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

func (a *SemVer) greater(b *SemVer) bool {
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
		a := newSemVer(vers[i])
		b := newSemVer(vers[j])
		return a.greater(b)
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

func newVerItem(name, remote, ver, sum, ext string) *VerItem {
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
		SemVer: newSemVer(ver),
	}
}

type VerItems []*VerItem

func (items VerItems) sort() {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Remote != "" {
			return items[i].Name > items[j].Name
		}

		if items[i].SemVer == nil {
			items[i].SemVer = newSemVer(items[i].Ver)
		}
		if items[j].SemVer == nil {
			items[j].SemVer = newSemVer(items[j].Ver)
		}
		return items[i].SemVer.greater(items[j].SemVer)
	})
}

type Versions struct {
	Dict map[string]*VerItem
}

func NewVersions() *Versions {
	return &Versions{
		Dict: make(map[string]*VerItem),
	}
}

func NewVersionsFromFile(filename string) (*Versions, error) {
	vers := NewVersions()
	if err := vers.readFile(filename); err != nil {
		return nil, err
	}
	return vers, nil
}

// WriteFile publishes a complete version list without truncating the old file.
func (vers *Versions) WriteFile(filename string) error {
	b, err := json.MarshalIndent(vers.Dict, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}
	mode := os.FileMode(0644)
	if info, err := os.Stat(filename); err == nil {
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(filename), ".lenv-versions-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if err := f.Chmod(mode); err != nil {
		return err
	}
	if _, err := io.Copy(f, bytes.NewReader(b)); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), filename)
}

func (vers *Versions) readFile(filename string) error {
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
	vers.Dict = dict

	return nil
}

func (vers *Versions) Add(name, ver, sum, ext string) bool {
	if IsSemVer(ver) {
		vers.Dict[ver] = newVerItem(name, "", ver, sum, ext)
		return true
	}
	return false
}

func (vers *Versions) AddBranch(name, remote string) {
	vers.Dict[name] = newVerItem(name, remote, "", "", "")
}

func (vers *Versions) GetItem(ver string) *VerItem {
	return vers.Dict[ver]
}

var reSemVerWithoutTag = regexp.MustCompile(`^\d+[.]\d+([.]\d+)?$`)

func (vers *Versions) PickItem(ver string) *VerItem {
	if item := vers.GetItem(ver); item != nil {
		return item
	}

	items := VerItems{}
	for _, item := range vers.Dict {
		if reSemVerWithoutTag.Match([]byte(item.Ver)) {
			items = append(items, item)
		}
	}
	items.sort()

	// pick a latest version
	if ver == "latest" {
		if len(items) == 0 {
			return nil
		}
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
	for ver, item := range vers.Dict {
		vitems = append(vitems, item)
		if len(ver) > maxlen {
			maxlen = len(ver)
		}
	}
	vitems.sort()
	return vitems, maxlen
}
