package epubx

import (
	"archive/zip"
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const filename = "test-data/The_Dark_Elf.epub"

func TestParseEPUB(t *testing.T) {
	t.Parallel()

	f, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	require.NoError(t, err)

	fs, err := f.Stat()
	require.NoError(t, err)

	epub, err := ParseEPUB(f, fs.Size())
	require.NoError(t, err)

	assert.Equal(t, "2.0", epub.Version)
	assert.Equal(t, "uuid_id", epub.UniqueIdentifier)

	meta := epub.Metadata

	assert.Equal(t, []Identifier{
		{ID: "5-94955-003-X", Scheme: "ISBN"},
		{ID: "07aada78-943e-44a1-b680-4b2edba53ba9", Scheme: "uuid"},
	}, meta.Identifiers)

	assert.Equal(t, []string{"Воин"}, meta.Titles)
	assert.Equal(t, []string{"ru"}, meta.Languages)

	assert.Equal(t, []Author{
		{Name: "Роберт Энтони Сальваторе", Role: "aut"},
	}, meta.Creators)

	assert.Equal(t, []Author{
		{Name: "calibre (2.55.0) [http://calibre-ebook.com]", Role: "bkp"},
	}, meta.Contributors)

	assert.Equal(t, []string{"ИЦ «Максима»"}, meta.Publishers)
	assert.Equal(t, []string{"sf_fantasy"}, meta.Subjects)
	assert.Contains(t, meta.Description, "Покинув подземный мир, темный эльф Дзирт До'Урден")
	assert.Equal(t, []string{"2007-09-15T00:00:00+00:00"}, meta.Dates)
	assert.Empty(t, meta.Rights)
	assert.Len(t, meta.Meta, 4)
}

// zipWithFile creates an in-memory zip containing a single file with the given name and data,
// then returns the *zip.File for that entry.
func zipWithFile(t *testing.T, name string, data []byte) *zip.File {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	fw, err := w.Create(name)
	require.NoError(t, err)
	_, err = fw.Write(data)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	return r.File[0]
}

// buildEPUBZip creates an in-memory zip with the given ordered files and returns a *bytes.Reader.
// Files are written in the order provided. The mimetype entry uses zip.Store (uncompressed) per EPUB spec.
func buildEPUBZip(t *testing.T, files ...zipEntry) *bytes.Reader {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, f := range files {
		method := zip.Deflate
		if f.name == "mimetype" {
			method = zip.Store
		}
		fw, err := w.CreateHeader(&zip.FileHeader{Name: f.name, Method: method})
		require.NoError(t, err)
		_, err = fw.Write(f.data)
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	return bytes.NewReader(buf.Bytes())
}

type zipEntry struct {
	name string
	data []byte
}

var (
	validMimetype  = []byte("application/epub+zip")
	validContainer = []byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`)
	minimalOPF = []byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
  </metadata>
</package>`)
)

func nestedOPFContainer(path string) []byte {
	return []byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="` + path + `" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`)
}

func TestParseEPUB_errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		files   []zipEntry
		wantErr error
	}{
		{
			name:    "missing container.xml",
			files:   []zipEntry{{name: "mimetype", data: validMimetype}},
			wantErr: ErrMissingContainerXML,
		},
		{
			name: "missing mimetype",
			files: []zipEntry{
				{name: "META-INF/container.xml", data: validContainer},
				{name: "content.opf", data: minimalOPF},
			},
			wantErr: ErrMimetypeNotFirst,
		},
		{
			name: "mimetype not first file",
			files: []zipEntry{
				{name: "META-INF/container.xml", data: validContainer},
				{name: "mimetype", data: validMimetype},
				{name: "content.opf", data: minimalOPF},
			},
			wantErr: ErrMimetypeNotFirst,
		},
		{
			name: "invalid mimetype content",
			files: []zipEntry{
				{name: "mimetype", data: []byte("text/plain")},
				{name: "META-INF/container.xml", data: validContainer},
				{name: "content.opf", data: minimalOPF},
			},
			wantErr: ErrInvalidMimetype,
		},
		{
			name: "missing content.opf pointed by container",
			files: []zipEntry{
				{name: "mimetype", data: validMimetype},
				{name: "META-INF/container.xml", data: nestedOPFContainer("OEBPS/package.opf")},
			},
			wantErr: ErrMissingContentOPF,
		},
		{
			name: "container.xml with empty rootfile path",
			files: []zipEntry{
				{name: "mimetype", data: validMimetype},
				{name: "META-INF/container.xml", data: nestedOPFContainer("")},
			},
			wantErr: ErrMissingRootFilePath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := buildEPUBZip(t, tt.files...)
			_, err := ParseEPUB(r, int64(r.Len()))
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestParseEPUB_nestedOPFPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opfPath string
	}{
		{name: "root level", opfPath: "content.opf"},
		{name: "OEBPS subdirectory", opfPath: "OEBPS/content.opf"},
		{name: "deeply nested", opfPath: "a/b/c/package.opf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := buildEPUBZip(t,
				zipEntry{name: "mimetype", data: validMimetype},
				zipEntry{name: "META-INF/container.xml", data: nestedOPFContainer(tt.opfPath)},
				zipEntry{name: tt.opfPath, data: minimalOPF},
			)

			epub, err := ParseEPUB(r, int64(r.Len()))
			require.NoError(t, err)
			assert.Equal(t, "3.0", epub.Version)
			assert.Equal(t, "uid", epub.UniqueIdentifier)
			assert.Equal(t, []string{"Test"}, epub.Metadata.Titles)
			assert.Equal(t, []string{"en"}, epub.Metadata.Languages)
		})
	}
}

func Test_parseContainer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		xml     string
		want    string
		wantErr bool
	}{
		{
			name: "standard container",
			xml: `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
   <rootfiles>
      <rootfile full-path="content.opf" media-type="application/oebps-package+xml"/>
   </rootfiles>
</container>`,
			want: "content.opf",
		},
		{
			name: "nested opf path",
			xml: `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
   <rootfiles>
      <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
   </rootfiles>
</container>`,
			want: "OEBPS/content.opf",
		},
		{
			name: "empty full-path",
			xml: `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
   <rootfiles>
      <rootfile full-path="" media-type="application/oebps-package+xml"/>
   </rootfiles>
</container>`,
			wantErr: true,
		},
		{
			name: "missing rootfile element",
			xml: `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
   <rootfiles>
   </rootfiles>
</container>`,
			wantErr: true,
		},
		{
			name:    "invalid xml",
			xml:     `not xml at all`,
			wantErr: true,
		},
		{
			name:    "empty file",
			xml:     ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			zf := zipWithFile(t, "META-INF/container.xml", []byte(tt.xml))

			got, err := parseContainer(zf)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
