package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html"
	"html/template"
	"net/url"
	"path/filepath"
)

// ThumbnailParams to generate links and HTML tags for images.
type ThumbnailParams struct {
	// Type of the output format (i.e., jpeg, png, webp; default: auto)
	Type string

	// Quality of the image (default, 90).
	Quality int

	// Width of the thumbnail.
	Width int

	// Height of the thumbnail.
	Height int

	// Method to call in imaginary for image transformation.
	// By default fit is used. See https://github.com/h2non/imaginary.
	Method string
}

// Images service.
type Images struct {
	core *Core
}

// Link of the image thumbnail.
func (i *Images) Link(path string, p ThumbnailParams) (string, error) {
	settings := i.core.Settings
	u, err := url.Parse(settings.ThumbnailServiceHost)
	if err != nil {
		return "", fmt.Errorf("cannot parse image thumbnail service address: %w", err)
	}
	method := p.Method
	if method == "" {
		method = "resize"
	}
	u.Path = "/" + method

	// TODO(henvic): Accept requests from "https://www." (maybe use the "www:" prefix).
	us, err := url.Parse(settings.FileStorageHost)
	if err != nil {
		return "", fmt.Errorf("cannot parse file storage service address: %w", err)
	}
	us.Path = filepath.Join(us.Path, path)
	q := u.Query()
	q.Set("url", us.String())
	if p.Type != "" {
		q.Set("type", p.Type)
	}
	if p.Quality == 0 {
		p.Quality = 90 // Default image quality value.
	}
	q.Set("quality", fmt.Sprint(p.Quality))
	if p.Width != 0 {
		q.Set("width", fmt.Sprint(p.Width))
	}
	if p.Height != 0 {
		q.Set("height", fmt.Sprint(p.Height))
	}
	u.RawQuery = q.Encode()
	return u.String() + "&sign=" + imageURLsignature("", u), nil // Append signature to the end, otherwise we change its value.
}

// HTML returns a img tag with src and srcset for regular and Retina display formats.
func (i *Images) HTML(path string, alt string, p ThumbnailParams) (template.HTML, error) {
	src1x, err := i.Link(path, p)
	if err != nil {
		return "", err
	}
	// Resize parameter to generate 'Retina display' thumbnails.
	p.Width *= 2
	p.Height *= 2
	src2x, err := i.Link(path, p)
	if err != nil {
		return "", err
	}
	img := `<img src="` + html.EscapeString(src1x) + `" srcset="` + html.EscapeString(src1x) + ` 1x, ` + html.EscapeString(src2x) + ` 2x"`
	if alt != "" {
		img += ` alt="` + html.EscapeString(alt) + `"`
	}
	img += `>`
	return template.HTML(img), nil
}

func imageURLsignature(signKey string, u *url.URL) string {
	h := hmac.New(sha256.New, []byte(signKey))
	h.Write([]byte(u.Path))
	h.Write([]byte(u.Query().Encode()))
	buf := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(buf)
}
