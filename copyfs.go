package copyfs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// CopyFS copies contents in the srcFS fs.FS into the local filesystem at destDir. destDir must be a directory.
func CopyFS(destDir string, srcFS fs.FS) error {
	destInfo, destErr := os.Lstat(destDir)
	if destErr != nil {
		return destErr
	}

	if !destInfo.IsDir() {
		return errors.New("the destination must be a directory")
	}

	return copyDir(destDir, srcFS)
}

// copyDir copies contents in src fs.FS. destDir must be a directory.
func copyDir(destDir string, srcFS fs.FS) error {
	sdents, err := fs.ReadDir(srcFS, ".")
	if err != nil {
		fmt.Errorf("reading the source directory: %w", err)
	}

	for _, sdent := range sdents {
		fi, fierr := sdent.Info()
		if fierr != nil {
			return fmt.Errorf("reading the file info: %w", err)
		}

		switch {
		case fi.IsDir():
			subDestDir := filepath.Join(destDir, sdent.Name())

			if err := os.Mkdir(subDestDir, 0700); err != nil {
				return fmt.Errorf("mkdir on %s: %w", subDestDir, err)
			}

			newSrcFS, err := fs.Sub(srcFS, sdent.Name())
			if err != nil {
				return fmt.Errorf("reading the sub directory: %w", err)
			}

			if err := copyDir(subDestDir, newSrcFS); err != nil {
				return err
			}

			// restore the original mode
			if err := os.Chmod(subDestDir, fi.Mode().Perm()); err != nil {
				return fmt.Errorf("chmod on %s: %w", subDestDir, err)
			}

		case fi.Mode()&fs.ModeSymlink != 0:
			// do nothing for the symlink
			// See https://github.com/golang/go/issues/49580

		case fi.Mode().IsRegular():
			srcf, err := srcFS.Open(fi.Name())
			if err != nil {
				return err
			}
			if err := copyFile(destDir, srcf); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(destDir string, src fs.File) error {
	srcStat, srcStatErr := src.Stat()
	if srcStatErr != nil {
		return srcStatErr
	}

	destFn := filepath.Join(destDir, srcStat.Name())

	destf, destErr := os.OpenFile(
		destFn,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		srcStat.Mode(),
	)
	if destErr != nil {
		return fmt.Errorf("opening a file for writing: %w", destErr)
	}

	_, err := io.Copy(destf, src)
	if err != nil {
		return fmt.Errorf("copying %s to %s", srcStat.Name(), destFn)
	}

	return destf.Close()
}
