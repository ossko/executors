package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const TimeLayout = "2006-01-02 15:04:05"
const KEYCHAIN_PATH = ".colonies"

var Verbose bool
var ColoniesServerHost string
var ColoniesServerPort int
var ColoniesInsecure bool
var ColoniesSkipTLSVerify bool
var ColoniesUseTLS bool
var ColonyID string
var ColonyPrvKey string
var ExecutorName string
var ExecutorID string
var ExecutorType string
var ExecutorPrvKey string
var LogDir string
var FsDir string
var SWName string
var SWType string
var SWVersion string
var HWCPU string
var HWModel string
var HWNodes int
var HWMem string
var HWStorage string
var HWGPUCount int
var HWGPUNodeCount int
var HWGPUName string
var HWGPUMem string
var LocDesc string
var Lat float64
var Long float64

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

var rootCmd = &cobra.Command{
	Use:   "ml_executor",
	Short: "Colonies ML executor",
	Long:  "Colonies ML executor",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}