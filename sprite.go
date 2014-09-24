package sprite_sass

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/rainycape/magick"
)

type Images []*magick.Image

type ImageList struct {
	Images
	Out      *magick.Image
	OutFile  string
	Combined bool
	Files    []string
	Vertical bool
}

func (l ImageList) String() string {
	files := ""
	for _, file := range l.Files {
		files += strings.TrimSuffix(filepath.Base(file),
			filepath.Ext(file)) + " "
	}
	return files
}

func (l ImageList) Lookup(f string) int {
	var base string
	for i, v := range l.Files {
		base = filepath.Base(v)
		base = strings.TrimSuffix(base, filepath.Ext(v))
		if f == v {
			return i
			//Do partial matches, for now
		} else if f == base {
			return i
		}
	}
	// TODO: Find a better way to send these to cli so tests
	// aren't impacted.
	// Debug.Printf("File not found: %s\n Try one of %s", f, l)

	return -1
}

// Return the X position of an image based
// on the layout (vertical/horizontal) and
// position in Image slice
func (l ImageList) X(pos int) int {
	x := 0
	if l.Vertical {
		return 0
	}
	for i := 0; i < pos; i++ {
		x += l.Images[i].Width()
	}
	return x
}

// Return the Y position of an image based
// on the layout (vertical/horizontal) and
// position in Image slice
func (l ImageList) Y(pos int) int {
	y := 0
	if !l.Vertical {
		return 0
	}
	for i := 0; i < pos; i++ {
		y += l.Images[i].Height()
	}
	return y
}

func (l ImageList) CSS(s string) string {
	pos := l.Lookup(s)
	if pos == -1 {
		log.Printf("File not found: %s\n Try one of: %s",
			s, l)
	}
	if l.OutFile == "" {
		return "transparent"
	}
	return fmt.Sprintf(`url("%s") %s`,
		l.OutFile, l.Position(s))
}

func (l ImageList) Position(s string) string {
	pos := l.Lookup(s)
	if pos == -1 {
		log.Printf("File not found: %s\n Try one of: %s",
			s, l)
	}

	return fmt.Sprintf(`%dpx %dpx`, -l.X(pos), -l.Y(pos))
}

func (l ImageList) Dimensions(s string) string {
	if pos := l.Lookup(s); pos > -1 {

		return fmt.Sprintf("width: %dpx;\nheight: %dpx",
			l.Images[pos].Width(), l.Images[pos].Height())
	}
	return ""
}

func (l ImageList) Inline() string {
	info := magick.NewInfo()
	info.SetFormat("png")
	r, w := io.Pipe()
	go func(w io.WriteCloser, info *magick.Info) {
		err := l.Images[0].Encode(w, info)
		if err != nil {
			panic(err)
		}
		w.Close()
	}(w, info)
	var scanned []byte
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanBytes)
	for scanner.Scan() {
		scanned = append(scanned, scanner.Bytes()...)
	}
	encstr := base64.StdEncoding.EncodeToString(scanned)
	return fmt.Sprintf("url('data:image/png;base64,%s')", encstr)
}

func (l ImageList) ImageWidth(s string) int {
	if pos := l.Lookup(s); pos > -1 {
		return l.Images[pos].Width()
	}
	return -1
}
func (l ImageList) ImageHeight(s string) int {
	if pos := l.Lookup(s); pos > -1 {
		return l.Images[pos].Height()
	}
	return -1
}

// Return the cumulative Height of the
// image slice.
func (l *ImageList) Height() int {
	h := 0
	ll := *l

	for _, img := range ll.Images {
		if l.Vertical {
			h += img.Height()
		} else {
			h = int(math.Max(float64(h), float64(img.Height())))
		}
	}
	return h
}

// Return the cumulative Width of the
// image slice.
func (l *ImageList) Width() int {
	w := 0

	for _, img := range l.Images {
		if !l.Vertical {
			w += img.Width()
		} else {
			w = int(math.Max(float64(w), float64(img.Width())))
		}
	}
	return w
}

// Accept a variable number of image globs appending
// them to the ImageList.
func (l *ImageList) Decode(rest ...string) error {

	// Invalidate the composite cache
	l.Out = nil
	var (
		paths []string
		ext   string
	)

	for _, r := range rest {
		matches, err := filepath.Glob(r)
		if err != nil {
			panic(err)
		}
		paths = append(paths, matches...)
	}

	if len(paths) > 0 {
		ext = filepath.Ext(paths[0])
		l.OutFile = filepath.Dir(paths[0]) + "-" +
			randString(6) + ext
	}

	for _, path := range paths {
		img, err := magick.DecodeFile(path)
		if err != nil {
			return err
		}
		l.Images = append(l.Images, img)
		l.Files = append(l.Files, path)
	}

	return nil
}

// Combine all images in the slice into a final output
// image.
func (l *ImageList) Combine() {

	var (
		maxW, maxH int
	)

	if l.Out != nil {
		return
	}

	maxW, maxH = l.Width(), l.Height()

	curH, curW := 0, 0
	l.Out, _ = magick.New(maxW, maxH)

	for _, img := range l.Images {
		err := l.Out.Composite(magick.CompositeCopy, img, curW, curH)
		if err != nil {
			panic(err)
		}
		if l.Vertical {
			curH += img.Height()
		} else {
			curW += img.Width()
		}
	}

	l.Combined = true
}

func randString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

// Export saves out the ImageList to a the specified file
func (l *ImageList) Export(path string) error {
	// Use the auto generated path if none is specified
	if path != "" {
		l.OutFile = path
	} else {
		path = l.OutFile
	}
	// Remove invalid characters from path
	path = strings.Replace(path, "/", "", -1)
	path = strings.Replace(path, "*", "", -1)

	fo, err := os.Create(l.OutFile)
	if err != nil {
		return err
	}

	//This call is cached if already run
	l.Combine()

	// Supported compressions http://www.imagemagick.org/RMagick/doc/info.html#compression
	defer fo.Close()

	if err != nil {
		return err
	}

	frmt := magick.NewInfo()
	frmt.SetFormat(strings.ToUpper(filepath.Ext(path)[1:]))

	err = l.Out.Encode(fo, frmt)
	if err != nil {
		return err
	}
	return nil
}
