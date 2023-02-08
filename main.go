package main

import (
	"fmt"
	"math"
	"os"
)

const (
	alphabetSize int = 26
	prime        int = 101
)

type DeltaData struct {
	Fingerprint   []byte
	Chunks        []byte
	NewChunks     []byte
	ChangedChunks []byte
	DeletedChunks []byte
}

func main() {
	_ = Signature("test_files/file_original.txt", "fingerprint.txt")
	data, _ := Delta("fingerprint.txt", "test_files/file_add_begining.txt")

	fmt.Println("Fingerprint", data.Fingerprint)
	fmt.Println("Chunks", data.Chunks)
	fmt.Println("NewChunks", data.NewChunks)
	fmt.Println("ChangedChunks", data.ChangedChunks)
	fmt.Println("DeletedChunks", data.DeletedChunks)
}

func Signature(filePath, fingerprintPath string) error {
	file, err := readFromFile(filePath)
	if err != nil {
		return err
	}
	_, chunks := rollingHash(file)
	return os.WriteFile(fingerprintPath, chunks, 0644)
}

func Delta(fingerprintPath, filePath string) (DeltaData, error) {
	fingerprint, err := readFromFile(fingerprintPath)
	if err != nil {
		return DeltaData{}, err
	}
	file, err := readFromFile(filePath)
	if err != nil {
		return DeltaData{}, err
	}

	_, chunks := rollingHash(file)
	new := make([]byte, 0, 1)
	changed := make([]byte, 0, 1)
	deleted := make([]byte, 0, 1)
	fingerprintMap := make(map[int]byte)

	if len(chunks) == len(fingerprint) {
		for i, c := range chunks {
			if c != fingerprint[i] {
				changed = append(changed, c)
			}
		}
	} else if len(chunks) < len(fingerprint) {
		for i, c := range chunks {
			if c != fingerprint[i] {
				changed = append(changed, c)
			}
		}
		deleted = append(deleted, fingerprint[len(chunks):]...)
	} else {
		for i, f := range fingerprint {
			fingerprintMap[i] = f
		}
		for i, c := range chunks {

			for j := 0; j < len(fingerprint); j++ {
				f := fingerprint[j]
				if f == c {
					break
				} else if f != c && i == j {
					new = append(new, c)
					break
				}
			}

		}
		if len(new) == 0 {
			new = append(new, chunks[len(fingerprint):]...)
		}
	}
	return DeltaData{
		Fingerprint:   fingerprint,
		Chunks:        chunks,
		NewChunks:     new,
		ChangedChunks: changed,
		DeletedChunks: deleted,
	}, nil
}

func readFromFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func fingerprintHash(w []byte) int {
	var fingerprint int
	for _, r := range w {
		fingerprint = (alphabetSize * int(r)) % prime
	}
	return fingerprint
}

func updateHash(s []byte, hash int, windowSize int) int {
	for i := 0; i < windowSize; i++ {
		hash = ((alphabetSize)*hash + int(s[i])) % prime
	}
	return hash
}

func rollingHash(b []byte) ([]int, []byte) {
	initChunkSize := 4
	boundary := 3
	return roll(b, initChunkSize, boundary)
}

func roll(b []byte, initChunkSize, boundary int) (boundaries []int, chunks []byte) {
	h := int(math.Pow(float64(alphabetSize), float64(initChunkSize-1))) % prime
	mask := 1 << boundary
	mask--
	divisor := 10
	n := len(b)
	hash := updateHash(b, 0, initChunkSize)
	chunks = make([]byte, 0, 1)
	boundaries = make([]int, 0, 1)

	for i := 0; i < n-initChunkSize; i++ {
		fingerprint := fingerprintHash(b[i : i+initChunkSize])
		if ((fingerprint % divisor) & mask) == 0 {
			boundaries = append(boundaries, i+initChunkSize)
			chunks = append(chunks, uint8(hash))
		}
		s_char1 := int(b[i])
		s_char2 := int(b[i+initChunkSize])
		hash = (alphabetSize*(hash-s_char1*h) + s_char2) % prime
		if hash < 0 {
			hash += prime
		}
	}
	if len(chunks) < 2 {
		initChunkSize--
		boundary--
		return roll(b, initChunkSize, boundary)
	}
	return
}
