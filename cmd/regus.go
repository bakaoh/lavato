package cmd

import (
	"fmt"

	"github.com/bakaoh/lavato/pkg/utils"
	"github.com/bakaoh/lavato/services/regus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var regusCmd = &cobra.Command{
	Use:   "regus",
	Short: "Starts Regus service",
	Long:  `Starts Regus service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return regus.NewServer().Run(fmt.Sprintf(":%d", viper.GetInt("regus.http_port")))
	},
}

var regusStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Regus service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return utils.KillProcess("lavato regus")
	},
}

func init() {
	RootCmd.AddCommand(regusCmd)
	regusCmd.AddCommand(regusStopCmd)
}
