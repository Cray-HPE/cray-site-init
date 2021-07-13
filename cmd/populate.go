package cmd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
)

// populateCmd moves csi files and node images onto the usb device
var populateCmd = &cobra.Command{
	Use:   "populate",
	Short: "Populates the LiveCD with configs",
	Long:  `Populates the LiveCD with network interface configs after running the format command.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("populate called")
	},
}

// copyAllFiles copies ONLY files from one spot to another
// this is meant as a quick way to dump a bunch of files to the prep dir
func copyAllFiles(src string, dest string) {
	srcFiles, srcErr := ioutil.ReadDir(src)

	if srcErr != nil {
		log.Fatal(srcErr)
	}

	for _, f := range srcFiles {

		s, serr := filepath.Abs(filepath.Join(src, f.Name()))

		if serr != nil {
			log.Fatal(serr)
		}
		dest, destErr := filepath.Abs(filepath.Join(dest, f.Name()))

		if destErr != nil {
			log.Fatal(destErr)
		}

		fi, ferr := os.Stat(s)
		if ferr != nil {
			fmt.Println(ferr)
			return
		}

		switch mode := fi.Mode(); {
		case mode.IsDir():
			// do nothing with dirs
			// fmt.Printf("%s> is a directory...Skipping\n", PadRight(f.Name(), "-", 30))
		case mode.IsRegular():
			// do file stuff
			fmt.Printf("%s> %s", PadRight(f.Name(), "-", 30), dest)
			// copy the file into place
			err := copyFile(s, dest)
			if err != nil {
				fmt.Printf("...Failed %q\n", err)
			} else {
				fmt.Printf("...OK\n")
			}
		}
	}
}

// PadRight adds nice formatting for strings
func PadRight(str, pad string, lenght int) string {
	for {
		str += pad
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}

// copyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

// CopyDirectory copies a directory
func CopyDirectory(scrDir, dest string) error {
	entries, err := ioutil.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}

		isSymlink := entry.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, entry.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

// Copy is the main copy function
func Copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

// Exists checks if a file exists
func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

// CreateIfNotExists creates it if it doesn't
func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

// CopySymLink checks and copy symlinks
func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}

// WalkMatch finds files on a pattern match
func WalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

// CopyArtifactsToPart copies files needed to the PITDATA partition
func CopyArtifactsToPart(src string, dest string, regex string) {
	artifacts, _ := WalkMatch(src, regex)
	if artifacts == nil {
		log.Fatalf("Error: unable to find %s in %s\n", regex, src)
	}
	for _, k := range artifacts {
		fname := filepath.Base(k)
		fmt.Printf("%s> %s", PadRight(fname, "-", 50), dest)
		copyErr := copyFile(k, filepath.Join(dest, fname))
		if copyErr != nil {
			fmt.Printf("...Failed %q\n", copyErr)
		} else {
			fmt.Printf("...OK\n")
		}
	}
}

func init() {
	pitCmd.AddCommand(populateCmd)
	populateCmd.DisableAutoGenTag = true
}
