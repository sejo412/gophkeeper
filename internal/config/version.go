package config

import "fmt"

const defaultVersion = "N/A"

var (
	BuildVersion = defaultVersion
	BuildCommit  = defaultVersion
	BuildDate    = defaultVersion
)

type Version struct {
	BuildVersion string
	BuildCommit  string
	BuildDate    string
}

func NewVersion() Version {
	return Version{
		BuildVersion: BuildVersion,
		BuildCommit:  BuildCommit,
		BuildDate:    BuildDate,
	}
}

func (v Version) Print() {
	fmt.Printf("Version: %s\n", v.BuildVersion)
	fmt.Printf("Commit: %s\n", v.BuildCommit)
	fmt.Printf("Build date: %s\n", v.BuildDate)
}
