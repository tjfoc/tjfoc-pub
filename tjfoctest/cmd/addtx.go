// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"github.com/spf13/cobra"
	"github.com/tjfoc/tjfoc/tjfoctest/process"
)

// addtxCmd represents the addtx command
var addtxCmd = &cobra.Command{
	Use:   "addtx",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		addtx()
	},
}

func init() {
	rootCmd.AddCommand(addtxCmd)
	addtxCmd.Flags().IntVarP(&process.Times, "gorutine", "g", 1, "并发次数")
	addtxCmd.Flags().StringVarP(&process.Address, "address", "a", process.DefaultAddress, "节点地址")

}

func addtx() {
	process.AddNewTranscation()
}
