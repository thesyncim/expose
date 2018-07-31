package cmd

import (
	"github.com/spf13/cobra"
	"github.com/thesyncim/expose/codegen"
	"log"
)

func init() {
	log.SetFlags(log.Lshortfile)
	rootCmd.AddCommand(exposeCmd)
	exposeCmd.Flags().StringVarP(&config.Iface, "iface", "i", "", "interface to expose  (required)")
	log.Println(exposeCmd.MarkFlagRequired("iface"))
	exposeCmd.Flags().StringVarP(&config.Outdir, "outdir", "o", config.Iface,
		"output directory for generated client , server and protocol\n"+
			"if empty iface value will be used.",
	)
	exposeCmd.Flags().StringVarP(&config.PkgPath, "package", "p", "",
		"interface package path //example.dsd", //todo
	)
	exposeCmd.MarkFlagRequired("package")

	exposeCmd.Flags().StringVarP(&config.ServiceName, "service-name", "s", config.Iface,
		"service name (default: iterface Name", //todo
	)
	exposeCmd.MarkFlagRequired("service-name")

}

var exposeCmd = &cobra.Command{
	Use:   "gen",
	Short: "Print the version number of Expose",
	Args:  cobra.OnlyValidArgs,
	Long:  `Print the version number of Expose`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(log.Lshortfile)

		log.Println(config)
		log.Println(Generate(config))
	},
}

func Generate(config *codegen.Config) error {
	generators := []codegen.Generator{
		codegen.GenerateProtocolWireTypes,
		codegen.GenerateProtobuf,
		codegen.GenerateClientWrapper,
		codegen.GenerateServerWrapper,
	}
	err := codegen.Generate(config, generators...)
	if err != nil {
		return err
	}
	return nil
}
