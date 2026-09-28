package edit

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePathAllowsSystemPrefixSymlinks(t *testing.T) {
	if err := ValidatePath(filepath.Join(t.TempDir(), "config")); err != nil {
		t.Fatalf("ValidatePath rejected an otherwise-safe temp path: %v", err)
	}
}

func TestValidatePathUnderRejectsEscapesAndSymlinks(t *testing.T) {
	boundary := tempDir(t)
	inside := filepath.Join(boundary, "config")
	outside := filepath.Join(tempDir(t), "config")
	if err := ValidatePathUnder(boundary, inside); err != nil {
		t.Fatalf("inside path rejected: %v", err)
	}
	for _, path := range []string{
		outside,
		boundary + string(os.PathSeparator) + "nested" + string(os.PathSeparator) + ".." + string(os.PathSeparator) + "config",
	} {
		if err := ValidatePathUnder(boundary, path); err == nil {
			t.Fatalf("unsafe path accepted: %q", path)
		}
	}
	linkedBoundary := filepath.Join(tempDir(t), "boundary")
	if err := os.Symlink(boundary, linkedBoundary); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePathUnder(linkedBoundary, filepath.Join(linkedBoundary, "config")); err == nil {
		t.Fatal("symlinked boundary accepted")
	}
	link := filepath.Join(boundary, "link")
	if err := os.Symlink(tempDir(t), link); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePathUnder(boundary, filepath.Join(link, "config")); err == nil {
		t.Fatal("symlinked descendant accepted")
	}
}

func TestEditWritesAtomicallyAndPreservesModeAndPriorValue(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("theme = old\n"), 0640); err != nil {
		t.Fatal(err)
	}
	got, err := Edit(path, func(in []byte) (Change, error) {
		return Change{Bytes: []byte("theme = hather\n"), Prior: string(in)}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != "theme = hather\n" || got.Prior != "theme = old\n" || info.Mode().Perm() != 0640 {
		t.Fatalf("bytes=%q prior=%q mode=%#o", bytes, got.Prior, info.Mode().Perm())
	}
	if matches, _ := filepath.Glob(filepath.Join(dir, ".hather-*")); len(matches) != 0 {
		t.Fatalf("temporary files remain: %v", matches)
	}
}

func tempDir(t *testing.T) string {
	t.Helper()
	path, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestEditAndWriteRejectEscapesAndSymlinkedAncestors(t *testing.T) {
	dir := tempDir(t)
	external := tempDir(t)
	original := []byte("theme = catppuccin\n")
	path := filepath.Join(external, "config")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(external, link); err != nil {
		t.Fatal(err)
	}
	for _, rejected := range []string{
		filepath.Join(link, "config"),
		dir + string(os.PathSeparator) + "nested" + string(os.PathSeparator) + ".." + string(os.PathSeparator) + "link" + string(os.PathSeparator) + "config",
	} {
		called := false
		if _, err := Edit(rejected, func([]byte) (Change, error) {
			called = true
			return Change{Bytes: []byte("changed")}, nil
		}); err == nil || called {
			t.Fatalf("Edit accepted unsafe path %q: err=%v callback=%t", rejected, err, called)
		}
		if err := Write(rejected, []byte("changed")); err == nil {
			t.Fatalf("Write accepted unsafe path %q", rejected)
		}
		got, err := os.ReadFile(path)
		if err != nil || string(got) != string(original) {
			t.Fatalf("unsafe path changed external target: bytes=%q err=%v", got, err)
		}
	}
}

func TestEditRefusesSymlinksAndLeavesRejectedInputByteIdentical(t *testing.T) {
	dir := tempDir(t)
	path := filepath.Join(dir, "config")
	original := []byte("theme = catppuccin\n")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Edit(path, func([]byte) (Change, error) { return Change{}, errors.New("ambiguous theme") }); err == nil {
		t.Fatal("ambiguous input was accepted")
	}
	bytes, _ := os.ReadFile(path)
	if string(bytes) != string(original) {
		t.Fatalf("rejected bytes changed: %q", bytes)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if _, err := Edit(link, func([]byte) (Change, error) { return Change{Bytes: []byte("changed")}, nil }); err == nil {
		t.Fatal("symlink was followed")
	}
	bytes, _ = os.ReadFile(path)
	if string(bytes) != string(original) {
		t.Fatalf("symlink changed target: %q", bytes)
	}

	realDir := filepath.Join(dir, "real")
	if err := os.Mkdir(realDir, 0700); err != nil {
		t.Fatal(err)
	}
	realPath := filepath.Join(realDir, "nested")
	if err := os.WriteFile(realPath, original, 0600); err != nil {
		t.Fatal(err)
	}
	linkedDir := filepath.Join(dir, "linked-dir")
	if err := os.Symlink(realDir, linkedDir); err != nil {
		t.Fatal(err)
	}
	called := false
	if _, err := Edit(filepath.Join(linkedDir, "nested"), func([]byte) (Change, error) {
		called = true
		return Change{Bytes: []byte("changed")}, nil
	}); err == nil || called {
		t.Fatalf("symlinked parent was read: err=%v callback=%t", err, called)
	}
}
