package opencv

import (
	"fmt"
	"gopkg.in/sensorbee/opencv.v0/bridge"
	"gopkg.in/sensorbee/sensorbee.v0/bql"
	"gopkg.in/sensorbee/sensorbee.v0/core"
	"gopkg.in/sensorbee/sensorbee.v0/data"
)

// FromDeviceCreator is a creator of a capture from device.
type FromDeviceCreator struct{}

var (
	deviceIDPath = data.MustCompilePath("device_id")
	widthPath    = data.MustCompilePath("width")
	heightPath   = data.MustCompilePath("height")
	fpsPath      = data.MustCompilePath("fps")
)

// CreateSource creates a frame generator using OpenCV video capture
// (`VideoCapture::open`).
//
// WITH parameters.
//
// device_id: [required] The ID of associated device.
//
// format: Output format style, default is "cvmat".
//
// width: Frame width, if set empty or "0" then will be ignore.
//
// height: Frame height, if set empty or "0" then will be ignore.
//
// fps: Frame per second, if set empty or "0" then will be ignore.
func (c *FromDeviceCreator) CreateSource(ctx *core.Context, ioParams *bql.IOParams,
	params data.Map) (core.Source, error) {
	cs, err := c.createCaptureFromDevice(ctx, ioParams, params)
	if err != nil {
		return nil, err
	}

	// Use ImplementSourceStop helper that can enable this source to stop
	// thread-safe.
	return core.ImplementSourceStop(cs), nil
}

func (c *FromDeviceCreator) createCaptureFromDevice(ctx *core.Context,
	ioParams *bql.IOParams, params data.Map) (core.Source, error) {
	did, err := params.Get(deviceIDPath)
	if err != nil {
		return nil, err
	}
	deviceID, err := data.AsInt(did)
	if err != nil {
		return nil, err
	}

	format := "cvmat"
	if fm, err := params.Get(formatPath); err == nil {
		if format, err = data.AsString(fm); err != nil {
			return nil, err
		}
	}

	w, err := params.Get(widthPath)
	if err != nil {
		w = data.Int(0) // will be ignored
	}
	width, err := data.AsInt(w)
	if err != nil {
		return nil, err
	}

	h, err := params.Get(heightPath)
	if err != nil {
		h = data.Int(0) // will be ignored
	}
	height, err := data.AsInt(h)
	if err != nil {
		return nil, err
	}

	f, err := params.Get(fpsPath)
	if err != nil {
		f = data.Int(0) // will be ignored
	}
	fps, err := data.AsInt(f)
	if err != nil {
		return nil, err
	}

	cs := &captureFromDevice{
		deviceID: deviceID,
		width:    width,
		height:   height,
		fps:      fps,
	}
	if format == "cvmat" {
		cs.formatFunc = toRawMap
	} else {
		return nil, fmt.Errorf("'%v' format is not supported", format)
	}
	return cs, nil
}

type captureFromDevice struct {
	deviceID   int64
	width      int64
	height     int64
	fps        int64
	formatFunc func(m *bridge.MatVec3b) data.Map
}

// GenerateStream streams video capture data. OpenCV parameters
// (e.g width, height...) are set when the source is initialized.
//
// Output
//
// format: The frame's format style, ex) "cvmat", "jpeg",...
//
// mode: The frame's format mode, ex) "BGR", "RGBA",...
//
// width: The frame's width.
//
// height: The frame's height.
//
// image: The binary data of frame image.
func (c *captureFromDevice) GenerateStream(ctx *core.Context, w core.Writer) error {
	vcap := bridge.NewVideoCapture()
	defer vcap.Delete()

	if ok := vcap.OpenDevice(int(c.deviceID)); !ok {
		return fmt.Errorf("error opening device: %v", c.deviceID)
	}

	// OpenCV video capture configuration
	if c.width > 0 {
		vcap.Set(bridge.CvCapPropFrameWidth, int(c.width))
	}
	if c.height > 0 {
		vcap.Set(bridge.CvCapPropFrameHeight, int(c.height))
	}
	if c.fps > 0 {
		vcap.Set(bridge.CvCapPropFps, int(c.fps))
	}

	// streaming, capture from vcap
	buf := bridge.NewMatVec3b()
	defer buf.Delete()
	ctx.Log().Infof("start reading camera device: %v", c.deviceID)
	for {
		if ok := vcap.Read(buf); !ok {
			return fmt.Errorf("cannot read a new file (device no: %d)", c.deviceID)
		}
		if buf.Empty() {
			continue
		}

		m := c.formatFunc(&buf)
		t := core.NewTuple(m)
		if err := w.Write(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (c *captureFromDevice) Stop(ctx *core.Context) error {
	return nil
}
