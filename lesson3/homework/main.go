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

type BlockReader interface {
	ReadBlock(uint64, uint64, int) ([]byte, []byte, []byte, error)
	GetReadPointer() uint64
	IsEnd() bool
}

type BlockWriter = io.Writer

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

func RelaxChars(opts *Options, text []byte, notSpace bool, isLastCharBlock bool) ([]byte, bool) {
	s := bytes.Runes(text)
	if isLastCharBlock && opts.Conv.IsContain(TrimSpaces) {
		for i := len(s) - 1; i >= 0; i-- {
			if !unicode.IsSpace(s[i]) {
				s = s[:i+1]
				break
			}
		}
	}
	res := []rune{}
	for i := 0; i < len(s); i++ {
		if opts.Conv.IsContain(TrimSpaces) && unicode.IsSpace(s[i]) && !notSpace {
			continue
		}
		if !unicode.IsSpace(s[i]) {
			notSpace = true
		}
		if opts.Conv.IsContain(UpperCase) {
			res = append(res, unicode.ToUpper(s[i]))
		} else if opts.Conv.IsContain(LowerCase) {
			res = append(res, unicode.ToLower(s[i]))
		} else {
			res = append(res, s[i])
		}
	}
	return []byte(string(res)), notSpace
}

func RebuildText(opts *Options, reader BlockReader, writer BlockWriter, textSize, charBlock uint64) error {
	var iter uint64 = 0
	var notSpace bool = false
	for !reader.IsEnd() && reader.GetReadPointer() < textSize {
		bgn, text, end, err := reader.ReadBlock(0, textSize, int(opts.BlockSize))
		if err != nil {
			return err
		}
		text, notSpace = RelaxChars(opts, text, notSpace, iter == charBlock)
		if err := CheckErrorsAndWrite(writer, bgn); err != nil {
			return err
		}
		if !(iter > charBlock && opts.Conv.IsContain(TrimSpaces)) {
			if err := CheckErrorsAndWrite(writer, text); err != nil {
				return err
			}
		}
		if err := CheckErrorsAndWrite(writer, end); err != nil {
			return err
		}

		iter++
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

func CheckErrorsAndWrite(writer BlockWriter, text []byte) error {
	_, err := writer.Write(text)
	return err
}

func TransmitSegment(opts *Options, reader BlockReader, writer BlockWriter, l, r uint64) (uint64, uint64, error) {
	var totalLen, iter, charBlock uint64 = 0, 0, 0
	for !reader.IsEnd() && reader.GetReadPointer() < r {
		bgn, text, end, err := reader.ReadBlock(l, r, int(opts.BlockSize))
		if err != nil {
			return 0, 0, nil
		}
		totalLen += uint64(len(bgn)) + uint64(len(text)) + uint64(len(end))
		if !IsSpaceBlock(text) {
			charBlock = iter
		}
		if err := CheckErrorsAndWrite(writer, bgn); err != nil {
			return 0, 0, err
		}
		if err := CheckErrorsAndWrite(writer, text); err != nil {
			return 0, 0, err
		}
		if err := CheckErrorsAndWrite(writer, end); err != nil {
			return 0, 0, err
		}
		iter++
	}
	return totalLen, charBlock, nil
}

func ReadFromStdin(opts *Options) (uint64, uint64, error) {
	writer, err := os.Create(TransferFileName)
	if err != nil {
		return 0, 0, err
	}
	reader := CustomReader{*bufio.NewReader(os.Stdin), 0, false}
	l := opts.Offset
	r := opts.Offset + opts.Limit
	return TransmitSegment(opts, &reader, writer, l, r)
}

func ReadFromFile(opts *Options) (uint64, uint64, error) {
	writer, err := os.Create(TransferFileName)
	if err != nil {
		return 0, 0, err
	}
	f, err := os.Open(opts.From)
	if err != nil {
		return 0, 0, err
	}
	reader := CustomReader{*bufio.NewReader(f), 0, false}
	l := opts.Offset
	r := opts.Offset + opts.Limit

	textLen, charBlock, err := TransmitSegment(opts, &reader, writer, l, r)
	e := f.Close()
	if err != nil {
		return textLen, charBlock, err
	}
	return textLen, charBlock, e
}

type InputFormatError struct{}

func (d InputFormatError) Error() string {
	return "Incorrect input data format"
}

type WrapError struct {
	Context string
	Err     error
}

func (d WrapError) Error() string {
	if d.Err != nil {
		return d.Context + ": " + d.Err.Error()
	}
	return ""
}

func (d WrapError) Unwrap() error {
	return d.Err
}

func BasicCheck(opts *Options) error {
	if opts.To != "" {
		_, err := os.Stat(opts.To)
		if err == nil {
			return WrapError{"A file with that name already exists", InputFormatError{}}
		}
	}

	if opts.Conv.IsContain(LowerCase) && opts.Conv.IsContain(UpperCase) {
		return WrapError{"using lower_case and upper_case flags together", InputFormatError{}}
	}
	for i := 0; i < len(opts.Conv); i++ {
		if opts.Conv[i] != LowerCase && opts.Conv[i] != UpperCase && opts.Conv[i] != TrimSpaces {
			return WrapError{"unexpected conv's argument", InputFormatError{}}
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
	err = BasicCheck(opts)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, errors.Unwrap(err))
		os.Exit(1)
	}

	var textLen, charBlock uint64
	if len(opts.From) == 0 {
		textLen, charBlock, err = ReadFromStdin(opts)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Can't transmit from stdin to the transfer file", err)
			os.Exit(1)
		}
	} else {
		textLen, charBlock, err = ReadFromFile(opts)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Can't transmit from input file to the transfer file", err)
			os.Exit(1)
		}
	}

	if textLen == 0 {
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
		RebuildText(opts, &reader, os.Stdout, textLen, charBlock)
	} else {
		fileTo, err := os.Create(opts.To)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "incorrect output file", err)
			os.Exit(1)
		}
		RebuildText(opts, &reader, fileTo, textLen, charBlock)
	}
	err = f.Close()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "can't close file", err)
		os.Exit(1)
	}
}
