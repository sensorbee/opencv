package opencv

import (
	"bytes"
	"fmt"
	"gopkg.in/sensorbee/opencv.v0/bridge"
	"gopkg.in/sensorbee/sensorbee.v0/data"
	"image"
	"image/jpeg"
)

var (
	imagePath = data.MustCompilePath("image")
)

// TypeImageFormat is an ID of image format type.
type TypeImageFormat int

const (
	typeUnknownFormat TypeImageFormat = iota
	// TypeCVMAT is OpenCV cv::Mat_<cv::Vec3b> format
	TypeCVMAT
	// TypeCVMAT4b is OpenCV cv::Mat_<cv::Vec4b> format
	TypeCVMAT4b
	// TypeJPEG is JPEG format
	TypeJPEG
)

func (t TypeImageFormat) String() string {
	switch t {
	case TypeCVMAT:
		return "cvmat"
	case TypeCVMAT4b:
		return "cvmat4b"
	case TypeJPEG:
		return "jpeg"
	default:
		return "unknown"
	}
}

// GetTypeImageFormat returns image format type.
func GetTypeImageFormat(str string) TypeImageFormat {
	switch str {
	case "cvmat":
		return TypeCVMAT
	case "cvmat4b":
		return TypeCVMAT4b
	case "jpeg":
		return TypeJPEG
	default:
		return typeUnknownFormat
	}
}

// RawData is represented of `cv::Mat_<cv::Vec3b>` structure.
type RawData struct {
	Format TypeImageFormat
	Width  int
	Height int
	Data   []byte
}

// ToRawData converts MatVec3b to RawData.
func ToRawData(m bridge.MatVec3b) RawData {
	w, h, data := m.ToRawData()
	return RawData{
		Format: TypeCVMAT,
		Width:  w,
		Height: h,
		Data:   data,
	}
}

// ToMatVec3b converts RawData to MatVec3b. Returned MatVec3b is required to
// delete after using.
func (r *RawData) ToMatVec3b() (bridge.MatVec3b, error) {
	if r.Format != TypeCVMAT {
		return bridge.MatVec3b{}, fmt.Errorf("'%v' cannot convert to 'MatVec3b'",
			r.Format)
	}
	return bridge.ToMatVec3b(r.Width, r.Height, r.Data), nil
}

func toRawMap(m *bridge.MatVec3b) data.Map {
	r := ToRawData(*m)
	return data.Map{
		"format": data.String(r.Format.String()), // = cv::Mat_<cv::Vec3b> = "cvmat"
		"width":  data.Int(r.Width),
		"height": data.Int(r.Height),
		"image":  data.Blob(r.Data),
	}
}

// ConvertMapToRawData returns RawData from data.Map. This function is
// utility method for other plug-in.
func ConvertMapToRawData(dm data.Map) (RawData, error) {
	var width int64
	if w, err := dm.Get(widthPath); err != nil {
		return RawData{}, err
	} else if width, err = data.ToInt(w); err != nil {
		return RawData{}, err
	}

	var height int64
	if h, err := dm.Get(heightPath); err != nil {
		return RawData{}, err
	} else if height, err = data.ToInt(h); err != nil {
		return RawData{}, err
	}

	var img []byte
	if b, err := dm.Get(imagePath); err != nil {
		return RawData{}, err
	} else if img, err = data.ToBlob(b); err != nil {
		return RawData{}, err
	}

	var format TypeImageFormat
	if f, err := dm.Get(formatPath); err != nil {
		return RawData{}, err
	} else if fmtStr, err := data.AsString(f); err != nil {
		return RawData{}, err
	} else {
		format = GetTypeImageFormat(fmtStr)
		if format == typeUnknownFormat {
			return RawData{}, fmt.Errorf("'%v' is not supported", fmtStr)
		}
	}

	return RawData{
		Format: format,
		Width:  int(width),
		Height: int(height),
		Data:   img,
	}, nil
}

// ConvertToDataMap returns data.map. This function is utility method for
// other plug-in.
func (r *RawData) ConvertToDataMap() data.Map {
	return data.Map{
		"format": data.String(r.Format.String()),
		"width":  data.Int(r.Width),
		"height": data.Int(r.Height),
		"image":  data.Blob(r.Data),
	}
}

// ToJpegData convert JPGE format image bytes.
func (r *RawData) ToJpegData(quality int) ([]byte, error) {
	if r.Format == TypeJPEG {
		return r.Data, nil
	}
	// BGR to RGB
	rgba := image.NewRGBA(image.Rect(0, 0, r.Width, r.Height))
	if r.Format == TypeCVMAT {
		for i, j := 0, 0; i < len(rgba.Pix); i, j = i+4, j+3 {
			rgba.Pix[i+0] = r.Data[j+2]
			rgba.Pix[i+1] = r.Data[j+1]
			rgba.Pix[i+2] = r.Data[j+0]
			rgba.Pix[i+3] = 0xFF
		}
	} else if r.Format == TypeCVMAT4b {
		for i, j := 0, 0; i < len(rgba.Pix); i, j = i+4, j+3 {
			rgba.Pix[i+0] = r.Data[j+2]
			rgba.Pix[i+1] = r.Data[j+1]
			rgba.Pix[i+2] = r.Data[j+0]
			rgba.Pix[i+3] = r.Data[j+3]
		}
	} else {
		return []byte{}, fmt.Errorf("'%v' cannot convert to JPEG", r.Format)
	}
	w := bytes.NewBuffer([]byte{})
	err := jpeg.Encode(w, rgba, &jpeg.Options{Quality: quality})
	return w.Bytes(), err
}
