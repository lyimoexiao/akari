package skinlib

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
)

// validateAndSanitizePNG validates texture dimensions and re-encodes to strip metadata.
// Returns the sanitized PNG data as a byte slice.
//
// Security: dimensions are checked via DecodeConfig (header-only, no pixel decompression)
// BEFORE reading the full file, preventing PNG bomb attacks.
func validateAndSanitizePNG(reader io.Reader) ([]byte, error) {
	// 1. Tee the input so we can re-read the header bytes after DecodeConfig consumes them.
	var headerBuf bytes.Buffer
	tee := io.TeeReader(reader, &headerBuf)

	// 2. Check dimensions via header-only decode (safe — no pixel data decompressed).
	cfg, err := png.DecodeConfig(tee)
	if err != nil {
		return nil, fmt.Errorf("无效的 PNG 文件: %w", err)
	}
	w, h := cfg.Width, cfg.Height

	// 3. Validate dimensions
	// Skin: 64x32 or 64x64 multiples
	// Cape: 64x32 multiples or 22x17 multiples
	isSkin := (w%64 == 0 && (h%32 == 0 || h%64 == 0))
	isCape := (w%64 == 0 && h%32 == 0) || (w%22 == 0 && h%17 == 0)

	if !isSkin && !isCape {
		return nil, fmt.Errorf("不支持的纹理尺寸: %dx%d，皮肤需要 64x32/64x64 的倍数，披风需要 64x32 或 22x17 的倍数", w, h)
	}

	// 4. Read remaining data (tee already captured the header into headerBuf,
	//    and reader has the rest after the header).
	rest, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("读取 PNG 数据失败: %w", err)
	}
	fullData := append(headerBuf.Bytes(), rest...)

	// 5. Decode the full image
	img, err := png.Decode(bytes.NewReader(fullData))
	if err != nil {
		return nil, fmt.Errorf("解码 PNG 失败: %w", err)
	}

	// 5. Pad non-standard capes (22x17 multiples → 64x32)
	needsPad := (w%22 == 0 && h%17 == 0) && !(w%64 == 0 && h%32 == 0)
	if needsPad {
		// Calculate the padded dimensions (next multiple of 64x32)
		padW := ((w + 63) / 64) * 64
		padH := ((h + 31) / 32) * 32
		if padW < 64 {
			padW = 64
		}
		if padH < 32 {
			padH = 32
		}

		dst := image.NewNRGBA(image.Rect(0, 0, padW, padH))
		draw.Draw(dst, dst.Bounds(), image.Transparent, image.Point{}, draw.Src)
		draw.Draw(dst, image.Rect(0, 0, w, h), img, image.Point{}, draw.Src)
		img = dst
	}

	// 6. Re-encode to strip metadata
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("重新编码 PNG 失败: %w", err)
	}

	return buf.Bytes(), nil
}
