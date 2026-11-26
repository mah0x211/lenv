package install

type options struct {
	Fetch           bool
	IfMissing       bool
	AllowMissingGPG bool
	Rebuild         bool
	NoRebuild       bool
	Version         string
	Build           []string
}
