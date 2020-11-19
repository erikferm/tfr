/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"

	tfr "github.com/emla2805/tfr/protobuf"
	"github.com/emla2805/tfr/utils"
)

var numberRecords int

var rootCmd = &cobra.Command{
	Use:   "tfr",
	Short: "A lightweight commandline TFRecords processor",
	Long: `tfr is tool for processing TFRecords. 

  Example:

  tfr data_tfrecord-00000-of-00001
  {
    "features": {
      "feature": {
        "a": {
          "bytesList": {
            "value": [
              "some text"
            ]
          }
        },
        "label": {
          "int64List": {
            "value": [
              0
            ]
          }
        }
      }
    }
  }

  `,

	RunE: func(cmd *cobra.Command, args []string) error {
		if isInputFromPipe() {
			return parseRecords(os.Stdin)
		} else {
			file, e := os.Open(args[0])
			if e != nil {
				return e
			}
			defer file.Close()
			return parseRecords(file)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type Example struct {
	Features []Feature
}

type Feature struct {
	Key   string
	Value interface{}
}

func init() {
	rootCmd.Flags().IntVarP(&numberRecords, "number", "n", math.MaxInt32, "Number of records to show")
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

func parseRecords(r io.Reader) error {
	reader := utils.NewReader(r)
	count := 0

	for {
		example, e := reader.Next()

		if count >= numberRecords {
			break
		}

		if e == io.EOF {
			break
		} else if e != nil {
			return e
		}

		ex := &Example{
			Features: []Feature{},
		}
		json1, _ := protojson.Marshal(example)
		for k, v := range example.Features.Feature {
			switch t := v.Kind.(type) {
			case *tfr.Feature_BytesList:
				for _, val := range t.BytesList.Value {
					ex.Features = append(ex.Features, Feature{Key: k, Value: val})
					fmt.Printf("%s: %s\n", k, val)
				}
			}
		}
		fmt.Println(ex)
		j, _ := json.Marshal(e)
		fmt.Fprintln(os.Stdout, string(json1))
		fmt.Fprintln(os.Stdout, string(j))

		count++
	}
	return nil
}
