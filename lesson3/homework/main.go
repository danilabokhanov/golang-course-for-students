package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
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
	LowerCase         = "lower_case"
	TrimSpaces        = "trim_spaces"
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

func (d *ConvArgs) IsContain(s string) bool {
	for i := 0; i < len(*d); i++ {
		if (*d)[i] == s {
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
	// todo: add required flags
}

type BlockReader interface {
	ReadBlock(uint64, uint64, int) ([]byte, []byte, []byte)
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

func DeleteBrokenBytesFromEnd(t *[]byte) []byte {
	end := []byte{}
	for len(*t) > 0 && !utf8.Valid(*t) {
		end = append(end, (*t)[len(*t)-1])
		*t = (*t)[:len(*t)-1]
	}
	for i := 0; i < len(end)/2; i++ {
		end[i], end[len(end)-1-i] = end[len(end)-1-i], end[i]
	}
	return end
}

func (d *CustomReader) ReadBlock(l, r uint64, size int) ([]byte, []byte, []byte) {
	for d.ptr < l {
		_, err := d.Rd.ReadByte()
		d.ptr++
		if err == io.EOF {
			d.isEnd = true
			return []byte{}, []byte{}, []byte{}
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	bgn := []byte{}
	text := []byte{}
	for d.ptr < r {
		cur, err := d.Rd.ReadByte()
		d.ptr++
		if err == io.EOF {
			d.isEnd = true
			return bgn, []byte{}, []byte{}
		}
		if err != nil {
			log.Fatal(err)
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
			return bgn, text, []byte{}
		}
		text = append(text, cur)
		if err != nil {
			log.Fatal(err)
		}

		if len(text) >= size && utf8.Valid(text) {
			return bgn, text, []byte{}
		}
	}
	end := DeleteBrokenBytesFromEnd(&text)
	return bgn, text, end
}

func RelaxChars(opts *Options, text *[]byte, notSpace *bool, isLastCharBlock bool) {
	s := bytes.Runes(*text)
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
		if opts.Conv.IsContain(TrimSpaces) && unicode.IsSpace(s[i]) && !(*notSpace) {
			continue
		}
		if !unicode.IsSpace(s[i]) {
			*notSpace = true
		}
		if opts.Conv.IsContain(UpperCase) {
			res = append(res, unicode.ToUpper(s[i]))
		} else if opts.Conv.IsContain(LowerCase) {
			res = append(res, unicode.ToLower(s[i]))
		} else {
			res = append(res, s[i])
		}
	}
	*text = []byte(string(res))
}

func RebuildText(opts *Options, reader BlockReader, writer BlockWriter, textSize, charBlock uint64) {
	var iter uint64 = 0
	var notSpace bool = false
	for !reader.IsEnd() && reader.GetReadPointer() < textSize {
		bgn, text, end := reader.ReadBlock(0, textSize, int(opts.BlockSize))

		RelaxChars(opts, &text, &notSpace, iter == charBlock)
		CheckErrorsAndWrite(writer, &bgn)
		if !(iter > charBlock && opts.Conv.IsContain(TrimSpaces)) {
			CheckErrorsAndWrite(writer, &text)
		}
		CheckErrorsAndWrite(writer, &end)

		iter++
	}
}

func IsSpaceBlock(text *[]byte) bool {
	s := bytes.Runes(*text)
	for i := 0; i < len(s); i++ {
		if !unicode.IsSpace(s[i]) {
			return false
		}
	}
	return true
}

func CheckErrorsAndWrite(writer BlockWriter, text *[]byte) {
	_, err := writer.Write(*text)
	if err != nil {
		log.Fatal(err)
	}
}

func TransmitSegment(opts *Options, reader BlockReader, writer BlockWriter, l, r uint64) (uint64, uint64) {
	var totalLen, iter, charBlock uint64 = 0, 0, 0
	for !reader.IsEnd() && reader.GetReadPointer() < r {
		bgn, text, end := reader.ReadBlock(l, r, int(opts.BlockSize))
		totalLen += uint64(len(bgn)) + uint64(len(text)) + uint64(len(end))
		if !IsSpaceBlock(&text) {
			charBlock = iter
		}
		CheckErrorsAndWrite(writer, &bgn)
		CheckErrorsAndWrite(writer, &text)
		CheckErrorsAndWrite(writer, &end)
		iter++
	}
	return totalLen, charBlock
}

func ReadFromStdin(opts *Options) (uint64, uint64) {
	writer, err := os.Create(TransferFileName)
	if err != nil {
		log.Fatal(err)
	}
	reader := CustomReader{*bufio.NewReader(os.Stdin), 0, false}
	l := opts.Offset
	r := opts.Offset + opts.Limit
	return TransmitSegment(opts, &reader, writer, l, r)
}

func ReadFromFile(opts *Options) (uint64, uint64) {
	writer, err := os.Create(TransferFileName)
	if err != nil {
		log.Fatal(err)
	}
	f, e := os.Open(opts.From)
	if e != nil {
		log.Fatal(e)
	}
	reader := CustomReader{*bufio.NewReader(f), 0, false}
	l := opts.Offset
	r := opts.Offset + opts.Limit
	return TransmitSegment(opts, &reader, writer, l, r)
}

func BasicCheck(opts *Options) {
	if opts.To != "" {
		_, err := os.Stat(opts.To)
		if err == nil {
			_, _ = fmt.Fprintln(os.Stderr, "A file with that name already exists")
			os.Exit(1)
		}
	}

	if opts.Conv.IsContain(LowerCase) && opts.Conv.IsContain(UpperCase) {
		_, _ = fmt.Fprintln(os.Stderr, "using lower_case and upper_case flags together")
		os.Exit(1)
	}
	for i := 0; i < len(opts.Conv); i++ {
		if opts.Conv[i] != LowerCase && opts.Conv[i] != UpperCase && opts.Conv[i] != TrimSpaces {
			_, _ = fmt.Fprintln(os.Stderr, "unexpected conv's argument:", opts.Conv[i])
			os.Exit(1)
		}
	}
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

	// todo: parse and validate all flags

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
	BasicCheck(opts)

	var textLen, charBlock uint64
	if len(opts.From) == 0 {
		textLen, charBlock = ReadFromStdin(opts)
	} else {
		textLen, charBlock = ReadFromFile(opts)
	}

	if textLen == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "offset is too large")
		os.Exit(1)
	}

	f, e := os.Open(TransferFileName)
	if e != nil {
		log.Fatal(e)
	}
	reader := CustomReader{*bufio.NewReader(f), 0, false}
	if len(opts.To) == 0 {
		RebuildText(opts, &reader, os.Stdout, textLen, charBlock)
	} else {
		fileTo, er := os.Create(opts.To)
		if er != nil {
			log.Fatal(er)
		}
		RebuildText(opts, &reader, fileTo, textLen, charBlock)
	}

	// todo: implement the functional requirements described in read.me
}
