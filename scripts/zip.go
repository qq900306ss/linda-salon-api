// Command zip packages the Lambda bootstrap binary into function.zip with
// the unix executable bit set (0755). Windows' Compress-Archive does not
// preserve unix file modes, which would make the Lambda runtime unable to
// execute the binary.
package main

import (
	"archive/zip"
	"io"
	"log"
	"os"
)

func main() {
	src := "bootstrap"
	dst := "function.zip"
	if len(os.Args) > 1 {
		src = os.Args[1]
	}
	if len(os.Args) > 2 {
		dst = os.Args[2]
	}

	in, err := os.Open(src)
	if err != nil {
		log.Fatalf("open %s: %v", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		log.Fatalf("create %s: %v", dst, err)
	}
	defer out.Close()

	zw := zip.NewWriter(out)

	header := &zip.FileHeader{
		Name:   "bootstrap",
		Method: zip.Deflate,
	}
	header.SetMode(0o755)

	w, err := zw.CreateHeader(header)
	if err != nil {
		log.Fatalf("create zip entry: %v", err)
	}
	if _, err := io.Copy(w, in); err != nil {
		log.Fatalf("write zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		log.Fatalf("close zip: %v", err)
	}
	log.Printf("created %s from %s", dst, src)
}
