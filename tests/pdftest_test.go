package tests

import (
	"os"
	"testing"
)

func TestPdftest(t *testing.T) {
	os.Chdir("..")
	Pdftest()
}
