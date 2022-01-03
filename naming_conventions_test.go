package copre

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitPathIntoWords(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]string{
		"Document,PDFLoader": "Document,PDF,Loader",
		"FooBar,FizzBuzz":    "Foo,Bar,Fizz,Buzz",
	}
	for inputStr, expectedStr := range tests {
		input := strings.Split(inputStr, ",")
		expected := strings.Split(expectedStr, ",")
		result := SplitPathIntoWords(input)
		assert.Equal(expected, result)
	}
}

func TestUpperSnakeCase(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]string{
		"Document,PDFLoader": "DOCUMENT_PDF_LOADER",
		"FooBar,FizzBuzz":    "FOO_BAR_FIZZ_BUZZ",
	}
	for inputStr, expected := range tests {
		input := strings.Split(inputStr, ",")
		result := UpperSnakeCase(input)
		assert.Equal(expected, result)
	}
}

func TestLowerSnakeCase(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]string{
		"Document,PDFLoader": "document_pdf_loader",
		"FooBar,FizzBuzz":    "foo_bar_fizz_buzz",
	}
	for inputStr, expected := range tests {
		input := strings.Split(inputStr, ",")
		result := LowerSnakeCase(input)
		assert.Equal(expected, result)
	}
}

func TestKebabCase(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]string{
		"Document,PDFLoader": "document-pdf-loader",
		"FooBar,FizzBuzz":    "foo-bar-fizz-buzz",
	}
	for inputStr, expected := range tests {
		input := strings.Split(inputStr, ",")
		result := KebabCase(input)
		assert.Equal(expected, result)
	}
}

func TestCamelCase(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]string{
		"Document,PDFLoader": "documentPDFLoader",
		"FooBar,FizzBuzz":    "fooBarFizzBuzz",
	}
	for inputStr, expected := range tests {
		input := strings.Split(inputStr, ",")
		result := CamelCase(input)
		assert.Equal(expected, result)
	}
}

func TestPascalCase(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]string{
		"Document,PDFLoader": "DocumentPDFLoader",
		"FooBar,FizzBuzz":    "FooBarFizzBuzz",
	}
	for inputStr, expected := range tests {
		input := strings.Split(inputStr, ",")
		result := PascalCase(input)
		assert.Equal(expected, result)
	}
}
