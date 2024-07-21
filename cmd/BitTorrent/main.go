package main

import (
	"log"
	"os"

	"github.com/MarioMottl/BitTorrent/pkg/torrentfile"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <input.torrent> <output.file>", os.Args[0])
	}

	inPath := os.Args[1]
	outPath := os.Args[2]

	tf, err := torrentfile.OpenTorrentFile(inPath)
	if err != nil {
		log.Fatal(err)
	}

	err = tf.DownloadToFile(outPath)
	if err != nil {
		log.Fatal(err)
	}
}
