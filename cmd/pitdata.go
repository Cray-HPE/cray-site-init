package cmd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// pitdataCmd represents the pitdata command
var pitdataCmd = &cobra.Command{
	Use:   "pitdata SRC DEST",
	Short: "Populates the PITDATA partition with necessary config files",
	Long: `Populates the PITDATA partition with necessary config files.

	SRC can be a path to a folder with squashfs image (-k and -c flags).
	SRC can be a path to a folder of csi-generated files (-b flag)
	SRC can be a path to any folder where you only want files copied over (-p flag)
	DEST should be a path to where KIS components will be copied to`,
	// Arg is the path to the csi generated files
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("basecamp") {
			// Copies data.json to the configs folder
			copyAllFiles(filepath.Join(args[0], "basecamp/"), filepath.Join(args[1]))
		}

		if viper.GetBool("kernel") {
			CopyArtifactsToPart(args[0], args[1], "*.kernel")
		}

		if viper.GetBool("initrd") {
			CopyArtifactsToPart(args[0], args[1], "initrd.*.xz")
		}

		if viper.GetBool("kubernetes") {
			// Find only kubernetes images matching this naming structure
			CopyArtifactsToPart(args[0], args[1], "kubernetes-*.squashfs")
		}

		if viper.GetBool("ceph") {
			// Find only ceph images matching this naming structure
			CopyArtifactsToPart(args[0], args[1], "storage-ceph-*.squashfs")
		}

		if viper.GetBool("prep") {
			os.MkdirAll(filepath.Join(args[1], "prep"), 0755)
			// Copies only files and puts them in the prep dir
			// This is useful for backing up the three config files needed by csi
			// as well as any other files like vars.sh, qnd, or anything else you want
			copyAllFiles(args[1], filepath.Join(args[1], "prep"))
		}
	},
}

func init() {
	populateCmd.AddCommand(pitdataCmd)
	// makedocsCmd.DisableAutoGenTag = true
	viper.SetEnvPrefix("pit") // will be uppercased automatically
	viper.AutomaticEnv()
	pitdataCmd.Flags().BoolP("basecamp", "b", false, "Copy any discovered basecamp config files to the 'configs' directory on the PITDATA partition")
	pitdataCmd.Flags().BoolP("kernel", "k", false, "Copy any discovered kernels from SRC to DEST")
	pitdataCmd.Flags().BoolP("initrd", "i", false, "Copy any discovered initrds from SRC to DEST")
	pitdataCmd.Flags().BoolP("prep", "p", false, "Copy only files from a directory from SRC to DEST")
	pitdataCmd.Flags().BoolP("ceph", "C", false, "Copy any discovered ceph squashfs images from SRC to DEST")
	pitdataCmd.Flags().BoolP("kubernetes", "K", false, "Copy any discovered k8s squashfs images from SRC to DEST")
}
