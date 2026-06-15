package custom

import (
	"bytes"
	"image/png"
	"io"
	"net/url"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"

	_ "embed"
)

//go:embed assets/QRIcon.png
var qrcodeLogo []byte

func registerQRCodeRoute(g *router.RouterGroup[*core.RequestEvent]) {
	g.GET("/qrcode", func(e *core.RequestEvent) error {
		data := e.Request.URL.Query().Get("data")
		if data == "" {
			return e.BadRequestError("missing data query param", nil)
		}

		image, err := png.Decode(bytes.NewReader(qrcodeLogo))
		if err != nil {
			return e.InternalServerError("failed to load qrcode logo", err)
		}

		decodedStr, err := url.QueryUnescape(data)
		if err != nil {
			return e.BadRequestError("invalid data", err)
		}

		qrc, err := qrcode.New(decodedStr)
		if err != nil {
			return e.BadRequestError("invalid qrcode data", err)
		}

		toWriteCloser := struct {
			io.Writer
			io.Closer
		}{
			Writer: e.Response,
			Closer: io.NopCloser(nil),
		}

		qrwriter := standard.NewWithWriter(toWriteCloser, standard.WithLogoImage(image))
		e.Response.Header().Set("Content-Type", "image/jpeg")
		e.Response.Header().Set("Content-Disposition", "attachment; filename=qrcode.jpg")
		return qrc.Save(qrwriter)
	})
}
