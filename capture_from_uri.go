package opencv

import (
	"fmt"
	"gopkg.in/sensorbee/opencv.v0/bridge"
	"gopkg.in/sensorbee/sensorbee.v0/bql"
	"gopkg.in/sensorbee/sensorbee.v0/core"
	"gopkg.in/sensorbee/sensorbee.v0/data"
)

// FromURICreator is a creator of a capture from URI.
type FromURICreator struct{}

var (
	uriPath            = data.MustCompilePath("uri")
	formatPath         = data.MustCompilePath("format")
	frameSkipPath      = data.MustCompilePath("frame_skip")
	nextFrameErrorPath = data.MustCompilePath("next_frame_error")
	rewindPath         = data.MustCompilePath("rewind")
	rewindablePath     = data.MustCompilePath("rewindable")
)

// CreateSource creates a frame generator using OpenCV video capture.
// URI can be set HTTP address or file path.
//
// WITH parameters.
//
// uri: [required] A capture data's URI (e.g. /data/test.avi).
//
// format: Output format style, default is "cvmat".
//
// frame_skip: The number of frame skip, if set empty or "0" then read all
// frames. FPS is depended on the URI's file (or device).
//
// next_frame_error: When this source cannot read a new frame, occur error or
// not decided by the flag. If the flag set `true` then return error. Default
// value is true.
//
// rewindable: If set `true` then user can use `REWIND SOURCE` query.
func (c *FromURICreator) CreateSource(ctx *core.Context,
	ioParams *bql.IOParams, params data.Map) (core.Source, error) {

	cs, err := c.createCaptureFromURI(ctx, ioParams, params)
	if err != nil {
		return nil, err
	}

	rewindFlag := false
	if rf, err := params.Get(rewindablePath); err == nil {
		if rewindFlag, err = data.AsBool(rf); err != nil {
			return nil, err
		}
	} else if rf, err := params.Get(rewindPath); err == nil {
		ctx.Log().Warnln(`"rewind" is deprecated and not supported in later releases in favor of "rewindable"`)
		if rewindFlag, err = data.AsBool(rf); err != nil {
			return nil, err
		}
	}
	// Use Rewindable and ImplementSourceStop helpers that can enable this
	// source to stop thread-safe.
	if rewindFlag {
		return core.NewRewindableSource(cs), nil
	}
	return core.ImplementSourceStop(cs), nil
}

func (c *FromURICreator) createCaptureFromURI(ctx *core.Context,
	ioParams *bql.IOParams, params data.Map) (core.Source, error) {

	uri, err := params.Get(uriPath)
	if err != nil {
		return nil, fmt.Errorf("capture source needs URI")
	}
	uriStr, err := data.AsString(uri)
	if err != nil {
		return nil, err
	}

	format := "cvmat"
	if fm, err := params.Get(formatPath); err == nil {
		if format, err = data.AsString(fm); err != nil {
			return nil, err
		}
	}

	fs, err := params.Get(frameSkipPath)
	if err != nil {
		fs = data.Int(0) // will be ignored
	}
	frameSkip, err := data.AsInt(fs)
	if err != nil {
		return nil, err
	}

	endErrFlag, err := params.Get(nextFrameErrorPath)
	if err != nil {
		endErrFlag = data.True
	}
	endErr, err := data.AsBool(endErrFlag)
	if err != nil {
		return nil, err
	}

	cs := &captureFromURI{
		uri:        uriStr,
		frameSkip:  frameSkip,
		endErrFlag: endErr,
	}
	if format == "cvmat" {
		cs.foramtFunc = toRawMap
	} else {
		return nil, fmt.Errorf("'%v' format is not supported", format)
	}
	return cs, nil
}

type captureFromURI struct {
	uri        string
	frameSkip  int64
	endErrFlag bool
	foramtFunc func(m *bridge.MatVec3b) data.Map
}

// GenerateStream streams video capture data. OpenCV video capture read frames
// from URI, user can control frame streaming frequency using FrameSkip. This
// source is rewindable.
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
//
// When a capture source is a file-style (e.g. AVI file), tuples' timestamp is
// NOT correspond with the file created time. The timestamp value is the time
// of this source capturing a new frame.
// And when complete to read the file's all frames, video capture cannot read a
// new frame. If the key "next_frame_error" set `false` then a no new frame
// error will not be occurred, User can also count the number of total frame to
// confirm complete of read file. The number of frames is logged.
func (c *captureFromURI) GenerateStream(ctx *core.Context, w core.Writer) error {
	vcap := bridge.NewVideoCapture()
	defer vcap.Delete()
	if ok := vcap.Open(c.uri); !ok {
		return fmt.Errorf("error opening video stream or file: %v", c.uri)
	}

	buf := bridge.NewMatVec3b()
	defer buf.Delete()

	cnt := 0
	ctx.Log().Infof("start reading video stream of file: %v", c.uri)
	for {
		cnt++
		if ok := vcap.Read(buf); !ok {
			ctx.Log().Infof("total read frames count is %d", cnt-1)
			if c.endErrFlag {
				return fmt.Errorf("cannot reed a new frame")
			}
			break
		}
		if c.frameSkip > 0 {
			vcap.Grab(int(c.frameSkip))
		}

		m := c.foramtFunc(&buf)
		t := core.NewTuple(m)
		if err := w.Write(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (c *captureFromURI) Stop(ctx *core.Context) error {
	return nil
}
