package edit

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Change struct {
	Bytes []byte
	Prior string
}

func Edit(path string, change func([]byte) (Change, error)) (Change, error) {
	if err := ValidatePath(path); err != nil {
		return Change{}, err
	}
	in, err := os.ReadFile(path)
	if err != nil {
		return Change{}, err
	}
	out, err := change(in)
	if err != nil {
		return Change{}, err
	}
	if string(in) == string(out.Bytes) {
		return out, nil
	}
	if err := write(path, out.Bytes, mode(path)); err != nil {
		return Change{}, err
	}
	return out, nil
}

func Write(path string, bytes []byte) error {
	if err := ValidatePath(path); err != nil {
		return err
	}
	return write(path, bytes, mode(path))
}

// ValidatePath rejects non-clean paths and symlinked targets or immediate parents.
func ValidatePath(path string) error {
	if path == "" || filepath.Clean(path) != path {
		return errors.New("refusing unclean path")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	for _, candidate := range []string{path, filepath.Dir(path)} {
		if err := rejectSymlink(candidate); err != nil {
			return err
		}
	}
	return nil
}

// ValidatePathUnder rejects paths outside boundary and symlinks from boundary through target.
func ValidatePathUnder(boundary, target string) error {
	if boundary == "" || filepath.Clean(boundary) != boundary || target == "" || filepath.Clean(target) != target {
		return errors.New("refusing unclean path")
	}
	boundary, err := filepath.Abs(boundary)
	if err != nil {
		return err
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(boundary, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return errors.New("refusing path outside boundary")
	}
	current := boundary
	if err := rejectSymlink(current); err != nil {
		return err
	}
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		if part == "." {
			continue
		}
		current = filepath.Join(current, part)
		if err := rejectSymlink(current); err != nil {
			return err
		}
	}
	return nil
}

func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("refusing symlink path")
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func mode(path string) os.FileMode {
	if info, err := os.Stat(path); err == nil {
		return info.Mode().Perm()
	}
	return 0600
}

func write(path string, bytes []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.New("refusing symlink directory")
	}
	tmp, err := os.CreateTemp(dir, ".hather-")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(bytes); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
