package epubx

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"io"
)

const containerPath = "META-INF/container.xml"

var (
	ErrInvalidMimetype     = errors.New("invalid mime type")
	ErrMissingMimetype     = errors.New("missing mimetype")
	ErrMissingContentOPF   = errors.New("missing content.opf")
	ErrMissingContainerXML = errors.New("missing container.xml")
)

// Package represents the root of the .opf file
type Package struct {
	XMLName          xml.Name `xml:"http://www.idpf.org/2007/opf package"`
	Version          string   `xml:"version,attr"`
	UniqueIdentifier string   `xml:"unique-identifier,attr"`
	Metadata         Metadata `xml:"metadata"`
}

// Metadata contains the Dublin Core elements and EPUB 3 meta properties
type Metadata struct {
	// Dublin Core Fields
	Identifiers  []Identifier `xml:"http://purl.org/dc/elements/1.1/ identifier"`
	Titles       []string     `xml:"http://purl.org/dc/elements/1.1/ title"`
	Languages    []string     `xml:"http://purl.org/dc/elements/1.1/ language"`
	Creators     []Author     `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Contributors []Author     `xml:"http://purl.org/dc/elements/1.1/ contributor"`
	Publishers   []string     `xml:"http://purl.org/dc/elements/1.1/ publisher"`
	Subjects     []string     `xml:"http://purl.org/dc/elements/1.1/ subject"`
	Description  string       `xml:"http://purl.org/dc/elements/1.1/ description"`
	Dates        []string     `xml:"http://purl.org/dc/elements/1.1/ date"`
	Rights       string       `xml:"http://purl.org/dc/elements/1.1/ rights"`

	// EPUB 3 Meta Properties (for Last Modified, etc.)
	Meta []MetaProperty `xml:"meta"`
}

type Identifier struct {
	ID     string `xml:",chardata"`
	Scheme string `xml:"scheme,attr,omitempty"`
}

type Author struct {
	Name string `xml:",chardata"`
	ID   string `xml:"id,attr,omitempty"`
	Role string `xml:"role,attr,omitempty"` // EPUB 2
}

type MetaProperty struct {
	Property string `xml:"property,attr"`
	Refines  string `xml:"refines,attr,omitempty"`
	ID       string `xml:"id,attr,omitempty"`
	Value    string `xml:",chardata"`
}

type EPUB struct {
	Package
}

func ParseEPUB(r io.ReaderAt, size int64) (EPUB, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return EPUB{}, err
	}

	fileMap := make(map[string]*zip.File, len(zr.File))
	for _, zfile := range zr.File {
		fileMap[zfile.Name] = zfile
	}

	if fileMap[containerPath] == nil {
		return EPUB{}, ErrMissingContainerXML
	}

	contentOPFFilePath, err := parseContainer(fileMap[containerPath])
	if err != nil {
		return EPUB{}, err
	}
	if fileMap["mimetype"] == nil {
		return EPUB{}, ErrMissingMimetype
	}
	if fileMap[contentOPFFilePath] == nil {
		return EPUB{}, ErrMissingContentOPF
	}

	mr, err := fileMap["mimetype"].Open()
	if err != nil {
		return EPUB{}, err
	}
	defer mr.Close()

	mdata, err := io.ReadAll(mr)
	if err != nil {
		return EPUB{}, err
	}
	if string(mdata) != "application/epub+zip" {
		return EPUB{}, ErrInvalidMimetype
	}

	cr, err := fileMap[contentOPFFilePath].Open()
	if err != nil {
		return EPUB{}, err
	}
	defer cr.Close()

	var epub EPUB
	err = xml.NewDecoder(cr).Decode(&epub)
	if err != nil {
		return EPUB{}, err
	}

	return epub, nil
}

func parseContainer(f *zip.File) (string, error) {
	r, err := f.Open()
	if err != nil {
		return "", err
	}
	defer r.Close()

	var container struct {
		ContentPath struct {
			FullPath  string `xml:"full-path,attr"`
			MediaType string `xml:"media-type,attr"`
		} `xml:"rootfiles>rootfile"`
	}
	err = xml.NewDecoder(r).Decode(&container)
	if err != nil {
		return "", err
	}

	return container.ContentPath.FullPath, nil
}
