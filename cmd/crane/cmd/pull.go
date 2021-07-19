// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/spf13/cobra"

	"gopkg.in/yaml.v2"
)

// sourceConfig contains all registries and images information read from the source YAML file
type sourceConfig map[string]map[string][]string

// NewCmdPull creates a new cobra.Command for the pull subcommand.
func NewCmdPull(options *[]crane.Option) *cobra.Command {
	var cachePath, format, imageList string
	cmd := &cobra.Command{
		Use:   "pull IMAGE TARBALL",
		Short: "Pull remote images by reference and store their contents in a tarball",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			imageMap := map[string]v1.Image{}
			var srcList []string
			var path = args[len(args)-1]
			if len(args) == 1 {
				if imageList == "" {
					return fmt.Errorf("image list should not be null")
				}
				var source []byte
				var err error
				source, err = ioutil.ReadFile(imageList)
				if err != nil {
					return err
				}
				var images sourceConfig
				if err = yaml.Unmarshal(source, &images); err != nil {
					return err
				}
				for key, value := range images {
					for k, v := range value {
						for _, tag := range v {
							srcList = append(srcList, fmt.Sprintf("%s/%s:%s", key, k, tag))
						}
					}
				}
			} else {
				srcList = args[:len(args)-1]
			}
			for _, src := range srcList {
				fmt.Printf("pulling manifest %s ...\n", src)
				img, err := crane.Pull(src, *options...)
				if err != nil {
					return fmt.Errorf("pulling %s: %v", src, err)
				}
				if cachePath != "" {
					img = cache.Image(img, cache.NewFilesystemCache(cachePath))
				}
				imageMap[src] = img
			}

			switch format {
			case "tarball":
				if err := crane.MultiSave(imageMap, path); err != nil {
					return fmt.Errorf("saving tarball %s: %v", path, err)
				}
			case "legacy":
				if err := crane.MultiSaveLegacy(imageMap, path); err != nil {
					return fmt.Errorf("saving legacy tarball %s: %v", path, err)
				}
			case "oci":
				if err := crane.MultiSaveOCI(imageMap, path); err != nil {
					return fmt.Errorf("saving oci image layout %s: %v", path, err)
				}
			default:
				return fmt.Errorf("unexpected --format: %q (valid values are: tarball, legacy, and oci)", format)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&cachePath, "cache_path", "c", "", "Path to cache image layers")
	cmd.Flags().StringVar(&format, "format", "tarball", fmt.Sprintf("Format in which to save images (%q, %q, or %q)", "tarball", "legacy", "oci"))
	cmd.Flags().StringVarP(&imageList, "image-list", "l", "", "image list to pull")

	return cmd
}
