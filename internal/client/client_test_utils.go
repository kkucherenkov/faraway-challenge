package client

type mockConnection struct {
	readFunc  func([]byte) (int, error)
	writeFunc func([]byte) (int, error)
}

func (m mockConnection) Read(p []byte) (n int, err error) {
	return m.readFunc(p)
}

func (m mockConnection) Write(p []byte) (n int, err error) {
	return m.writeFunc(p)
}

// fillTestReadBytes - helper to easier mock Reader
func fillTestReadBytes(str string, p []byte) int {
	dataBytes := []byte(str)
	counter := 0
	for i := range dataBytes {
		p[i] = dataBytes[i]
		counter++
		if counter >= len(p) {
			break
		}
	}
	return counter
}
