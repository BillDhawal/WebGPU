package server

import (
	"path/filepath"

	"fmt"
	"io"
	"os"
	. "wb/app/config"

	"github.com/revel/revel"
)

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
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

func MakeOpenACCCompileCommand(s *WorkerState) string {
	switch OperatingSystem {
	case "Linux":
		//CopyFile(filepath.Join(CToolsDir, SystemId, "libwb.so"), filepath.Join(s.TemporaryDirectory, "libwb.so"))
		cmp := "#!/bin/sh\n" +
			filepath.Join(PGCCCompilerLocation, PGCCCompiler) +
			" -O3 " +
			" -acc " +
			" -ta=nvidia " +
			" -Minfo=accel " +
			" " + filepath.Join(s.TemporaryDirectory, s.ProgramFileName) +
			" -o " + filepath.Join(s.TemporaryDirectory, s.ExecutableFileName) +
			" -I " + filepath.Join(NVCCCompilerLocation, "..", "..", "include") +
			" " + filepath.Join(NVCCCompilerLocation, "..", "..", "lib64", "libcudart_static.a") +
			" -I" + CToolsDir +
			" -I" + filepath.Join(CUDAToolkitDirectory, "include") +
			" " + filepath.Join(CToolsDir, SystemId, "libwb.a") +
			" -L." +
			" -lrt " +
			" -lcudart " +
			" -lstdc++ " +
			" -lm " +
			" -lcuda " +
			" -DWB_USE_CUSTOM_MALLOC " +
			" -DWB_USE_COURSERA " +
			// " -DWB_USE_CUDA " +
			" 2>&1\n"
		revel.TRACE.Println(cmp)
		return cmp
	default:
		return "todo"
	}
}
