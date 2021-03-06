package cli

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	giturls "github.com/whilp/git-urls"
	"github.com/windmilleng/wmclient/pkg/analytics"

	"github.com/windmilleng/tilt/internal/hud/webview"
)

const tiltAppName = "tilt"
const disableAnalyticsEnvVar = "TILT_DISABLE_ANALYTICS"
const analyticsURLEnvVar = "TILT_ANALYTICS_URL"

var analyticsService analytics.Analytics

// Testing analytics locally:
// (after `npm install http-echo-server -g`)
// In one window: `PORT=9988 http-echo-server`
// In another: `TILT_ANALYTICS_URL=http://localhost:9988 tilt up`
// Analytics requests will show up in the http-echo-server window.

func initAnalytics(rootCmd *cobra.Command) error {
	var analyticsCmd *cobra.Command
	var err error

	options := []analytics.Option{}
	options = append(options, analytics.WithGlobalTags(globalTags()))
	analyticsURL := os.Getenv(analyticsURLEnvVar)
	if analyticsURL != "" {
		options = append(options, analytics.WithReportURL(analyticsURL))
	}
	if isAnalyticsDisabledFromEnv() {
		options = append(options, analytics.WithEnabled(false))
	}
	analyticsService, analyticsCmd, err = analytics.Init(tiltAppName, options...)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(analyticsCmd)
	if webview.NewAnalyticsOn() {
		return nil
	}

	status, err := analytics.OptStatus()
	if err != nil {
		return err
	}

	if status == analytics.OptDefault {
		_, err := fmt.Fprintf(os.Stderr, "Send anonymized usage data to Windmill [y/n]? ")
		if err != nil {
			return err
		}

		buf := bufio.NewReader(os.Stdin)
		c, _, _ := buf.ReadRune()
		if c == rune(0) || c == '\n' || c == 'y' || c == 'Y' {
			err = analytics.SetOpt(analytics.OptIn)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(os.Stderr, "Thanks! Setting 'tilt analytics opt in'")
			if err != nil {
				return err
			}
		} else {
			err = analytics.SetOpt(analytics.OptOut)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(os.Stderr, "Thanks! Setting 'tilt analytics opt out'")
			if err != nil {
				return err
			}
		}

		_, err = fmt.Fprintln(os.Stderr, "You set can update your privacy preferences later with 'tilt analytics'")
		if err != nil {
			return err
		}
	}

	return nil
}

func globalTags() map[string]string {
	ret := map[string]string{
		"version": provideTiltInfo().AnalyticsVersion(),
		"os":      runtime.GOOS,
	}

	// store a hash of the git remote to help us guess how many users are running it on the same repository
	origin := normalizeGitRemote(gitOrigin("."))
	if origin != "" {
		h := md5.Sum([]byte(origin))
		ret["git.origin"] = base64.StdEncoding.EncodeToString(h[:])
	}

	return ret
}

func gitOrigin(fromDir string) string {
	cmd := exec.Command("git", "-C", fromDir, "remote", "get-url", "origin")
	b, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimRight(string(b), "\n")
}

func normalizeGitRemote(s string) string {
	u, err := giturls.Parse(string(s))
	if err != nil {
		return s
	}

	// treat "http://", "https://", "git://", "ssh://", etc as equiv
	u.Scheme = ""

	u.User = nil

	// github.com/windmilleng/tilt is the same as github.com/windmilleng/tilt/
	if strings.HasSuffix(u.Path, "/") {
		u.Path = u.Path[:len(u.Path)-1]
	}

	// github.com/windmilleng/tilt is the same as github.com/windmilleng/tilt.git
	if strings.HasSuffix(u.Path, ".git") {
		u.Path = u.Path[:len(u.Path)-4]
	}

	return u.String()
}

func isAnalyticsDisabledFromEnv() bool {
	return os.Getenv(disableAnalyticsEnvVar) != ""
}

func provideAnalytics() (analytics.Analytics, error) {
	if analyticsService == nil {
		return nil, fmt.Errorf("internal error: no available analytics service")
	}
	return analyticsService, nil
}
