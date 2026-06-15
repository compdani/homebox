package custom

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/hay-kot/homebox/backend/internal/pocketbase/collections"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

const labelDPI = 300

var (
	locationLabelSize = image.Pt(labelDPI*315/100, labelDPI*197/100) // 3.15" x 1.97" landscape
	productLabelSize  = image.Pt(labelDPI*157/100, labelDPI*118/100) // 1.57" x 1.18" landscape
)

func registerLabelRoutes(g *router.RouterGroup[*core.RequestEvent], deps Deps) {
	g.GET("/locations/{id}/label.png", func(e *core.RequestEvent) error {
		return writeCollectionLabel(e, deps, collections.Locations, locationLabelSize, func(rec *core.Record) string {
			return requestOrigin(e) + "/location/" + rec.Id
		})
	})
	g.GET("/products/{id}/label.png", func(e *core.RequestEvent) error {
		return writeCollectionLabel(e, deps, collections.Products, productLabelSize, func(rec *core.Record) string {
			return requestOrigin(e) + "/product/" + rec.Id
		})
	})
	g.GET("/items/{id}/label.png", func(e *core.RequestEvent) error {
		rec, err := deps.Store.GetGroupRecord(collections.Items, e.Request.PathValue("id"), authGroupID(e))
		if err != nil {
			return e.NotFoundError("item not found", err)
		}
		if rec.GetString("product") != "" {
			return e.BadRequestError("product-linked items use the product label", nil)
		}
		return writeLabelResponse(e, rec.GetString("name"), requestOrigin(e)+"/item/"+rec.Id, productLabelSize)
	})
}

func writeCollectionLabel(
	e *core.RequestEvent,
	deps Deps,
	collection string,
	size image.Point,
	qrURL func(*core.Record) string,
) error {
	rec, err := deps.Store.GetGroupRecord(collection, e.Request.PathValue("id"), authGroupID(e))
	if err != nil {
		return e.NotFoundError("record not found", err)
	}
	return writeLabelResponse(e, rec.GetString("name"), qrURL(rec), size)
}

func writeLabelResponse(e *core.RequestEvent, title, data string, size image.Point) error {
	img, err := renderLabel(title, data, size)
	if err != nil {
		return e.InternalServerError("failed to render label", err)
	}
	e.Response.Header().Set("Content-Type", "image/png")
	e.Response.Header().Set("Content-Disposition", "attachment; filename=label.png")
	return png.Encode(e.Response, img)
}

func renderLabel(title, data string, size image.Point) (image.Image, error) {
	qrBytes, err := encodeQR(data)
	if err != nil {
		return nil, err
	}
	qrImg, _, err := image.Decode(bytes.NewReader(qrBytes))
	if err != nil {
		return nil, err
	}

	canvas := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	padding := size.Y / 10
	if padding < 10 {
		padding = 10
	}
	gap := padding / 2
	if gap < 6 {
		gap = 6
	}

	// Keep QR square but cap width so the title has room to breathe.
	maxQRWidth := size.X * 38 / 100
	qrSize := size.Y - padding*2
	if qrSize > maxQRWidth {
		qrSize = maxQRWidth
	}
	if qrSize < size.Y/3 {
		qrSize = size.Y / 3
	}

	qrRect := image.Rect(size.X-qrSize-padding, (size.Y-qrSize)/2, size.X-padding, (size.Y+qrSize)/2)
	drawScaledQR(canvas, qrImg, qrRect)

	textRect := image.Rect(padding, padding, qrRect.Min.X-gap, size.Y-padding)
	drawLabelText(canvas, title, textRect)

	return canvas, nil
}

func encodeQR(data string) ([]byte, error) {
	qrc, err := qrcode.New(data)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := standard.NewWithWriter(struct {
		io.Writer
		io.Closer
	}{Writer: &buf, Closer: io.NopCloser(nil)})
	if err := qrc.Save(w); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawScaledQR(dst *image.RGBA, src image.Image, rect image.Rectangle) {
	srcBounds := src.Bounds()
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			sx := srcBounds.Min.X + (x-rect.Min.X)*srcBounds.Dx()/rect.Dx()
			sy := srcBounds.Min.Y + (y-rect.Min.Y)*srcBounds.Dy()/rect.Dy()
			dst.Set(x, y, src.At(sx, sy))
		}
	}
}

func drawLabelText(dst *image.RGBA, title string, rect image.Rectangle) {
	fontData, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return
	}
	title = strings.TrimSpace(title)
	if title == "" {
		title = "Untitled"
	}

	maxLines := 3
	if rect.Dy() > rect.Dx() {
		maxLines = 4
	}

	fontSize, lines := fitLabelLines(fontData, title, rect, maxLines)
	if len(lines) == 0 {
		return
	}

	c := freetype.NewContext()
	c.SetDst(dst)
	c.SetClip(rect)
	c.SetSrc(image.Black)
	c.SetFont(fontData)
	c.SetFontSize(fontSize)

	lineHeight := int(fontSize * 1.25)
	totalHeight := lineHeight * len(lines)
	y := rect.Min.Y + (rect.Dy()-totalHeight)/2 + int(fontSize)
	for _, line := range lines {
		pt := freetype.Pt(rect.Min.X, y)
		_, _ = c.DrawString(line, pt)
		y += lineHeight
	}
}

func fitLabelLines(fontData *truetype.Font, title string, rect image.Rectangle, maxLines int) (float64, []string) {
	minSize := 12.0
	maxSize := float64(rect.Dy()) / float64(maxLines) * 1.1
	if maxSize > 56 {
		maxSize = 56
	}

	bestSize := minSize
	var bestLines []string

	for size := maxSize; size >= minSize; size -= 1 {
		face := truetype.NewFace(fontData, &truetype.Options{Size: size, DPI: 72})
		lines := wrapText(title, face, rect.Dx())
		if len(lines) > maxLines {
			continue
		}
		if lineTooWide(lines, face, rect.Dx()) {
			continue
		}
		bestSize = size
		bestLines = lines
		break
	}

	if len(bestLines) == 0 {
		face := truetype.NewFace(fontData, &truetype.Options{Size: minSize, DPI: 72})
		bestLines = wrapText(title, face, rect.Dx())
		if len(bestLines) > maxLines {
			bestLines = bestLines[:maxLines]
			bestLines[maxLines-1] = truncateWithEllipsis(bestLines[maxLines-1], face, rect.Dx())
		}
	}

	return bestSize, bestLines
}

func lineTooWide(lines []string, face font.Face, maxWidth int) bool {
	for _, line := range lines {
		if font.MeasureString(face, line).Ceil() > maxWidth {
			return true
		}
	}
	return false
}

func truncateWithEllipsis(text string, face font.Face, maxWidth int) string {
	ellipsis := "..."
	if font.MeasureString(face, text).Ceil() <= maxWidth {
		return text
	}
	for len(text) > 0 {
		text = text[:len(text)-1]
		candidate := strings.TrimRight(text, " ") + ellipsis
		if font.MeasureString(face, candidate).Ceil() <= maxWidth {
			return candidate
		}
	}
	return ellipsis
}

func wrapText(text string, face font.Face, maxWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{"Untitled"}
	}
	var lines []string
	current := words[0]
	for _, word := range words[1:] {
		candidate := current + " " + word
		if font.MeasureString(face, candidate).Ceil() <= maxWidth {
			current = candidate
			continue
		}
		lines = append(lines, current)
		current = word
	}
	if font.MeasureString(face, current).Ceil() > maxWidth {
		lines = append(lines, truncateWithEllipsis(current, face, maxWidth))
	} else {
		lines = append(lines, current)
	}
	return lines
}

func requestOrigin(e *core.RequestEvent) string {
	scheme := "http"
	if e.Request.TLS != nil || strings.EqualFold(e.Request.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := e.Request.Host
	if forwarded := e.Request.Header.Get("X-Forwarded-Host"); forwarded != "" {
		host = strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	return scheme + "://" + host
}
