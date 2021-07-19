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
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// NewCmdPush creates a new cobra.Command for the push subcommand.
func NewCmdPushs(options *[]crane.Option) *cobra.Command {
	var imageList, registry string
	var cmd = &cobra.Command{
		Use:   "pushs TARBALL IMAGE",
		Short: "Push list images contents as a tarball to a remote registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) (err error) {
			var path = args[0]
			if imageList == "" {
				return fmt.Errorf("image list should not be null")
			}
			if registry == "" {
				return fmt.Errorf("destination registry should not be null")
			}
			var sourceData []byte
			if sourceData, err = ioutil.ReadFile(imageList); err != nil {
				return
			}
			var source sourceConfig
			if err = yaml.Unmarshal(sourceData, &source); err != nil {
				return
			}
			var srcList []string
			for key, value := range source {
				for k, v := range value {
					for _, tag := range v {
						srcList = append(srcList, fmt.Sprintf("%s/%s:%s", key, k, tag))
					}
				}
			}
			for _, src := range srcList {
				var image v1.Image
				if image, err = crane.Loads(path, src); err != nil {
					return fmt.Errorf("loading %s as tarball: %v", path, err)
				}
				var tag name.Tag
				if tag, err = name.NewTag(src, name.StrictValidation); err != nil {
					return
				}
				fmt.Println(fmt.Sprintf("68: %s/%s:%s", registry, tag.RepositoryStr(), tag.TagStr()))
				if err = crane.Push(image, fmt.Sprintf("%s/%s:%s", registry, tag.RepositoryStr(), tag.TagStr()), *options...); err != nil {
					return
				}
			}
			return
		},
	}
	cmd.Flags().StringVarP(&imageList, "image-list", "l", "", "image list to pull")
	cmd.Flags().StringVarP(&registry, "registry", "r", "", "destination registry")
	return cmd
}
