package texture

import (
	"encoding/binary"
	"errors"
	"image/png"
	"io"
	"strconv"

	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

func ExtractDDS(ctx extractor.Context) error {
	if !ctx.File().Exists(stingray.DataMain) {
		return errors.New("no main data")
	}
	r, err := ctx.File().OpenMulti(ctx.Ctx(), stingray.DataMain, stingray.DataStream, stingray.DataGPU)
	if err != nil {
		return err
	}
	defer r.Close()

	if _, err := texture.DecodeInfo(r); err != nil {
		return err
	}

	info, err := dds.DecodeInfo(r)
	if err != nil {
		return err
	}
	if width, contains := ctx.Config()["width"]; contains {
		widthInt, err := strconv.Atoi(width)
		if err != nil {
			return err
		}
		if info.Header.Width != uint32(widthInt) {
			return nil
		}
	}
	if height, contains := ctx.Config()["height"]; contains {
		heightInt, err := strconv.Atoi(height)
		if err != nil {
			return err
		}
		if info.Header.Height != uint32(heightInt) {
			return nil
		}
	}
	out, err := ctx.CreateFile(".dds")
	if err != nil {
		return err
	}
	defer out.Close()
	if err := binary.Write(out, binary.LittleEndian, [4]uint8{'D', 'D', 'S', ' '}); err != nil {
		return err
	}
	if err := binary.Write(out, binary.LittleEndian, info.Header); err != nil {
		return err
	}
	if info.Header.PixelFormat.FourCC == [4]uint8{'D', 'X', '1', '0'} {
		if err := binary.Write(out, binary.LittleEndian, info.DXT10Header); err != nil {
			return err
		}
	}
	if _, err := io.Copy(out, r); err != nil {
		return err
	}
	return nil
}

func ConvertToPNG(ctx extractor.Context) error {
	origTex, err := texture.Decode(ctx.Ctx(), ctx.File(), false)
	if err != nil {
		return err
	}

	tex := origTex
	if len(origTex.Images) > 1 {
		tex = dds.StackLayers(origTex)
	}

	out, err := ctx.CreateFile(".png")
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, tex)
}
