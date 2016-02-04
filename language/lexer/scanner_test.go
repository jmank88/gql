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
	testString100   = testString(100)
	testString1000  = testString(1000)
	testString10000 = testString(10000)
)

func testString(n int) string {
	b := &bytes.Buffer{}
	for i := 0; i < n; i++ {
		b.WriteRune('A') //TODO randomize this?
	}
	return b.String()
}

func benchmarkScan(b *testing.B, initScanner func() Scanner) {
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

func benchmarkTailScan(b *testing.B, initScanner func() Scanner) {
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

func BenchmarkStringScan100(b *testing.B) {
	benchmarkScan(b, func() Scanner {
		return &stringScanner{source: testString100}
	})
}

func BenchmarkReaderScan100(b *testing.B) {
	benchmarkScan(b, func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(strings.NewReader(testString100))}
	})
}

func BenchmarkFileReaderScan100(b *testing.B) {
	f, err := os.Open("test_data/testScan100")
	if err != nil {
		b.Fatal(err)
	}
	benchmarkScan(b, func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(f)}
	})
}

func BenchmarkStringScan1000(b *testing.B) {
	benchmarkScan(b, func() Scanner {
		return &stringScanner{source: testString1000}
	})
}

func BenchmarkReaderScan1000(b *testing.B) {
	benchmarkScan(b, func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(strings.NewReader(testString1000))}
	})
}

func BenchmarkFileReaderScan1000(b *testing.B) {
	f, err := os.Open("test_data/testScan1000")
	if err != nil {
		b.Fatal(err)
	}
	benchmarkScan(b, func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(f)}
	})
}

func BenchmarkStringScan10000(b *testing.B) {
	benchmarkScan(b, func() Scanner {
		return &stringScanner{source: testString10000}
	})
}

func BenchmarkReaderScan10000(b *testing.B) {
	benchmarkScan(b, func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(strings.NewReader(testString10000))}
	})
}

func BenchmarkFileReaderScan10000(b *testing.B) {
	f, err := os.Open("test_data/testScan10000")
	if err != nil {
		b.Fatal(err)
	}
	benchmarkScan(b, func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(f)}
	})
}

func BenchmarkStringTailScan100(b *testing.B) {
	benchmarkTailScan(b, func() Scanner {
		return &stringScanner{source: testString100}
	})
}

func BenchmarkReaderTailScan100(b *testing.B) {
	benchmarkTailScan(b, func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(strings.NewReader(testString100))}
	})
}

func BenchmarkFileReaderTailScan100(b *testing.B) {
	f, err := os.Open("test_data/testScan100")
	if err != nil {
		b.Fatal(err)
	}
	benchmarkTailScan(b, func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(f)}
	})
}

func BenchmarkStringTailScan1000(b *testing.B) {
	benchmarkTailScan(b, func() Scanner {
		return &stringScanner{source: testString1000}
	})
}

func BenchmarkReaderTailScan1000(b *testing.B) {
	benchmarkTailScan(b, func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(strings.NewReader(testString1000))}
	})
}

func BenchmarkFileReaderTailScan1000(b *testing.B) {
	f, err := os.Open("test_data/testScan1000")
	if err != nil {
		b.Fatal(err)
	}
	benchmarkTailScan(b, func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(f)}
	})
}

func BenchmarkStringTailScan10000(b *testing.B) {
	benchmarkTailScan(b, func() Scanner {
		return &stringScanner{source: testString10000}
	})
}

func BenchmarkReaderTailScan10000(b *testing.B) {
	benchmarkTailScan(b, func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(strings.NewReader(testString10000))}
	})
}

func BenchmarkFileReaderTailScan10000(b *testing.B) {
	f, err := os.Open("test_data/testScan10000")
	if err != nil {
		b.Fatal(err)
	}
	benchmarkTailScan(b, func() Scanner {
		return &bufferedScanner{source: bufio.NewReader(f)}
	})
}
