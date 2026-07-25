package captcha

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/lyimoexiao/akari/internal/cache"
	"github.com/lyimoexiao/akari/internal/config"
	"github.com/lyimoexiao/akari/internal/logger"
	"github.com/wenlng/go-captcha-assets/resources/fonts/fzshengsksjw"
	captchaimages "github.com/wenlng/go-captcha-assets/resources/images"
	"github.com/wenlng/go-captcha-assets/resources/thumbs"
	"github.com/wenlng/go-captcha-assets/resources/tiles"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/click"
	"github.com/wenlng/go-captcha/v2/rotate"
	"github.com/wenlng/go-captcha/v2/slide"
)

// Service manages captcha generation and verification.
type Service struct {
	cfg    *config.CaptchaConfig
	cache  cache.Cache
	client *http.Client
	click  click.Captcha
	rotate rotate.Captcha
	slide  slide.Captcha
}

// New creates a new captcha Service.
func New(cfg *config.CaptchaConfig, c cache.Cache) *Service {
	s := &Service{
		cfg:    cfg,
		cache:  c,
		client: &http.Client{Timeout: 10 * time.Second},
	}
	if cfg.Provider == "gocaptcha" {
		s.init()
	}
	return s
}

// init initialises the captcha engines based on the configured type.
func (s *Service) init() {
	switch s.cfg.Type {
	case "rotate":
		s.rotate = s.newRotateCaptcha()
	case "slide":
		s.slide = s.newSlideCaptcha()
	default:
		// default to click
		s.click = s.newClickCaptcha()
	}
}

// newClickCaptcha creates a click-style captcha builder.
func (s *Service) newClickCaptcha() click.Captcha {
	font, err := fzshengsksjw.GetFont()
	if err != nil {
		panic(fmt.Errorf("load captcha font: %w", err))
	}

	bgImages, err := captchaimages.GetImages()
	if err != nil {
		panic(fmt.Errorf("load captcha images: %w", err))
	}

	thumbImages, err := thumbs.GetThumbs()
	if err != nil {
		panic(fmt.Errorf("load captcha thumbs: %w", err))
	}

	builder := click.NewBuilder(
		click.WithImageSize(option.Size{Width: 300, Height: 240}),
		click.WithRangeLen(option.RangeVal{Min: 4, Max: 6}),
		click.WithRangeVerifyLen(option.RangeVal{Min: 2, Max: 4}),
		click.WithRangeColors([]string{
			"#1e3a8a", "#166534", "#881337", "#1e40af", "#0f766e",
			"#a21caf", "#b91c1c", "#0d9488", "#7c3aed", "#2563eb",
		}),
		click.WithRangeAnglePos([]option.RangeVal{
			{Min: 20, Max: 330},
		}),
		click.WithRangeSize(option.RangeVal{Min: 32, Max: 48}),
		click.WithDisplayShadow(false),
		click.WithImageAlpha(1.0),
	)

	builder.SetResources(
		click.WithChars([]string{
			"这", "是", "一", "个", "测", "试", "验", "证", "码",
			"安", "全", "校", "验", "通", "过", "点", "击", "确", "认",
		}),
		click.WithFonts([]*truetype.Font{font}),
		click.WithBackgrounds(bgImages),
		click.WithThumbBackgrounds(thumbImages),
	)

	return builder.Make()
}

// newRotateCaptcha creates a rotate-style captcha builder.
func (s *Service) newRotateCaptcha() rotate.Captcha {
	builder := rotate.NewBuilder(
		rotate.WithImageSquareSize(280),
		rotate.WithRangeAnglePos([]option.RangeVal{
			{Min: 20, Max: 330},
		}),
		rotate.WithThumbImageAlpha(0.8),
	)

	bgImages, err := captchaimages.GetImages()
	if err != nil {
		panic(fmt.Errorf("load captcha images: %w", err))
	}

	builder.SetResources(
		rotate.WithImages(bgImages),
	)

	return builder.Make()
}

// newSlideCaptcha creates a slide-style captcha builder.
func (s *Service) newSlideCaptcha() slide.Captcha {
	builder := slide.NewBuilder(
		slide.WithImageSize(option.Size{Width: 300, Height: 240}),
		slide.WithRangeGraphSize(option.RangeVal{Min: 40, Max: 60}),
		slide.WithRangeGraphAnglePos([]option.RangeVal{
			{Min: 20, Max: 330},
		}),
		slide.WithImageAlpha(1.0),
	)

	bgImages, err := captchaimages.GetImages()
	if err != nil {
		panic(fmt.Errorf("load captcha images: %w", err))
	}

	graphTiles, err := tiles.GetTiles()
	if err != nil {
		panic(fmt.Errorf("load captcha tiles: %w", err))
	}

	slideTiles := make([]*slide.GraphImage, 0, len(graphTiles))
	for _, t := range graphTiles {
		slideTiles = append(slideTiles, &slide.GraphImage{
			OverlayImage: t.OverlayImage,
			ShadowImage:  t.ShadowImage,
			MaskImage:    t.MaskImage,
		})
	}

	builder.SetResources(
		slide.WithBackgrounds(bgImages),
		slide.WithGraphImages(slideTiles),
	)

	return builder.Make()
}

// Generate creates a new captcha and returns the image data along with
// a unique captcha ID that can be used for verification.
// Returns nil if captcha is disabled.
func (s *Service) Generate(ctx context.Context) (map[string]any, error) {
	if !s.cfg.Enabled {
		return nil, nil
	}

	switch s.cfg.Provider {
	case "turnstile":
		return map[string]any{
			"provider":  "turnstile",
			"site_key":  s.cfg.Turnstile.SiteKey,
		}, nil
	default:
		return s.generateGoCaptcha(ctx)
	}
}

func (s *Service) generateGoCaptcha(ctx context.Context) (map[string]any, error) {
	switch s.cfg.Type {
	case "rotate":
		return s.genRotate(ctx)
	case "slide":
		return s.genSlide(ctx)
	default:
		return s.genClick(ctx)
	}
}

// genClick generates a click captcha.
func (s *Service) genClick(ctx context.Context) (map[string]any, error) {
	data, err := s.click.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate click captcha: %w", err)
	}

	dots := data.GetData()
	masterImage := data.GetMasterImage()
	thumbImage := data.GetThumbImage()

	// Encode images to base64
	masterB64, err := masterImage.ToBase64()
	if err != nil {
		return nil, fmt.Errorf("encode master image: %w", err)
	}
	thumbB64, err := thumbImage.ToBase64()
	if err != nil {
		return nil, fmt.Errorf("encode thumb image: %w", err)
	}

	// Collect dot info for the frontend (without text, for security)
	// Note: dots are NOT returned to frontend; stored server-side for verification
	captchaID := generateID()
	storedDots := make([]map[string]any, 0, len(dots))
	for _, dot := range dots {
		storedDots = append(storedDots, map[string]any{
			"index":  dot.Index,
			"x":      dot.X,
			"y":      dot.Y,
			"width":  dot.Width,
			"height": dot.Height,
			"text":   dot.Text,
		})
	}
	if err := s.cacheAnswer(ctx, captchaID, map[string]any{"dots": storedDots}); err != nil {
		return nil, err
	}

	logger.L.Debugw("genClick captcha created",
		"captcha_id", captchaID,
		"type", "click",
		"dot_count", len(storedDots),
		"dots", storedDots)

	return map[string]any{
		"captcha_id":   captchaID,
		"type":         "click",
		"master_image": masterB64,
		"thumb_image":  thumbB64,
	}, nil
}

// genRotate generates a rotate captcha.
func (s *Service) genRotate(ctx context.Context) (map[string]any, error) {
	data, err := s.rotate.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate rotate captcha: %w", err)
	}

	block := data.GetData()
	masterImage := data.GetMasterImage()
	thumbImage := data.GetThumbImage()

	masterB64, err := masterImage.ToBase64()
	if err != nil {
		return nil, fmt.Errorf("encode master image: %w", err)
	}
	thumbB64, err := thumbImage.ToBase64()
	if err != nil {
		return nil, fmt.Errorf("encode thumb image: %w", err)
	}

	// Store the correct angle in cache
	captchaID := generateID()
	answer := map[string]any{"angle": float64(block.Angle)}
	if err := s.cacheAnswer(ctx, captchaID, answer); err != nil {
		return nil, err
	}

	return map[string]any{
		"captcha_id":   captchaID,
		"type":         "rotate",
		"master_image": masterB64,
		"thumb_image":  thumbB64,
		"angle":        block.Angle,
	}, nil
}

// genSlide generates a slide captcha.
func (s *Service) genSlide(ctx context.Context) (map[string]any, error) {
	data, err := s.slide.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate slide captcha: %w", err)
	}

	block := data.GetData()
	masterImage := data.GetMasterImage()
	tileImage := data.GetTileImage()

	masterB64, err := masterImage.ToBase64()
	if err != nil {
		return nil, fmt.Errorf("encode master image: %w", err)
	}
	tileB64, err := tileImage.ToBase64()
	if err != nil {
		return nil, fmt.Errorf("encode tile image: %w", err)
	}

	// Store the correct X position in cache
	captchaID := generateID()
	answer := map[string]any{"x": float64(block.X), "y": float64(block.Y)}
	if err := s.cacheAnswer(ctx, captchaID, answer); err != nil {
		return nil, err
	}

	return map[string]any{
		"captcha_id":   captchaID,
		"type":         "slide",
		"master_image": masterB64,
		"tile_image":   tileB64,
		"tile_width":   block.Width,
		"tile_height":  block.Height,
		"thumb_x":      block.X,
		"thumb_y":      block.Y,
	}, nil
}

// Verify checks if the user-provided answer matches the stored answer.
// Returns true if captcha is disabled (no verification needed).
func (s *Service) Verify(ctx context.Context, captchaID string, userAnswer map[string]any) (bool, error) {
	if !s.cfg.Enabled {
		return true, nil
	}

	switch s.cfg.Provider {
	case "turnstile":
		return true, nil // handled by VerifyToken
	default:
		return s.verifyGoCaptcha(ctx, captchaID, userAnswer)
	}
}

// VerifyToken verifies a Turnstile token against the Cloudflare API.
func (s *Service) VerifyToken(ctx context.Context, token string) (bool, error) {
	body, _ := json.Marshal(map[string]string{
		"secret":   s.cfg.Turnstile.SecretKey,
		"response": token,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://challenges.cloudflare.com/turnstile/v0/siteverify", bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("create turnstile request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("call turnstile API: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return false, fmt.Errorf("parse turnstile response: %w", err)
	}
	return result.Success, nil
}

func (s *Service) verifyGoCaptcha(ctx context.Context, captchaID string, userAnswer map[string]any) (bool, error) {
	if captchaID == "" {
		return false, fmt.Errorf("captcha_id is required")
	}

	// Retrieve stored answer from cache
	key := s.cfg.CachePrefix + captchaID
	var storedAnswer map[string]any
	if err := s.cache.Get(ctx, key, &storedAnswer); err != nil {
		return false, fmt.Errorf("get captcha answer from cache: %w", err)
	}
	if storedAnswer == nil {
		return false, fmt.Errorf("captcha expired or not found")
	}

	// Delete the cached answer regardless of result (one-time use)
	_ = s.cache.Del(ctx, key)

	switch s.cfg.Type {
	case "rotate":
		return s.verifyRotate(storedAnswer, userAnswer)
	case "slide":
		return s.verifySlide(storedAnswer, userAnswer)
	default:
		return s.verifyClick(storedAnswer, userAnswer)
	}
}

// verifyClick verifies a click captcha answer by matching click positions to stored dots.
func (s *Service) verifyClick(storedData, userData map[string]any) (bool, error) {
	// storedData is {dots: [{index, x, y, text}, ...]}
	storedDotsRaw, ok := storedData["dots"]
	if !ok {
		return false, nil
	}
	storedDots, ok := storedDotsRaw.([]any)
	if !ok {
		return false, nil
	}

	userDotsRaw, ok := userData["dots"]
	if !ok {
		return false, nil
	}
	userDots, ok := userDotsRaw.([]any)
	if !ok {
		return false, nil
	}

	if len(userDots) != len(storedDots) {
		logger.L.Debugw("verifyClick dot count mismatch",
			"user_dots", len(userDots),
			"stored_dots", len(storedDots))
		return false, nil
	}

	// For each user click, find the nearest stored dot
	for clickIdx, userClick := range userDots {
		uc, ok := userClick.(map[string]any)
		if !ok {
			return false, nil
		}

		clickX, _ := uc["x"].(float64)
		clickY, _ := uc["y"].(float64)

		// Find the nearest stored dot that hasn't been matched yet
		bestDist := -1.0
		bestIdx := -1

		for j, sd := range storedDots {
			dot, ok := sd.(map[string]any)
			if !ok {
				continue
			}

			// Skip already matched dots
			if matched, _ := dot["_matched"].(bool); matched {
				continue
			}

			dotX, _ := dot["x"].(float64)
			dotY, _ := dot["y"].(float64)
			dotW, _ := dot["width"].(float64)
			dotH, _ := dot["height"].(float64)

			if dotW == 0 {
				dotW = 40
			}
			if dotH == 0 {
				dotH = 40
			}

			dx := clickX - dotX
			dy := clickY - dotY
			dist := dx*dx + dy*dy

			// Check if click is within roughly 1 dot size distance from center
			tolerance := dotW*dotW + dotH*dotH
			if dist <= tolerance {
				if bestDist < 0 || dist < bestDist {
					bestDist = dist
					bestIdx = j
				}
			}
		}

		if bestIdx < 0 {
			logger.L.Debugw("verifyClick no matching dot found",
				"click_idx", clickIdx,
				"click_x", clickX,
				"click_y", clickY)
			return false, nil
		}

		// Mark as matched
		sd := storedDots[bestIdx].(map[string]any)
		sd["_matched"] = true
		logger.L.Debugw("verifyClick dot matched",
			"click_idx", clickIdx,
			"click_x", clickX,
			"click_y", clickY,
			"matched_idx", bestIdx,
			"matched_x", sd["x"],
			"matched_y", sd["y"],
			"distance", bestDist)
	}

	// All dots matched successfully
	return true, nil
}

// verifyRotate verifies a rotate captcha answer.
func (s *Service) verifyRotate(storedData, userData map[string]any) (bool, error) {
	storedAngle, ok := storedData["angle"].(float64)
	if !ok {
		return false, nil
	}
	userAngle, ok := userData["angle"].(float64)
	if !ok {
		return false, nil
	}
	// Allow a small tolerance (padding) of 5 degrees
	diff := userAngle - storedAngle
	if diff < 0 {
		diff = -diff
	}
	return diff <= 5, nil
}

// verifySlide verifies a slide captcha answer.
func (s *Service) verifySlide(storedData, userData map[string]any) (bool, error) {
	storedX, ok := storedData["x"].(float64)
	if !ok {
		return false, nil
	}
	userX, ok := userData["x"].(float64)
	if !ok {
		return false, nil
	}
	// Allow a small tolerance (padding) of 5 pixels
	diff := userX - storedX
	if diff < 0 {
		diff = -diff
	}
	return diff <= 5, nil
}

// cacheAnswer stores the captcha answer in cache with a TTL of 5 minutes.
func (s *Service) cacheAnswer(ctx context.Context, captchaID string, answer any) error {
	key := s.cfg.CachePrefix + captchaID
	return s.cache.Set(ctx, key, answer, 5*time.Minute)
}

// IsEnabled returns whether the captcha service is enabled.
func (s *Service) IsEnabled() bool {
	return s.cfg.Enabled
}

// generateID creates a random captcha ID using crypto/rand.
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// fallback: timestamp + hex — still unique even if crypto rand fails
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(b)
}
