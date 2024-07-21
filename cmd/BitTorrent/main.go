package main

import (
	"log"
	"os"

	"github.com/MarioMottl/BitTorrent/pkg/torrentfile"
)

func main() {
	inPath := os.Args[1]
	outPath := os.Args[2]

	file, err := os.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}

	tf, err := torrentfile.OpenTorrentFile(file)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}

	err = tf.DownloadToFile(outPath)
	if err != nil {
		log.Fatal(err)
	}
}
