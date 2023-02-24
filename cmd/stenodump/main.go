package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/google/stenographer/base"
	"github.com/google/stenographer/blockfile"
	"github.com/google/stenographer/filecache"
	"go.uber.org/zap"
)

func processFile(infilepath string, outpath string) error {
	blk, err := blockfile.NewBlockFile(infilepath, filecache.NewCache(10))

	if err != nil {
		zap.S().Error(err)
		return err
	}

	infile := filepath.Base(infilepath)
	outfile := infile + ".pcap"
	outfilepath := outpath + "/" + outfile
	zap.S().Infof("Dumping stenographer file %s to %s", infilepath, outfilepath)

	f, err := os.OpenFile(outfilepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		zap.S().Error(err)
		return err
	}
	defer f.Close()
	limit := base.Limit{}
	limit.Bytes = blk.Size()
	err = base.PacketsToFile(blk.AllPackets(), f, limit)
	if err != nil {
		zap.S().Error(err)
		return err
	}
	return nil
}

func main() {
	fmt.Printf("Stenodump!\n\n")

	inputPath := flag.String("i", "./", "input path of stenographer files.  Recursively crawled")
	outputPath := flag.String("o", "./", "output folder for pcap")

	flag.Parse()
	start := time.Now()
	loggerz, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(loggerz)

	f, err := os.Open(*inputPath)
	if err != nil {
		zap.S().Info(err)
		return
	}
	defer f.Close()
	fileInfo, err := f.Stat()
	if err != nil {
		zap.S().Info(err)
		return
	}

	if fileInfo.IsDir() {
		filepath.Walk(*inputPath, func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() {
				err := processFile(path, *outputPath)
				if err != nil {
					zap.S().Info(err)
				}
			}
			return nil
		})
	} else {
		err = processFile(*inputPath, *outputPath)
		zap.S().Error(err)
	}

	zap.S().Infof("Done processing; time: %s", time.Since(start))

}
