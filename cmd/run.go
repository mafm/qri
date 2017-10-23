// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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
	"time"

	"github.com/qri-io/cafs/memfs"
	"github.com/qri-io/dataset"
	"github.com/qri-io/dataset/dsfs"
	sql "github.com/qri-io/dataset_sql"
	"github.com/qri-io/qri/repo"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a query",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			ErrExit(fmt.Errorf("Please provide a query string to execute"))
		}

		var (
			structure *dataset.Structure
			results   []byte
		)

		r := GetRepo(false)

		store, err := GetIpfsFilestore(false)
		ExitIfErr(err)

		ds := &dataset.Dataset{
			Timestamp:   time.Now().In(time.UTC),
			QuerySyntax: "sql",
			QueryString: args[0],
			// TODO - set query schema
		}

		format, err := dataset.ParseDataFormatString(cmd.Flag("format").Value.String())
		if err != nil {
			ErrExit(fmt.Errorf("invalid data format: %s", cmd.Flag("format").Value.String()))
		}
		structure, results, err = sql.Exec(store, ds, func(o *sql.ExecOpt) {
			o.Format = format
		})
		ExitIfErr(err)

		// TODO - move this into setting on the dataset outparam
		ds.Structure = structure
		ds.Length = len(results)
		ds.Data, err = store.Put(memfs.NewMemfileBytes("data", results), false)
		ExitIfErr(err)

		name, err := cmd.Flags().GetString("name")
		ExitIfErr(err)

		pin := name != ""

		dspath, err := dsfs.SaveDataset(store, ds, pin)
		ExitIfErr(err)

		// err = r.PutDataset(dspath, ds)
		// ExitIfErr(err)

		if name != "" {
			err = r.PutName(name, dspath)
			ExitIfErr(err)
		}

		err = r.LogQuery(&repo.DatasetRef{Name: name, Path: dspath, Dataset: ds})
		ExitIfErr(err)

		// rgraph.AddResult(dspath, dspath)
		// err = SaveQueryResultsGraph(rgraph)
		// ExitIfErr(err)

		// TODO - restore
		// rqgraph, err := r.repo.ResourceQueries()
		// if err != nil {
		// 	return err
		// }

		// for _, key := range ds.Resources {
		// 	rqgraph.AddQuery(key, dspath)
		// }
		// err = r.repo.SaveResourceQueries(rqgraph)
		// if err != nil {
		// 	return err
		// }

		o := cmd.Flag("output").Value.String()
		if o != "" {
			ioutil.WriteFile(o, results, 0666)
			return
		}

		PrintResults(structure, results, format)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	// runCmd.Flags().StringP("save", "s", "", "save the resulting dataset to a given address")
	runCmd.Flags().StringP("output", "o", "", "file to write to")
	runCmd.Flags().StringP("format", "f", "csv", "set output format [csv,json]")
	runCmd.Flags().StringP("name", "n", "", "save output to local repository with given name")
}
