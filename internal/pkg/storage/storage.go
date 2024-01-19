package storage

import (
	"bufio"
	"io"
	"log/slog"
	"math"
	"os"
	"strings"
	"sync"
)

type Storage interface {
	Add(item string)
	Get(index int) string
	Size() int
	Clean()
}

type localStorage struct {
	DB           []string
	storageMutex sync.RWMutex
	logger       *slog.Logger
}

func (s *localStorage) Add(item string) {
	s.storageMutex.Lock()
	s.DB = append(s.DB, item)
	s.storageMutex.Unlock()
}

func (s *localStorage) Size() int {
	return len(s.DB)
}

func (s *localStorage) Get(index int) string {
	s.storageMutex.RLock()
	defer s.storageMutex.RUnlock()
	return s.DB[index]
}

func (s *localStorage) Clean() {
	s.storageMutex.Lock()
	s.DB = make([]string, 0)
	s.storageMutex.Unlock()
}

func CreateStorage(dataFile string) Storage {
	storage := localStorage{DB: make([]string, 0), logger: slog.New(slog.NewJSONHandler(os.Stdout, nil))}
	if len(dataFile) > 0 {
		err := storage.load(dataFile)
		if err != nil {
			return nil
		}
	}
	return &storage
}

func (s *localStorage) load(fileName string) error {
	file, err := os.Open(fileName)

	if err != nil {
		s.logger.Error("not able to read the file", "error", err)
		return nil
	}

	defer func() {
		err = file.Close()
		if err != nil {
			s.logger.Error("not able to close the file", "error", err)
		}
	}() //close after checking err

	fileStat, err := file.Stat()
	if err != nil {
		s.logger.Error("Could not able to get the file stat", "error", err)
		return err
	}

	fileSize := fileStat.Size()
	offset := fileSize - 1
	lastLineSize := 0

	for {
		b := make([]byte, 1)
		n, err := file.ReadAt(b, offset)
		if err != nil {
			s.logger.Error("Error reading file", "error", err)
			break
		}
		char := string(b[0])
		if char == "\n" {
			break
		}
		offset--
		lastLineSize += n
	}

	lastLine := make([]byte, lastLineSize)
	_, err = file.ReadAt(lastLine, offset+1)

	if err != nil {
		s.logger.Error("Not able to read last line with offset", "offset", offset, "lastLine size", lastLineSize, "error", err)
		return err
	}

	err = process(file, s)
	if err != nil {
		s.logger.Error("file process failed", "error", err)
		return err
	}
	return nil
}

func process(f *os.File, s Storage) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	linesPool := sync.Pool{New: func() interface{} {
		lines := make([]byte, 250*1024)
		return lines
	}}

	stringPool := sync.Pool{New: func() interface{} {
		lines := ""
		return lines
	}}

	r := bufio.NewReader(f)

	var wg sync.WaitGroup

	for {
		buf := linesPool.Get().([]byte)

		n, err := r.Read(buf)
		buf = buf[:n]

		if n == 0 {
			if err == io.EOF {
				break
			} else if err != nil {
				logger.Error("read file error", "error", err)
				break
			}
			return err
		}

		nextUntilNewline, err := r.ReadBytes('\n')

		if err != io.EOF {
			buf = append(buf, nextUntilNewline...)
		}

		wg.Add(1)
		go func() {
			processChunk(buf, &linesPool, &stringPool, s)
			wg.Done()
		}()

	}

	wg.Wait()
	return nil
}

func processChunk(chunk []byte, linesPool *sync.Pool, stringPool *sync.Pool, s Storage) {

	var wg2 sync.WaitGroup

	_ = stringPool.Get().(string)
  logs := string(chunk)

	linesPool.Put(chunk)

	linesSlice := strings.Split(logs, "\n")

	stringPool.Put(logs)

	chunkSize := 300
	n := len(linesSlice)
	noOfThread := n / chunkSize

	if n%chunkSize != 0 {
		noOfThread++
	}

	for i := 0; i < (noOfThread); i++ {

		wg2.Add(1)
		go func(s int, e int, st Storage) {
			defer wg2.Done() //to avoid deadlocks
			for i := s; i < e; i++ {
				text := linesSlice[i]
				if len(text) == 0 {
					continue
				}
				st.Add(text)
			}

		}(i*chunkSize, int(math.Min(float64((i+1)*chunkSize), float64(len(linesSlice)))), s)
	}

	wg2.Wait()
	linesSlice = nil
}
