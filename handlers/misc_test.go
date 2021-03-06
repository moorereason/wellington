package handlers

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	libsass "github.com/wellington/go-libsass"
)

func TestFontURLFail(t *testing.T) {
	r, w, _ := os.Pipe()
	_ = r
	old := os.Stdout
	defer func() { os.Stdout = old }()
	os.Stdout = w
	in := bytes.NewBufferString(`@font-face {
  src: font-url("arial.eot");
}`)
	var out bytes.Buffer
	ctx := libsass.NewContext()
	err := ctx.Compile(in, &out)

	e := "error in C function font-url: font-url: font path not set"
	if !strings.Contains(err.Error(), e) {
		t.Errorf("got:\n%s\nwanted:\n%s\n", err, e)
	}

	// Removed this as part of making font-url fail instead of
	// output garbage
	//
	// outC := make(chan string)
	// go func(r *os.File) {
	// 	var buf bytes.Buffer
	// 	io.Copy(&buf, r)
	// 	outC <- buf.String()
	// }(r)

	// w.Close()
	// stdout := <-outC

}

func ExampleFontURL() {
	in := bytes.NewBufferString(`
$path: font-url($raw: true, $path: "arial.eot");
@font-face {
  src: font-url("arial.eot");
  src: url("#{$path}");
}`)

	_, _, err := setupCtx(in, os.Stdout)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// @font-face {
	//   src: url("../font/arial.eot");
	//   src: url("../font/arial.eot"); }

}

func TestFontURL_invalid(t *testing.T) {
	r, w, _ := os.Pipe()
	_, _ = r, w
	old := os.Stdout
	defer func() { os.Stdout = old }()
	//os.Stdout = w
	in := bytes.NewBufferString(`@font-face {
  src: font-url(5px);
}`)
	var out bytes.Buffer
	ctx := libsass.NewContext()
	err := ctx.Compile(in, &out)

	e := `Error > stdin:2
error in C function font-url: Invalid Sass type expected: string got: libs.SassNumber value: 5px`
	if !strings.HasPrefix(err.Error(), e) {
		t.Errorf("got:\n%s\nwanted:\n%s", err, e)
	}
}
