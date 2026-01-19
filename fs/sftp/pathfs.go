package sftp

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	pkgsftp "github.com/pkg/sftp"
	"github.com/spf13/afero"
)

// RootPathFs wraps an afero.Fs to handle root directory ("/") specially for SFTP
type RootPathFs struct {
	source     afero.Fs
	homeDir    string
	sftpClient *pkgsftp.Client
}

// NewRootPathFs creates a new RootPathFs
func NewRootPathFs(source afero.Fs, homeDir string, sftpClient *pkgsftp.Client) afero.Fs {
	return &RootPathFs{
		source:     source,
		homeDir:    homeDir,
		sftpClient: sftpClient,
	}
}

// translatePath converts FTP paths to SFTP paths relative to home directory
func (r *RootPathFs) translatePath(name string) string {
	var result string
	if name == "/" || name == "" {
		// Root directory: use current directory marker
		if r.homeDir == "/" {
			result = "."
		} else {
			result = r.homeDir
		}
	} else {
		// Remove leading slash and join with home directory
		relativePath := strings.TrimPrefix(name, "/")
		if r.homeDir == "/" {
			// If home is root, use relative paths from current directory
			result = "./" + relativePath
		} else {
			result = r.homeDir + "/" + relativePath
		}
	}
	return result
}

func (r *RootPathFs) Create(name string) (afero.File, error) {
	return r.source.Create(r.translatePath(name))
}

func (r *RootPathFs) Mkdir(name string, perm os.FileMode) error {
	return r.source.Mkdir(r.translatePath(name), perm)
}

func (r *RootPathFs) MkdirAll(path string, perm os.FileMode) error {
	return r.source.MkdirAll(r.translatePath(path), perm)
}

func (r *RootPathFs) Open(name string) (afero.File, error) {
	translatedPath := r.translatePath(name)

	// Try to open with the translated path first
	file, err := r.source.Open(translatedPath)

	// If it failed and we have SFTP client, use direct ReadDir
	// This works around issues with some SFTP servers (like FileZilla Pro)
	if err != nil && r.sftpClient != nil {
		return &sftpDirFile{
			client: r.sftpClient,
			path:   translatedPath,
		}, nil
	}

	return file, err
}

// sftpDirFile wraps SFTP client for directory listing
type sftpDirFile struct {
	client   *pkgsftp.Client
	path     string
	entries  []fs.FileInfo
	position int
}

func (f *sftpDirFile) Close() error {
	return nil
}

func (f *sftpDirFile) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("cannot read from directory")
}

func (f *sftpDirFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, fmt.Errorf("cannot read from directory")
}

func (f *sftpDirFile) Seek(offset int64, whence int) (int64, error) {
	return 0, fmt.Errorf("cannot seek directory")
}

func (f *sftpDirFile) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("cannot write to directory")
}

func (f *sftpDirFile) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, fmt.Errorf("cannot write to directory")
}

func (f *sftpDirFile) Name() string {
	return f.path
}

func (f *sftpDirFile) Readdir(count int) ([]fs.FileInfo, error) {
	if f.entries == nil {
		entries, err := f.client.ReadDir(f.path)
		if err != nil {
			return nil, err
		}
		f.entries = entries
		f.position = 0
	}

	if count <= 0 {
		result := f.entries[f.position:]
		f.position = len(f.entries)
		return result, nil
	}

	end := f.position + count
	if end > len(f.entries) {
		end = len(f.entries)
	}

	result := f.entries[f.position:end]
	f.position = end

	return result, nil
}

func (f *sftpDirFile) Readdirnames(n int) ([]string, error) {
	infos, err := f.Readdir(n)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name()
	}

	return names, nil
}

func (f *sftpDirFile) Stat() (fs.FileInfo, error) {
	return f.client.Stat(f.path)
}

func (f *sftpDirFile) Sync() error {
	return nil
}

func (f *sftpDirFile) Truncate(size int64) error {
	return fmt.Errorf("cannot truncate directory")
}

func (f *sftpDirFile) WriteString(s string) (ret int, err error) {
	return 0, fmt.Errorf("cannot write to directory")
}

func (r *RootPathFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return r.source.OpenFile(r.translatePath(name), flag, perm)
}

func (r *RootPathFs) Remove(name string) error {
	return r.source.Remove(r.translatePath(name))
}

func (r *RootPathFs) RemoveAll(path string) error {
	return r.source.RemoveAll(r.translatePath(path))
}

func (r *RootPathFs) Rename(oldname, newname string) error {
	return r.source.Rename(r.translatePath(oldname), r.translatePath(newname))
}

func (r *RootPathFs) Stat(name string) (os.FileInfo, error) {
	return r.source.Stat(r.translatePath(name))
}

func (r *RootPathFs) Name() string {
	return "RootPathFs"
}

func (r *RootPathFs) Chmod(name string, mode os.FileMode) error {
	return r.source.Chmod(r.translatePath(name), mode)
}

func (r *RootPathFs) Chown(name string, uid, gid int) error {
	return r.source.Chown(r.translatePath(name), uid, gid)
}

func (r *RootPathFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return r.source.Chtimes(r.translatePath(name), atime, mtime)
}
