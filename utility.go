package gz

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/form3tech-oss/jwt-go"
	gzerrors "github.com/gazebo-web/gz-go/v9/errors"
)

// GetUserIdentity returns the user identity found in the http request's JWT
// token.
func GetUserIdentity(r *http.Request) (identity string, ok bool) {
	// We use the claimed subject contained in the JWT as the ID.
	jwtUser := r.Context().Value("user")
	if jwtUser == nil {
		return
	}
	var sub interface{}
	sub, ok = jwtUser.(*jwt.Token).Claims.(jwt.MapClaims)["sub"]
	if !ok {
		return
	}
	identity, ok = sub.(string)
	return
}

// ReadEnvVar reads an environment variable and return an error if not present
func ReadEnvVar(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", errors.New("Missing " + name + " env variable.")
	}
	return value, nil
}

// Unzip a memory buffer
// TODO: remove. Unused code.
func Unzip(buff bytes.Buffer, size int64, dest string, verbose bool) error {
	reader, err := zip.NewReader(bytes.NewReader(buff.Bytes()), size)
	if err != nil {
		return errors.New("unzip: Unable to read byte buffer")
	}
	return UnzipImpl(reader, dest, verbose)
}

// UnzipFile extracts a compressed .zip file
// TODO: remove. Unused code.
func UnzipFile(zipfile string, dest string, verbose bool) error {
	reader, err := zip.OpenReader(zipfile)
	if err != nil {
		return errors.New("unzip: Unable to open [" + zipfile + "]")
	}
	defer reader.Close()
	return UnzipImpl(&reader.Reader, dest, verbose)
}

// UnzipImpl is a helper unzip implementation
// TODO: remove. Unused code.
func UnzipImpl(reader *zip.Reader, dest string, verbose bool) error {
	for _, f := range reader.File {
		zipped, err := f.Open()
		if err != nil {
			return errors.New("unzip: Unable to open [" + f.Name + "]")
		}

		defer func() {
			_ = zipped.Close()
		}()

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(path, f.Mode())
			if verbose {
				fmt.Println("Creating directory", path)
			}
		} else {
			// Ensure we create the parent folder
			err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err != nil {
				return errors.New("unzip: Unable to create parent folder [" + path + "]")
			}

			writer, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, f.Mode())
			if err != nil {
				return errors.New("unzip: Unable to create [" + path + "]")
			}

			defer writer.Close()

			if _, err = io.Copy(writer, zipped); err != nil {
				return errors.New("unzip: Unable to create content in [" + path + "]")
			}

			if verbose {
				fmt.Println("Decompressing : ", path)
			}
		}
	}
	return nil
}

// Trace returns the filename, line and function name of its caller. The skip
// parameter is the number of stack frames to skip, with 1 identifying the
// Trace frame itself. Skip will be set to 1 if the passed in value is <= 0.
// Ref: http://stackoverflow.com/questions/25927660/golang-get-current-scope-of-function-name
func Trace(skip int) string {
	skip = Max(skip, 1)

	// At least one entry needed
	pc := make([]uintptr, 10)
	count := Min(runtime.Callers(skip, pc), 10)

	result := ""
	for i := 0; i < count; i++ {
		f := runtime.FuncForPC(pc[i])
		file, line := f.FileLine(pc[i])
		result = fmt.Sprintf("%s%d  %s : %d in %s\n", result, i,
			filepath.Base(file), line, f.Name())
	}

	return result
}

// RandomString creates a random string of a given length.
// Ref: https://siongui.github.io/2015/04/13/go-generate-random-string/
func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// Min is an implementation of "int" Min
// See https://mrekucci.blogspot.com.ar/2015/07/dont-abuse-mathmax-mathmin.html
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Max is an implementation of "int" Max
// See https://mrekucci.blogspot.com.ar/2015/07/dont-abuse-mathmax-mathmin.html
func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// StrToSlice returns the slice of strings with all tags parsed from the input
// string.
// It will trim leading and trailing whitespace, and reduce middle whitespaces to 1 space.
// It will also remove 'empty' tags (ie. whitespaces enclosed with commas, ',   ,')
// The input string contains tags separated with commas.
// E.g. input string: " tag1, tag2,  tag3 ,   , "
// E.g. output: ["tag1", "tag2", "tag3"]
func StrToSlice(tagsStr string) []string {
	if tagsStr == "" {
		return nil
	}
	noSpaces := strings.TrimSpace(tagsStr)
	noSpaces = strings.TrimPrefix(noSpaces, ",")
	noSpaces = strings.TrimSuffix(noSpaces, ",")
	// regexp to remove duplicate spaces
	reInsideWhtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	var result []string
	for _, t := range strings.Split(noSpaces, ",") {
		t = strings.TrimSpace(t)
		if len(t) > 0 {
			result = append(result, reInsideWhtsp.ReplaceAllString(t, " "))
		}
	}
	return result
}

// SameElements returns True if the two given string slices contain the same
// elements, even in different order.
func SameElements(a0, b0 []string) bool {
	// shallow copy input arrays
	a := append([]string(nil), a0...)
	b := append([]string(nil), b0...)

	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ParseTemplate opens a template and replaces placeholders with values.
func ParseTemplate(templateFilename string, data interface{}) (string, error) {
	t, err := template.ParseFiles(templateFilename)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ParseHTMLTemplate opens an HTML template and replaces placeholders with values.
func ParseHTMLTemplate(templateFilename string, data interface{}) (string, error) {
	t, err := htmlTemplate.ParseFiles(templateFilename)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// IsError returns true when err is the target error.
func IsError(err error, target error) bool {
	return strings.Contains(err.Error(), target.Error())
}

func IsDirEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return true, err
	}
	if len(entries) == 0 {
		return true, nil
	}
	return false, nil
}

// CopyDir recursively copies the whole directory with its content from src to dst.
func CopyDir(dst, src string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	var info os.FileInfo
	if info, err = os.Stat(src); err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			if err = os.MkdirAll(filepath.Join(dst, name), info.Mode()); err != nil {
				return err
			}
			err := CopyDir(filepath.Join(dst, name), filepath.Join(src, name))
			if err != nil {
				return err
			}
		} else {
			err := CopyFile(filepath.Join(dst, name), filepath.Join(src, name))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CopyFile copies the file from src into dst. It assigns the same permissions as the original file.
func CopyFile(dst string, src string) error {
	// Source file
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	// Destination file
	fd, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = io.Copy(fd, f)
	if err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		return err
	}

	err = fd.Chmod(info.Mode())
	if err != nil {
		return err
	}
	return nil
}

// Zip the src folder into dst. It leaves the closing responsibility of the os.File to the consumer of this function.
func Zip(dst, src string) (*os.File, error) {
	if _, err := os.Stat(dst); errors.Is(err, os.ErrExist) {
		f, err := os.Open(dst)
		if err != nil {
			return nil, err
		}
		return f, nil
	}
	zipFile, err := os.Create(dst)
	if err != nil {
		return nil, err
	}

	w := zip.NewWriter(zipFile)
	defer w.Close()

	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Set compression
		header.Method = zip.Deflate

		// Set relative path of a file as the header name
		header.Name, err = filepath.Rel(filepath.Dir(src), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		// Create writer for the file header and save content of the file
		headerWriter, err := w.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
	if err != nil {
		return nil, err
	}

	return zipFile, nil
}

// Close runs the Close method from the io.Closer.
// If there's an error, it will print the error in the stdout.
func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Println("Failed to close:", err)
	}
}

// RemoveIfFound checks the file or directory in the given path exists, if it exists, it will attempt to remove it.
// If it's a directory, it must be empty.
func RemoveIfFound(path string) error {
	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

// ParseURL parses the given raw URL in string format to an url.URL.
func ParseURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	err = ValidateURL(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// ValidateURL validates if the given URL is well-formed or not.
func ValidateURL(u *url.URL) error {
	if u.Scheme == "" && u.Host == "" {
		return gzerrors.ErrInvalidURL
	}
	return nil
}
