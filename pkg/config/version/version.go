package version

import (
	"strings"
)

var (
	// set during build with '-X 0xacab.org/leap/bitmask-vpn/pkg/config/version.appVersion'
	appVersion = "unknown"

    // set when `git archive` is used, the "$Fromat:%(describe)" will be replaced
    // by the o/p of `git describe` https://git-scm.com/docs/gitattributes#_export_subst
	gitArchiveVersion = "$Format:%(describe)$"
)

func Version() string {
	switch {
	case !strings.HasPrefix(gitArchiveVersion, "$Format:"):
		return gitArchiveVersion
	case appVersion != "":
		return appVersion
	default:
		// should not reach here
		return ""
	}
}
