package lexer

import (
	"bufio"
	"bytes"
	"io"
	"os"
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
	scanBenchString100   = scanBenchString(100)
	scanBenchString1000  = scanBenchString(1000)
	scanBenchString10000 = scanBenchString(10000)
)

//TODO randomize this?
func scanBenchString(n int) string {
	b := &bytes.Buffer{}
	for i := 0; i < n; i++ {
		b.WriteRune('A')
	}
	return b.String()
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

func BenchmarkStringScan100(b *testing.B)   { scan(b, strScanner(scanBenchString100)) }
func BenchmarkStringScan1000(b *testing.B)  { scan(b, strScanner(scanBenchString1000)) }
func BenchmarkStringScan10000(b *testing.B) { scan(b, strScanner(scanBenchString10000)) }

func readerScanner(source string) func() Scanner {
	return func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(strings.NewReader(source))}
	}
}

func BenchmarkReaderScan100(b *testing.B)   { scan(b, readerScanner(scanBenchString100)) }
func BenchmarkReaderScan1000(b *testing.B)  { scan(b, readerScanner(scanBenchString1000)) }
func BenchmarkReaderScan10000(b *testing.B) { scan(b, readerScanner(scanBenchString10000)) }

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

func BenchmarkFileReaderScan100(b *testing.B)   { scanFile(b, "test_data/testScan100") }
func BenchmarkFileReaderScan1000(b *testing.B)  { scanFile(b, "test_data/testScan1000") }
func BenchmarkFileReaderScan10000(b *testing.B) { scanFile(b, "test_data/testScan10000") }

func BenchmarkStringTailScan100(b *testing.B)   { tailScan(b, strScanner(scanBenchString100)) }
func BenchmarkStringTailScan1000(b *testing.B)  { tailScan(b, strScanner(scanBenchString1000)) }
func BenchmarkStringTailScan10000(b *testing.B) { tailScan(b, strScanner(scanBenchString10000)) }

func BenchmarkReaderTailScan100(b *testing.B)   { tailScan(b, readerScanner(scanBenchString100)) }
func BenchmarkReaderTailScan1000(b *testing.B)  { tailScan(b, readerScanner(scanBenchString1000)) }
func BenchmarkReaderTailScan10000(b *testing.B) { tailScan(b, readerScanner(scanBenchString10000)) }

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

func BenchmarkFileReaderTailScan100(b *testing.B)   { tailScanFile(b, "test_data/testScan100") }
func BenchmarkFileReaderTailScan1000(b *testing.B)  { tailScanFile(b, "test_data/testScan1000") }
func BenchmarkFileReaderTailScan10000(b *testing.B) { tailScanFile(b, "test_data/testScan10000") }
