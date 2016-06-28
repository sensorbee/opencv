package opencv

import (
	"fmt"
	"gopkg.in/sensorbee/opencv.v0/bridge"
	"gopkg.in/sensorbee/sensorbee.v0/core"
	"gopkg.in/sensorbee/sensorbee.v0/data"
)

var (
	configFilePath = data.MustCompilePath("file")
	xPath          = data.MustCompilePath("x")
	yPath          = data.MustCompilePath("y")
)

// NewCascadeClassifier returns cascadeClassifier state.
//
// file: cascade configuration file path for detection.
// e.g. "haarcascade_frontalface_default.xml".
func NewCascadeClassifier(ctx *core.Context, params data.Map) (core.SharedState,
	error) {
	var filePath string
	if fp, err := params.Get(configFilePath); err != nil {
		return nil, err
	} else if filePath, err = data.AsString(fp); err != nil {
		return nil, err
	}

	cc := bridge.NewCascadeClassifier()
	if !cc.Load(filePath) {
		return nil, fmt.Errorf("cannot load the file '%v'", filePath)
	}

	return &cascadeClassifier{
		classifier: cc,
	}, nil
}

type cascadeClassifier struct {
	classifier bridge.CascadeClassifier
}

func (c *cascadeClassifier) Terminate(ctx *core.Context) error {
	c.classifier.Delete()
	return nil
}

func lookupCascadeClassifier(ctx *core.Context, name string) (*cascadeClassifier,
	error) {
	st, err := ctx.SharedStates.Get(name)
	if err != nil {
		return nil, err
	}

	if s, ok := st.(*cascadeClassifier); ok {
		return s, nil
	}
	return nil, fmt.Errorf("state '%v' cannot be converted to cascade_classifier.state",
		name)
}

// DetectMultiScale classifies and detect image.
//
// classifierName: cascadeClassifier state name.
//
// img: target image as RawData map structure.
func DetectMultiScale(ctx *core.Context, classifierName string, img data.Map) (
	data.Array, error) {
	raw, err := ConvertMapToRawData(img)
	if err != nil {
		return nil, err
	}
	mat, err := raw.ToMatVec3b()
	if err != nil {
		return nil, err
	}
	defer mat.Delete()

	classifier, err := lookupCascadeClassifier(ctx, classifierName)
	if err != nil {
		return nil, err
	}
	rects := classifier.classifier.DetectMultiScale(mat)
	ret := make(data.Array, len(rects))
	for i, r := range rects {
		rect := data.Map{
			"x":      data.Int(r.X),
			"y":      data.Int(r.Y),
			"width":  data.Int(r.Width),
			"height": data.Int(r.Height),
		}
		ret[i] = rect
	}
	return ret, nil
}

// DrawRectsToImage draws rectangle information on target image. The image is
// required to structured as RawData.
func DrawRectsToImage(img data.Map, rects data.Array) (data.Map, error) {
	if len(rects) == 0 {
		return img, nil
	}
	raw, err := ConvertMapToRawData(img)
	if err != nil {
		return nil, err
	}
	// copy image binary
	temp := make([]byte, len(raw.Data))
	copy(temp, raw.Data)
	raw.Data = temp
	mat, err := raw.ToMatVec3b()
	if err != nil {
		return nil, err
	}
	defer mat.Delete()

	brRects, err := convertToBridgeRects(rects)
	if err != nil {
		return nil, err
	}

	bridge.DrawRectsToImage(mat, brRects)
	retRaw := ToRawData(mat)
	return retRaw.ConvertToDataMap(), nil
}

func convertToBridgeRects(rects data.Array) ([]bridge.Rect, error) {
	brRects := make([]bridge.Rect, len(rects))
	for i, r := range rects {
		rmap, err := data.AsMap(r)
		if err != nil {
			return nil, err
		}
		var x int64
		if xv, err := rmap.Get(xPath); err != nil {
			return nil, err
		} else if x, err = data.ToInt(xv); err != nil {
			return nil, err
		}
		var y int64
		if yv, err := rmap.Get(yPath); err != nil {
			return nil, err
		} else if y, err = data.ToInt(yv); err != nil {
			return nil, err
		}
		var width int64
		if wv, err := rmap.Get(widthPath); err != nil {
			return nil, err
		} else if width, err = data.ToInt(wv); err != nil {
			return nil, err
		}
		var height int64
		if hv, err := rmap.Get(heightPath); err != nil {
			return nil, err
		} else if height, err = data.ToInt(hv); err != nil {
			return nil, err
		}
		rect := bridge.Rect{
			X:      int(x),
			Y:      int(y),
			Width:  int(width),
			Height: int(height),
		}
		brRects[i] = rect
	}
	return brRects, nil
}

// NewSharedImage returns shared image file to reduce I/O cost.
func NewSharedImage(ctx *core.Context, params data.Map) (core.SharedState, error) {
	var filePath string
	if fp, err := params.Get(configFilePath); err != nil {
		return nil, err
	} else if filePath, err = data.AsString(fp); err != nil {
		return nil, err
	}
	mat := bridge.LoadAlphaImage(filePath)
	return &sharedImage{
		img: mat,
	}, nil
}

type sharedImage struct {
	img bridge.MatVec4b
}

func (s *sharedImage) Terminate(ctx *core.Context) error {
	s.img.Delete()
	return nil
}

func lookupSharedImage(ctx *core.Context, name string) (*sharedImage, error) {
	st, err := ctx.SharedStates.Get(name)
	if err != nil {
		return nil, err
	}

	if s, ok := st.(*sharedImage); ok {
		return s, nil
	}
	return nil, fmt.Errorf("state '%v' cannot be converted to shared_image.state",
		name)
}

// MountAlphaImage draw target image on back image.
func MountAlphaImage(ctx *core.Context, imgName string, back data.Map,
	rects data.Array) (data.Map, error) {
	if len(rects) == 0 {
		return back, nil
	}
	img, err := lookupSharedImage(ctx, imgName)
	if err != nil {
		return nil, err
	}

	raw, err := ConvertMapToRawData(back)
	if err != nil {
		return nil, err
	}
	// copy image binary
	temp := make([]byte, len(raw.Data))
	copy(temp, raw.Data)
	raw.Data = temp
	mat, err := raw.ToMatVec3b()
	if err != nil {
		return nil, err
	}
	defer mat.Delete()

	brRects, err := convertToBridgeRects(rects)
	if err != nil {
		return nil, err
	}

	bridge.MountAlphaImage(img.img, mat, brRects)
	retRaw := ToRawData(mat)
	return retRaw.ConvertToDataMap(), nil
}
