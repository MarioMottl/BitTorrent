package p2p

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"log"
	"runtime"

	"github.com/MarioMottl/BitTorrent/pkg/client"
	"github.com/MarioMottl/BitTorrent/pkg/message"
	"github.com/MarioMottl/BitTorrent/pkg/peers"
)

const MaxBlockSize = 1 << 14 // 16KB
const MaxBacklog = 5         // Max number of unfulfilled requests to keep in memory

type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read()
	if err != nil {
		return err
	}
	if msg == nil {
		return nil
	}

	switch msg.ID {
	case message.MsgUnchoke:
		state.client.Choked = false
	case message.MsgChoke:
		state.client.Choked = true
	case message.MsgHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(index)
	case message.MsgPiece:
		n, err := message.ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil

}

func tryDownloadPiece(client *client.Client, pw pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:      pw.index,
		client:     client,
		buf:        make([]byte, pw.length),
		downloaded: 0,
		requested:  0,
		backlog:    0,
	}

	for state.downloaded < pw.length {
		if state.backlog < MaxBacklog && !state.client.Choked {
			blockSize := MaxBlockSize
			if pw.length-state.downloaded < blockSize {
				blockSize = pw.length - state.downloaded
			}
			err := state.client.SendRequest(pw.index, state.downloaded, blockSize)
			if err != nil {
				return nil, err
			}
			state.requested += blockSize
			state.backlog++
		}
		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}
	return state.buf, nil
}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return errors.New("piece has invalid hash")
	}
	return nil
}

func (t *Torrent) startDownloadWorker(peer peers.Peer, queue chan *pieceWork, result chan *pieceResult) {
	client, err := client.New(peer, t.PeerID, t.InfoHash)
	if err != nil {
		return
	}
	defer client.Conn.Close()

	for pw := range queue {
		buf, err := tryDownloadPiece(client, *pw)
		if err != nil {
			continue
		}
		err = checkIntegrity(pw, buf)
		if err != nil {
			continue
		}
		result <- &pieceResult{pw.index, buf}
	}
}

func (t *Torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

func (t *Torrent) Download() ([]byte, error) {
	log.Println("Starting download for", t.Name)
	// Init queues for workers to retrieve work and send results
	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult)
	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		workQueue <- &pieceWork{index, hash, length}
	}

	// Start workers
	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}

	// Collect results into a buffer until full
	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {
		res := <-results
		begin, end := t.calculateBoundsForPiece(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		numWorkers := runtime.NumGoroutine() - 1 // subtract 1 for main thread
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
	}
	close(workQueue)

	return buf, nil
}
