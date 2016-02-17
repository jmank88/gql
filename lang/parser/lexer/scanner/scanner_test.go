package scanner

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestBufferedScanner(t *testing.T) {
	var s Scanner = &bufferedScanner{source: bufio.NewReader(strings.NewReader("foo"))}
	// Scan 'f'
	if err := s.Scan(); err != nil {
		t.Errorf("unexpected error scanning 'f': ", err)
	}
	if s.Rune() != 'f' {
		t.Errorf("expected 'f' but got %s", s.Rune())
	}

	// Start tail at 'f'
	s.StartTail()

	// Scan 'o'
	if err := s.Scan(); err != nil {
		t.Errorf("unexpected error scanning 'o': ", err)
	}
	if s.Rune() != 'o' {
		t.Errorf("expected 'o' but got %s", s.Rune())
	}

	// Scan 'o'
	if err := s.Scan(); err != nil {
		t.Errorf("unexpected error scanning 'o': ", err)
	}
	if s.Rune() != 'o' {
		t.Errorf("expected 'o' but got %s", s.Rune())
	}

	// Scan EOF
	if err := s.Scan(); err == nil {
		t.Error("expected EOF error")
	} else if err != io.EOF {
		t.Errorf("expected EOF but got %s", err)
	}

	tail := s.EndTail()
	if tail != "foo" {
		t.Errorf("expected tail 'foo' but got %q", tail)
	}
}

func TestStringScanner(t *testing.T) {
	var s Scanner = &stringScanner{source: "foo"}
	// Scan 'f'
	if err := s.Scan(); err != nil {
		t.Errorf("unexpected error scanning 'f': ", err)
	}
	if s.Rune() != 'f' {
		t.Errorf("expected 'f' but got %s", s.Rune())
	}

	// Start tail at 'f'
	s.StartTail()

	// Scan 'o'
	if err := s.Scan(); err != nil {
		t.Errorf("unexpected error scanning 'o': ", err)
	}
	if s.Rune() != 'o' {
		t.Errorf("expected 'o' but got %s", s.Rune())
	}

	// Scan 'o'
	if err := s.Scan(); err != nil {
		t.Errorf("unexpected error scanning 'o': ", err)
	}
	if s.Rune() != 'o' {
		t.Errorf("expected 'o' but got %s", s.Rune())
	}

	// Scan EOF
	if err := s.Scan(); err == nil {
		t.Error("expected EOF error")
	} else if err != io.EOF {
		t.Errorf("expected EOF but got %s", err)
	}

	tail := s.EndTail()
	if tail != "foo" {
		t.Errorf("expected tail 'foo' but got %q", tail)
	}
}

var (
	scanBenchString100    = scanBenchString(100)
	scanBenchString1000   = scanBenchString(1000)
	scanBenchString10000  = scanBenchString(10000)
	scanBenchString100000 = scanBenchString(100000)
)

func scanBenchString(size int64) string {
	filename := "test_data/testScan" + strconv.FormatInt(size, 10)
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("failed to open test file: %q: %s", filename, err))
	}
	return string(b)
}

func scan(b *testing.B, initScanner func() Scanner) {
	for n := 0; n < b.N; n++ {
		s := initScanner()

		var err error
		for err == nil {
			err = s.Scan()
		}
		if err != nil && err != io.EOF {
			b.Fatal(err)
		}
	}
}

func tailScan(b *testing.B, initScanner func() Scanner) {
	for n := 0; n < b.N; n++ {
		s := initScanner()

		err := s.Scan()
		if err == nil {
			s.StartTail()
			for err == nil {
				err = s.Scan()
			}
		}
		if err != nil && err != io.EOF {
			b.Fatal(err)
		}
		_ = s.EndTail()
	}
}

func strScanner(source string) func() Scanner {
	return func() Scanner {
		return &stringScanner{source: source}
	}
}

func BenchmarkScanString100(b *testing.B)    { scan(b, strScanner(scanBenchString100)) }
func BenchmarkScanString1000(b *testing.B)   { scan(b, strScanner(scanBenchString1000)) }
func BenchmarkScanString10000(b *testing.B)  { scan(b, strScanner(scanBenchString10000)) }
func BenchmarkScanString100000(b *testing.B) { scan(b, strScanner(scanBenchString100000)) }

func readerScanner(source string) func() Scanner {
	return func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(strings.NewReader(source))}
	}
}

func BenchmarkScanReader100(b *testing.B)    { scan(b, readerScanner(scanBenchString100)) }
func BenchmarkScanReader1000(b *testing.B)   { scan(b, readerScanner(scanBenchString1000)) }
func BenchmarkScanReader10000(b *testing.B)  { scan(b, readerScanner(scanBenchString10000)) }
func BenchmarkScanReader100000(b *testing.B) { scan(b, readerScanner(scanBenchString100000)) }

func scanFile(b *testing.B, filename string) {
	f, err := os.Open(filename)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	scan(b, func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(f)}
	})
}

func BenchmarkScanFile100(b *testing.B)    { scanFile(b, "test_data/testScan100") }
func BenchmarkScanFile1000(b *testing.B)   { scanFile(b, "test_data/testScan1000") }
func BenchmarkScanFile10000(b *testing.B)  { scanFile(b, "test_data/testScan10000") }
func BenchmarkScanFile100000(b *testing.B) { scanFile(b, "test_data/testScan100000") }

func BenchmarkTailScanString100(b *testing.B)    { tailScan(b, strScanner(scanBenchString100)) }
func BenchmarkTailScanString1000(b *testing.B)   { tailScan(b, strScanner(scanBenchString1000)) }
func BenchmarkTailScanString10000(b *testing.B)  { tailScan(b, strScanner(scanBenchString10000)) }
func BenchmarkTailScanString100000(b *testing.B) { tailScan(b, strScanner(scanBenchString100000)) }

func BenchmarkTailScanReader100(b *testing.B)    { tailScan(b, readerScanner(scanBenchString100)) }
func BenchmarkTailScanReader1000(b *testing.B)   { tailScan(b, readerScanner(scanBenchString1000)) }
func BenchmarkTailScanReader10000(b *testing.B)  { tailScan(b, readerScanner(scanBenchString10000)) }
func BenchmarkTailScanReader100000(b *testing.B) { tailScan(b, readerScanner(scanBenchString100000)) }

func tailScanFile(b *testing.B, filename string) {
	f, err := os.Open(filename)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	tailScan(b, func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(f)}
	})
}

func BenchmarkTailScanFile100(b *testing.B)    { tailScanFile(b, "test_data/testScan100") }
func BenchmarkTailScanFile1000(b *testing.B)   { tailScanFile(b, "test_data/testScan1000") }
func BenchmarkTailScanFile10000(b *testing.B)  { tailScanFile(b, "test_data/testScan10000") }
func BenchmarkTailScanFile100000(b *testing.B) { tailScanFile(b, "test_data/testScan100000") }
