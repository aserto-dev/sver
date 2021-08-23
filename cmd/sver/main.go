package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/aserto-dev/sver/pkg/sver"
	"github.com/aserto-dev/sver/pkg/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	flagNext        = ""
	flagMajorOnly   = false
	flagMinorOnly   = false
	flagPreRelease  = ""
	flagForce       = false
	flagReleaseOnly = false
	flagPrefix      = false

	flagTagsServerURL = ""
	flagTagsUsername  = ""
	flagTagsPassword  = ""
)

var rootCmd = &cobra.Command{
	Use: "sver [flags]",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagPreRelease != "" && flagReleaseOnly {
			return errors.New("Asked for a pre-release version, but the --release flag is on.")
		}

		version, err := sver.CurrentVersion(flagReleaseOnly, flagForce)
		if err != nil {
			return err
		}

		if flagPreRelease != "" {
			version = sver.PreRelease(version, flagPreRelease)
		}

		if flagNext != "" {
			version, err = sver.Next(version, flagNext)
			if err != nil {
				return err
			}
		}

		if flagMinorOnly && flagMajorOnly {
			return errors.New("can't use --minor and --major in the same run")
		}

		if flagMinorOnly {
			major, minor, _, tail, err := sver.Parts(version)
			if err != nil {
				return errors.Wrap(err, "failed to get version parts")
			}
			if tail != "" {
				return errors.Errorf("'%s' is a development version - can't use the --minor flag", version)
			}
			version = fmt.Sprintf("%d.%d", major, minor)
		}

		if flagMajorOnly {
			major, _, _, tail, err := sver.Parts(version)
			if err != nil {
				return errors.Wrap(err, "failed to get version parts")
			}
			if tail != "" {
				return errors.Errorf("'%s' is a development version - can't use the --major flag", version)
			}
			version = fmt.Sprintf("%d", major)
		}

		if flagPrefix {
			fmt.Println("v" + version)
		} else {
			fmt.Println(version)
		}

		return nil
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and exit",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("sver %s\n", version.GetInfo().String())
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

var tagsCmd = &cobra.Command{
	Use:   "tags <flags> [image]",
	Short: "Prints the tags that should be pushed to a docker registry",
	Long: `Connects to a docker registry and lists all tags for an image.
Depending on whether the current version is a development version and if
it's the latest one, it returns the appropriate tags to be pushed.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagPreRelease != "" && flagReleaseOnly {
			return errors.New("Asked for a pre-release version, but the --release flag is on.")
		}

		version, err := sver.CurrentVersion(flagReleaseOnly, flagForce)
		if err != nil {
			return err
		}

		if flagPreRelease != "" {
			version = sver.PreRelease(version, flagPreRelease)
		}

		serverURL, err := url.Parse(flagTagsServerURL)
		if err != nil {
			return err
		}

		host := flagTagsServerURL
		if serverURL.Host != "" {
			host = serverURL.Host
		}

		existingTags, err := sver.ImageTags(host+"/"+args[0], flagTagsUsername, flagTagsPassword)
		if err != nil {
			return err
		}

		tags, err := sver.CalculateTagsForVersion(version, existingTags)
		if err != nil {
			return err
		}

		for _, tag := range tags {
			if flagPrefix {
				fmt.Println("v" + tag)
			} else {
				fmt.Println(tag)
			}
		}

		return nil
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

func main() {
	rootCmd.Flags().StringVarP(&flagNext, "next", "n", "", "Prints the next version. Possible values are 'major', 'minor' or 'patch'.")
	rootCmd.Flags().StringVarP(&flagPreRelease, "pre-release", "", os.ExpandEnv("${PRE_RELEASE}"), `Adds a pre release identifier to the version. (env "PRE_RELEASE")`)
	rootCmd.Flags().BoolVarP(&flagMajorOnly, "major-only", "m", false, "Only prints the major version. Fails if version is a development version.")
	rootCmd.Flags().BoolVarP(&flagMinorOnly, "minor-only", "r", false, "Only prints the major and minor versions. Fails if version is a development version.")
	rootCmd.Flags().BoolVarP(&flagReleaseOnly, "release", "", false, "Fail if this is a dev, pre-release or dirty version.")
	rootCmd.Flags().BoolVarP(&flagForce, "force", "f", false, "Ignore a dirty repository.")
	rootCmd.Flags().BoolVarP(&flagPrefix, "prefix", "p", false, "Add the 'v' prefix to the output version.")

	tagsCmd.Flags().StringVarP(&flagTagsServerURL, "server", "s", "https://registry-1.docker.io/", "Registry server to connect to.")
	tagsCmd.Flags().StringVarP(&flagTagsUsername, "user", "u", "", "Username for the registry.")
	tagsCmd.Flags().StringVarP(&flagTagsPassword, "password", "p", "", "Password for the registry.")
	tagsCmd.Flags().StringVarP(&flagPreRelease, "pre-release", "", os.ExpandEnv("${PRE_RELEASE}"), `Adds a pre release identifier to the version. (env "PRE_RELEASE")`)

	rootCmd.AddCommand(
		versionCmd,
		tagsCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
