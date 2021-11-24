package cmd

import (
	"fmt"
	"github.com/Masterminds/semver"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/jsiebens/hashi-up/pkg/archive"
	"github.com/jsiebens/hashi-up/pkg/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func GetCommand(product string) *cobra.Command {

	var version string
	var arch string
	var destination string
	var extract bool

	var command = &cobra.Command{
		Use:          "get",
		Short:        fmt.Sprintf("Download %s on your local machine", strings.Title(product)),
		Long:         fmt.Sprintf("Download %s on your local machine", strings.Title(product)),
		SilenceUsage: true,
	}

	title := strings.Title(product)

	command.Flags().StringVarP(&version, "version", "v", "", fmt.Sprintf("Version of %s to install", title))
	command.Flags().StringVar(&arch, "arch", runtime.GOARCH, "Target architecture")
	command.Flags().BoolVar(&extract, "extract", true, "Extract the binary from the downloaded archive")
	command.Flags().StringVarP(&destination, "dest", "d", expandPath("~/bin"), "Target directory for the downloaded archive or binary")

	command.RunE = func(command *cobra.Command, args []string) error {

		if len(version) == 0 {
			latest, err := config.GetLatestVersion(product)

			if err != nil {
				return errors.Wrapf(err, "unable to get latest version number, define a version manually with the --version flag")
			}

			version = latest
		}

		semVersion, err := semver.NewVersion(version)
		if err != nil {
			return err
		}

		file, err := downloadFile(config.GetDownloadURL(product, arch, semVersion))

		if err != nil {
			return errors.Wrapf(err, "unable to download %s distribution", title)
		}

		if extract {
			if err := archive.Unzip(file, destination); err != nil {
				return errors.Wrapf(err, "unable to install %s distribution", title)
			}
		} else {
			if err := os.Rename(file, filepath.Join(destination, filepath.Base(file))); err != nil {
				return errors.Wrapf(err, "unable to install %s distribution", title)
			}
		}

		return nil
	}

	return command
}

func downloadFile(downloadURL string) (string, error) {
	fmt.Printf("Downloading file %s \n", downloadURL)
	res, err := http.DefaultClient.Get(downloadURL)
	if err != nil {
		return "", err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("incorrect status for downloading %s: %d", downloadURL, res.StatusCode)
	}

	_, fileName := path.Split(downloadURL)
	tmp := os.TempDir()
	outFilePath := path.Join(tmp, fileName)
	wrappedReader := withProgressBar(res.Body, int(res.ContentLength))
	out, err := os.Create(outFilePath)
	if err != nil {
		return "", err
	}

	defer out.Close()
	defer wrappedReader.Close()

	if _, err := io.Copy(out, wrappedReader); err != nil {
		return "", err
	}

	return outFilePath, nil
}

func withProgressBar(r io.ReadCloser, length int) io.ReadCloser {
	bar := pb.Simple.New(length).Start()
	return bar.NewProxyReader(r)
}
