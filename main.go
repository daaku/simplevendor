// Command simplevendor provides for a simple vendoring solution.
package main

import (
	"flag"
	"go/build"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	dryrun  = flag.Bool("n", false, "dry run")
	verbose = flag.Bool("v", false, "verbose mode")

	specials = []string{
		"license",
		"LICENSE",
		"patents",
		"PATENTS",
		"readme",
		"README",
		"readme.md",
		"README.md",
	}
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()
	if *dryrun {
		*verbose = true
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	vendorDir := filepath.Join(pwd, "vendor")

	// current directory may include already vendored packages which we need to
	// ignore.
	vendorImportPrefix, err := build.ImportDir(vendorDir, build.FindOnly)
	if err != nil {
		log.Fatal(err)
	}

	// figure out all the packages under the current directory
	var packages []*build.Package
	err = filepath.Walk(
		pwd,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return nil
			}
			pkg, err := build.ImportDir(path, 0)
			if err != nil {
				if _, ok := err.(*build.NoGoError); ok {
					return nil
				}
				return err
			}
			if strings.HasPrefix(pkg.ImportPath, vendorImportPrefix.ImportPath) {
				return nil
			}
			packages = append(packages, pkg)
			return nil
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// packages within this directory are NOT vendored.
	// the assumption is that you vendor inside your source controlled directory.
	localImports := map[string]bool{}
	for _, pkg := range packages {
		localImports[pkg.ImportPath] = true
	}

	if *verbose {
		for pkg := range localImports {
			log.Printf("local: %s\n", pkg)
		}
	}

	transitivePackages, err := getTransitiveImports(localImports)
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range transitivePackages {
		if localImports[pkg.ImportPath] {
			continue
		}
		if *verbose {
			log.Printf("vendor: %s\n", pkg.ImportPath)
		}
		if err := vendor(vendorDir, pkg); err != nil {
			log.Fatal(err)
		}
	}
}

type transitiveImports struct {
	testEnabled map[string]bool
	packages    map[string]*build.Package
}

func (d *transitiveImports) analyze(path string) error {
	if isStd(path) {
		return nil
	}
	if d.packages[path] != nil {
		return nil
	}
	pkg, err := build.Import(path, "", 0)
	if err != nil {
		return err
	}
	d.packages[path] = pkg
	sets := [][]string{pkg.Imports}
	if d.testEnabled[path] {
		sets = append(sets, pkg.TestImports, pkg.XTestImports)
	}
	for _, set := range sets {
		for _, dep := range set {
			if err := d.analyze(dep); err != nil {
				return err
			}
		}
	}
	return nil
}

func getTransitiveImports(paths map[string]bool) ([]*build.Package, error) {
	d := transitiveImports{
		testEnabled: paths,
		packages:    map[string]*build.Package{},
	}
	for p := range paths {
		if err := d.analyze(p); err != nil {
			return nil, err
		}
	}
	var result []*build.Package
	for _, p := range d.packages {
		result = append(result, p)
	}
	return result, nil
}

func isStd(path string) bool {
	if path == "C" {
		return true
	}
	p, err := build.Import(path, "", build.FindOnly)
	if err != nil {
		return false
	}
	return p.Goroot
}

func vendor(dir string, pkg *build.Package) error {
	targetDir := filepath.Join(dir, pkg.ImportPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	// pkg.SFiles doesn't include ignored stuff, this will
	sFiles, err := filepath.Glob(pkg.Dir + "/*.s")
	if err != nil {
		return err
	}
	for i, v := range sFiles {
		sFiles[i] = filepath.Base(v)
	}

	sets := [][]string{
		pkg.GoFiles,
		pkg.CgoFiles,
		pkg.IgnoredGoFiles,
		pkg.CFiles,
		pkg.CXXFiles,
		pkg.MFiles,
		pkg.HFiles,
		sFiles,
		pkg.SwigFiles,
		pkg.SwigCXXFiles,
		pkg.SysoFiles,
	}
	for _, set := range sets {
		for _, file := range set {
			if strings.HasSuffix(file, "_test.go") {
				continue
			}
			src := filepath.Join(pkg.Dir, file)
			dst := filepath.Join(targetDir, file)
			if *verbose {
				log.Printf("cp %s => %s\n", src, dst)
			}
			if !*dryrun {
				if err := cp(src, dst); err != nil {
					return err
				}
			}
		}
	}
	if !*dryrun {
		for _, file := range specials {
			cp(filepath.Join(pkg.Dir, file), filepath.Join(targetDir, file))
		}
	}
	return nil
}

func cp(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()
	dstF, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	if _, err := io.Copy(dstF, srcF); err != nil {
		dstF.Close()
		return err
	}
	if err := dstF.Close(); err != nil {
		return err
	}
	return os.Chtimes(dst, time.Now(), info.ModTime())
}
