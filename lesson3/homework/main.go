package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

const DefaultOffset uint64 = 0
const DefaultLimit uint64 = 1024 * 1024 * 1024 * 10
const DefaultBlockSize uint64 = 1024
const MinimumBlockSize uint64 = 256
const TransferFileName string = "transfer_file"

const (
	UpperCase  string = "upper_case"
	LowerCase  string = "lower_case"
	TrimSpaces string = "trim_spaces"
)

type ConvArgs []string

func (d *ConvArgs) String() string {
	return fmt.Sprint(*d)
}

func (d *ConvArgs) Set(s string) error {
	for _, cur := range strings.Split(s, ",") {
		*d = append(*d, cur)
	}
	return nil
}

func (d ConvArgs) IsContain(s string) bool {
	for i := 0; i < len(d); i++ {
		if d[i] == s {
			return true
		}
	}
	return false
}

type Options struct {
	From      string
	To        string
	Offset    uint64
	Limit     uint64
	BlockSize uint64
	Conv      ConvArgs
}

type ConvWriter struct {
	writer io.Writer

	opts Options

	textLen, charBlock, iter  uint64
	notSpace, isLastCharBlock bool
}

type BlockReader interface {
	ReadBlock(uint64, uint64, int) ([]byte, []byte, []byte, error)
	GetReadPointer() uint64
	IsEnd() bool
}

type CustomReader struct {
	Rd bufio.Reader

	ptr   uint64
	isEnd bool
}

func (d *CustomReader) GetReadPointer() uint64 {
	return d.ptr
}

func (d *CustomReader) IsEnd() bool {
	return d.isEnd
}

func DeleteBrokenBytesFromEnd(t []byte) ([]byte, []byte) {
	end := []byte{}
	for len(t) > 0 && !utf8.Valid(t) {
		end = append(end, t[len(t)-1])
		t = t[:len(t)-1]
	}
	for i := 0; i < len(end)/2; i++ {
		end[i], end[len(end)-1-i] = end[len(end)-1-i], end[i]
	}
	return t, end
}

func (d *CustomReader) ReadBlock(l, r uint64, size int) ([]byte, []byte, []byte, error) {
	for d.ptr < l {
		_, err := d.Rd.ReadByte()
		d.ptr++
		if err == io.EOF {
			d.isEnd = true
			return []byte{}, []byte{}, []byte{}, nil
		}
		if err != nil {
			return []byte{}, []byte{}, []byte{}, err
		}
	}
	bgn := []byte{}
	text := []byte{}
	for d.ptr < r {
		cur, err := d.Rd.ReadByte()
		d.ptr++
		if err == io.EOF {
			d.isEnd = true
			return bgn, []byte{}, []byte{}, nil
		}
		if err != nil {
			return []byte{}, []byte{}, []byte{}, err
		}

		if utf8.RuneStart(cur) {
			text = append(text, cur)
			break
		} else {
			bgn = append(bgn, cur)
		}
	}
	for d.ptr < r {
		cur, err := d.Rd.ReadByte()
		d.ptr++
		if err == io.EOF {
			d.isEnd = true
			return bgn, text, []byte{}, nil
		}
		text = append(text, cur)
		if err != nil {
			return []byte{}, []byte{}, []byte{}, err
		}

		if len(text) >= size && utf8.Valid(text) {
			return bgn, text, []byte{}, nil
		}
	}
	text, end := DeleteBrokenBytesFromEnd(text)
	return bgn, text, end, nil
}

func (d *ConvWriter) Write(bgn, text, end []byte) error {
	s := bytes.Runes(text)
	if d.isLastCharBlock && d.opts.Conv.IsContain(TrimSpaces) {
		for i := len(s) - 1; i >= 0; i-- {
			if !unicode.IsSpace(s[i]) {
				s = s[:i+1]
				break
			}
		}
	}
	res := []rune{}
	for i := 0; i < len(s); i++ {
		if d.opts.Conv.IsContain(TrimSpaces) && unicode.IsSpace(s[i]) && !d.notSpace {
			continue
		}
		if !unicode.IsSpace(s[i]) {
			d.notSpace = true
		}
		if d.opts.Conv.IsContain(UpperCase) {
			res = append(res, unicode.ToUpper(s[i]))
		} else if d.opts.Conv.IsContain(LowerCase) {
			res = append(res, unicode.ToLower(s[i]))
		} else {
			res = append(res, s[i])
		}
	}

	text = []byte(string(res))
	if _, err := d.writer.Write(bgn); err != nil {
		return err
	}
	if !(d.iter > d.charBlock && d.opts.Conv.IsContain(TrimSpaces)) {
		if _, err := d.writer.Write(text); err != nil {
			return err
		}
	}
	if _, err := d.writer.Write(end); err != nil {
		return err
	}
	return nil
}

func RebuildText(reader BlockReader, convWriter *ConvWriter) error {
	convWriter.iter = 0
	for !reader.IsEnd() && reader.GetReadPointer() < convWriter.textLen {
		bgn, text, end, err := reader.ReadBlock(0, convWriter.textLen, int(convWriter.opts.BlockSize))
		if err != nil {
			return err
		}
		convWriter.isLastCharBlock = convWriter.iter == convWriter.charBlock
		err = convWriter.Write(bgn, text, end)
		if err != nil {
			return err
		}
		convWriter.iter++
	}
	return nil
}

func IsSpaceBlock(text []byte) bool {
	s := bytes.Runes(text)
	for i := 0; i < len(s); i++ {
		if !unicode.IsSpace(s[i]) {
			return false
		}
	}
	return true
}

func CheckErrorsAndWrite(writer io.Writer, text []byte) error {
	_, err := writer.Write(text)
	return err
}

func TransmitSegment(reader BlockReader, convWriter *ConvWriter, l, r uint64) error {
	var iter uint64 = 0
	for !reader.IsEnd() && reader.GetReadPointer() < r {
		bgn, text, end, err := reader.ReadBlock(l, r, int(convWriter.opts.BlockSize))
		if err != nil {
			return nil
		}
		convWriter.textLen += uint64(len(bgn)) + uint64(len(text)) + uint64(len(end))
		if !IsSpaceBlock(text) {
			convWriter.charBlock = iter
		}
		if err := CheckErrorsAndWrite(convWriter.writer, bgn); err != nil {
			return err
		}
		if err := CheckErrorsAndWrite(convWriter.writer, text); err != nil {
			return err
		}
		if err := CheckErrorsAndWrite(convWriter.writer, end); err != nil {
			return err
		}
		iter++
	}
	return nil
}

func ReadFromStdin(convWriter *ConvWriter) error {
	var err error
	convWriter.writer, err = os.Create(TransferFileName)
	if err != nil {
		return err
	}
	reader := CustomReader{*bufio.NewReader(os.Stdin), 0, false}
	l := convWriter.opts.Offset
	r := convWriter.opts.Offset + convWriter.opts.Limit
	return TransmitSegment(&reader, convWriter, l, r)
}

func ReadFromFile(convWriter *ConvWriter) error {
	var err error
	convWriter.writer, err = os.Create(TransferFileName)
	if err != nil {
		return err
	}
	f, err := os.Open(convWriter.opts.From)
	if err != nil {
		return err
	}
	reader := CustomReader{*bufio.NewReader(f), 0, false}
	l := convWriter.opts.Offset
	r := convWriter.opts.Offset + convWriter.opts.Limit

	err = TransmitSegment(&reader, convWriter, l, r)
	e := f.Close()
	if err != nil {
		return err
	}
	return e
}

func BasicCheck(convWriter *ConvWriter) error {
	if convWriter.opts.To != "" {
		_, err := os.Stat(convWriter.opts.To)
		if err == nil {
			return fmt.Errorf("A file with that name already exists")
		}
	}

	if convWriter.opts.Conv.IsContain(LowerCase) && convWriter.opts.Conv.IsContain(UpperCase) {
		return fmt.Errorf("using lower_case and upper_case flags together")
	}
	for i := 0; i < len(convWriter.opts.Conv); i++ {
		if convWriter.opts.Conv[i] != LowerCase && convWriter.opts.Conv[i] != UpperCase &&
			convWriter.opts.Conv[i] != TrimSpaces {
			return fmt.Errorf("unexpected conv's argument")
		}
	}
	return nil
}

func ParseFlags() (*Options, error) {
	var opts Options

	flag.StringVar(&opts.From, "from", "", "file to read. by default - stdin")
	flag.StringVar(&opts.To, "to", "", "file to write. by default - stdout")
	flag.Uint64Var(&opts.Offset, "offset", DefaultOffset, "number of bytes to skip")
	flag.Uint64Var(&opts.Limit, "limit", DefaultLimit, "number of bytes to read")
	flag.Uint64Var(&opts.BlockSize, "block-size", DefaultBlockSize, "block size for reading")
	flag.Var(&opts.Conv, "conv", `config actions with the file. args:
1) upper_case - convert to uppercase
2) lower_case - convert to lowercase
3) trim_spaces - erase space chars from the beginning and end`)

	flag.Parse()

	return &opts, nil
}

func main() {
	opts, err := ParseFlags()
	convWriter := ConvWriter{opts: *opts}
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "can not parse flags:", err)
		os.Exit(1)
	}

	if opts.Offset > DefaultLimit {
		opts.Offset = DefaultOffset
	}
	if opts.BlockSize < MinimumBlockSize {
		opts.BlockSize = MinimumBlockSize
	}
	err = BasicCheck(&convWriter)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, errors.Unwrap(err))
		os.Exit(1)
	}

	if len(opts.From) == 0 {
		err = ReadFromStdin(&convWriter)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Can't transmit from stdin to the transfer file", err)
			os.Exit(1)
		}
	} else {
		err = ReadFromFile(&convWriter)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Can't transmit from input file to the transfer file", err)
			os.Exit(1)
		}
	}

	if convWriter.textLen == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "offset is too large")
		os.Exit(1)
	}

	f, err := os.Open(TransferFileName)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	reader := CustomReader{*bufio.NewReader(f), 0, false}
	if len(opts.To) == 0 {
		convWriter.writer = os.Stdout
	} else {
		convWriter.writer, err = os.Create(opts.To)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "incorrect output file", err)
			os.Exit(1)
		}
	}
	err = RebuildText(&reader, &convWriter)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "can't apply conv's flags", err)
		os.Exit(1)
	}

	err = f.Close()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "can't close file", err)
		os.Exit(1)
	}
}
