package action

import (
	"io"
	"os"
	"path/filepath"

	"github.com/deis/helm/helm/log"
)

// Fetch gets a chart from the source repo and copies to the workdir.
//
// - chart is the source
// - lname is the local name for that chart (chart-name); if blank, it is set to the chart.
// - homedir is the home directory for the user
// - ns is the namespace for this package. If blank, it is set to the DefaultNS.
func Fetch(chart, lname, homedir string) {

	if lname == "" {
		lname = chart
	}
	src := filepath.Join(homedir, CacheChartPath, chart)
	dest := filepath.Join(homedir, WorkdirChartPath, lname)

	if fi, err := os.Stat(src); err != nil {
	} else if !fi.IsDir() {
		log.Die("Could not find %s: %s", chart, err)
		log.Die("Malformed chart %s: Chart must be in a directory.", chart)
	}

	if err := os.MkdirAll(dest, 0755); err != nil {
		log.Die("Could not create %q: %s", dest, err)
	}

	log.Info("Fetching %s to %s", src, dest)
	if err := copyDir(src, dest); err != nil {
	}
}

// Copy a directory and its subdirectories.
func copyDir(src, dst string) error {

	walker := func(fname string, fi os.FileInfo, e error) error {
		if e != nil {
			log.Warn("Encounter error walking %q: %s", fname, e)
			return nil
		}

		log.Info("Copying %s", fname)
		rf, err := filepath.Rel(src, fname)
		if err != nil {
			log.Warn("Could not find relative path: %s", err)
			return nil
		}
		df := filepath.Join(dst, rf)

		// Handle directories by creating mirrors.
		if fi.IsDir() {
			if err := os.MkdirAll(df, fi.Mode()); err != nil {
				log.Warn("Could not create %q: %s", df, err)
			}
			return nil
		}

		// Otherwise, copy files.
		in, err := os.Open(fname)
		if err != nil {
			log.Warn("Skipping file %s: %s", fname, err)
			return nil
		}
		out, err := os.Create(df)
		if err != nil {
			in.Close()
			log.Warn("Skipping file copy %s: %s", fname, err)
			return nil
		}
		if _, err = io.Copy(out, in); err != nil {
			log.Warn("Copy from %s to %s failed: %s", fname, df, err)
		}

		if err := out.Close(); err != nil {
			log.Warn("Failed to close %q: %s", df, err)
		}
		if err := in.Close(); err != nil {
			log.Warn("Failed to close reader %q: %s", fname, err)
		}

		return nil
	}
	return filepath.Walk(src, walker)
}
