package main // import "hello-targz"

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func compress(source, target string) error {
	tgzfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tgzfile.Close()

	// gw := gzip.NewWriter(tgzfile)
	gw, err := gzip.NewWriterLevel(tgzfile, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, fpath)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = strings.TrimPrefix(fpath, source+"\\")
		}

		header.Name = strings.Replace(header.Name, "\\", "/", -1)

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Format = tar.FormatGNU
		}

		if header.Name != baseDir && header.Name != baseDir+"/" {
			err = tw.WriteHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(fpath)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tw, file)
			return err
		}

		return nil
	})

	return err
}

func decompress(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		path := filepath.Join(dest, header.Name)
		info := header.FileInfo()

		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(path, info.Mode())
			continue
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(path), info.Mode())
			f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, tr)
			if err != nil {
				return err
			}
		default:
			return errors.New("unknown type: " + string(header.Typeflag))
		}
	}

	return nil
}

func jobMain() {
	targetPath := ""
	outputName := ""
	inputName := ""
	destination := ""

	switch len(os.Args) {
	case 2:
		info, err := os.Stat(os.Args[1])
		if err != nil {
			panic(err)
		}

		if info.IsDir() {
			targetPath = strings.TrimSuffix(os.Args[1], "\\")
			targetPath = strings.TrimSuffix(targetPath, "/")
			outputName = targetPath + ".tar.gz"
			err := compress(targetPath, outputName)
			if err != nil {
				panic(err)
			}
		} else {
			inputName = os.Args[1]
			filename := filepath.Base(inputName)
			destination = strings.TrimSuffix(filename, ".tar.gz")
			err := decompress(inputName, destination)
			if err != nil {
				panic(err)
			}
		}
	default:
		log.Println("Usage: hello-zip [directory (=compress) or filename (=decompress)]")
	}
}

func main() {
	jobMain()
}
