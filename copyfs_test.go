package copyfs_test

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/nabeken/go-copyfs"
	"github.com/stretchr/testify/require"
)

//go:embed testdata *.go
var testdata embed.FS

// TestFS serves to show the contents in testdata.
func TestFS(t *testing.T) {
	fs.WalkDir(testdata, ".", func(path string, d fs.DirEntry, err error) error {
		fi, err := d.Info()
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("TEST_FS: %s (%o)", path, fi.Mode().Perm())
		return nil
	})
}

func TestCopyFS(t *testing.T) {
	require := require.New(t)

	destDir, err := os.MkdirTemp("", "go-copyfs-test")
	require.NoError(err)

	t.Logf("Copying the files into '%s'...", destDir)

	t.Cleanup(cleanupDestDir(t, destDir))

	err = copyfs.CopyFS(destDir, testdata)

	type file struct {
		IsDir bool
		Perm  fs.FileMode
	}

	testCases := map[string]file{
		"copyfs.go": file{
			Perm: 0444,
		},
		"copyfs_test.go": file{
			Perm: 0444,
		},
		"testdata": file{
			IsDir: true,
			Perm:  0555,
		},
		"testdata/a.txt": file{
			Perm: 0444,
		},
		"testdata/b.txt": file{
			Perm: 0444,
		},
		"testdata/c": file{
			IsDir: true,
			Perm:  0555,
		},
		"testdata/c/d.txt": file{
			Perm: 0444,
		},
	}

	destFs := os.DirFS(destDir)
	fs.WalkDir(destFs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Fatal(err)
		}

		if path == "." {
			return nil
		}

		fi, err := d.Info()
		if err != nil {
			t.Fatal(err)
		}

		expected, found := testCases[path]
		require.Truef(found, "%s must exist", path)
		require.Equalf(expected.IsDir, d.IsDir(), "%s must be a directory", path)
		require.Equalf(expected.Perm, fi.Mode().Perm(), "%s must have a correct type but got '%o'", path, fi.Mode().Perm())

		return nil
	})

	require.NoError(err)
}

func cleanupDestDir(t *testing.T, destDir string) func() {
	return func() {
		t.Logf("Removing all the files in '%s'...", destDir)

		// will add write permission to the directories to remove it completely
		destFs := os.DirFS(destDir)

		fs.WalkDir(destFs, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				t.Fatal(err)
			}

			if path == "." {
				return nil
			}

			fp := filepath.Join(destDir, path)

			if d.IsDir() {
				if err := os.Chmod(fp, 0700); err != nil {
					t.Fatalf("Unable to chmod on '%s'", err)
				}
			}

			return nil
		})

		if err := os.RemoveAll(destDir); err != nil {
			t.Logf("Unable to remove all the files: %v", err)
		}
	}
}
