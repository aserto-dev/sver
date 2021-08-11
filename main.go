package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	flagNext       = ""
	flagMajorOnly  = false
	flagMinorOnly  = false
	flagPreRelease = ""

	flagTagsServerURL = ""
	flagTagsUsername  = ""
	flagTagsPassword  = ""
)

var rootCmd = &cobra.Command{
	Use: "calc-version [flags]",
	RunE: func(cmd *cobra.Command, args []string) error {
		version, err := currentVersion()
		if err != nil {
			return err
		}

		if flagPreRelease != "" {
			version = preRelease(version, flagPreRelease)
		}

		if flagNext != "" {
			version, err = next(version, flagNext)
			if err != nil {
				return err
			}
		}

		if flagMinorOnly && flagMajorOnly {
			return errors.New("can't use --minor and --major in the same run")
		}

		if flagMinorOnly {
			major, minor, _, tail, err := parts(version)
			if err != nil {
				return errors.Wrap(err, "failed to get version parts")
			}
			if tail != "" {
				return errors.Errorf("'%s' is a development version - can't use the --minor flag", version)
			}
			version = fmt.Sprintf("%d.%d", major, minor)
		}

		if flagMajorOnly {
			major, _, _, tail, err := parts(version)
			if err != nil {
				return errors.Wrap(err, "failed to get version parts")
			}
			if tail != "" {
				return errors.Errorf("'%s' is a development version - can't use the --major flag", version)
			}
			version = fmt.Sprintf("%d", major)
		}

		fmt.Println(version)

		return nil
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and exit",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("calc-version %s\n", GetInfo().String())
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
		version, err := currentVersion()
		if err != nil {
			return err
		}

		if flagPreRelease != "" {
			version = preRelease(version, flagPreRelease)
		}

		serverURL, err := url.Parse(flagTagsServerURL)
		if err != nil {
			return err
		}

		host := flagTagsServerURL
		if serverURL.Host != "" {
			host = serverURL.Host
		}

		existingTags, err := imageTags(host+"/"+args[0], flagTagsUsername, flagTagsPassword)
		if err != nil {
			return err
		}

		tags, err := calculateTagsForVersion(version, existingTags)
		if err != nil {
			return err
		}

		for _, tag := range tags {
			fmt.Println(tag)
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
