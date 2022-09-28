/*
Copyright © 2021 SK Telecom <https://github.com/openinfradev>

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
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/openinfradev/tks-proto/tks_pb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
)

// clusterCreateCmd represents the create command
var clusterCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a TKS Cluster.",
	Long: `Create a TKS Cluster.
  
Example:
tks cluster create <CLUSTERNAME> [--template TEMPLATE_NAME]`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Println("You must specify cluster name.")
			return errors.New("Usage: tks cluster create <CLUSTERNAME>")
		}
		var conn *grpc.ClientConn
		tksClusterLcmUrl = viper.GetString("tksClusterLcmUrl")
		if tksClusterLcmUrl == "" {
			return errors.New("You must specify tksClusterLcmUrl at config file")
		}
		conn, err := grpc.Dial(tksClusterLcmUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Could not connect to LCM server: %s", err)
		}
		defer conn.Close()

		client := pb.NewClusterLcmServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		/* Parse command line arguments */
		ClusterName := args[0]
		ContractId, _ := cmd.Flags().GetString("contract-id")
		CspId, _ := cmd.Flags().GetString("csp-id")
		conf := pb.ClusterRawConf{}
		conf.SshKeyName, _ = cmd.Flags().GetString("ssh-key-name")
		conf.Region, _ = cmd.Flags().GetString("region")
		conf.MachineType, _ = cmd.Flags().GetString("machine-type")

		numOfAz, _ := cmd.Flags().GetInt("num-of-az")
		conf.NumOfAz = int32(numOfAz)

		machineReplicas, _ := cmd.Flags().GetInt("machine-replicas")
		conf.MachineReplicas = int32(machineReplicas)

		templateName, _ := cmd.Flags().GetString("template")

		/* Construct request map */
		data := pb.CreateClusterRequest{
			Name:         ClusterName,
			ContractId:   ContractId,
			CspId:        CspId,
			Conf:         &conf,
			TemplateName: templateName,
		}

		m := protojson.MarshalOptions{
			Indent:        "  ",
			UseProtoNames: true,
		}
		jsonBytes, _ := m.Marshal(&data)
		fmt.Println("Proto Json data: ")
		fmt.Println(string(jsonBytes))

		r, err := client.CreateCluster(ctx, &data)
		fmt.Println("Response:\n", r)
		if err != nil {
			return fmt.Errorf("Error: %s", err)
		} else {
			fmt.Println("Success: The request to create cluster ", args[0], " was accepted.")
		}

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(clusterCreateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clusterCreateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clusterCreateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	clusterCreateCmd.Flags().String("contract-id", "", "Contract ID")
	clusterCreateCmd.Flags().String("csp-id", "", "CSP ID")
	clusterCreateCmd.Flags().String("region", "", "AWS Region")
	clusterCreateCmd.Flags().Int("num-of-az", 3, "Number of availability zones in selected region")
	clusterCreateCmd.Flags().String("ssh-key-name", "", "SSH key name for EC2 instance connection")
	clusterCreateCmd.Flags().String("machine-type", "", "machine type of worker node")
	clusterCreateCmd.Flags().Int("machine-replicas", 3, "machine replicas of worker node")
	clusterCreateCmd.Flags().String("template", "aws-reference", "Template name for the cluster")
}
