package opencv

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/sensorbee/sensorbee.v0/bql"
	"gopkg.in/sensorbee/sensorbee.v0/core"
	"gopkg.in/sensorbee/sensorbee.v0/data"
	"testing"
)

func TestGenerateStreamURIError(t *testing.T) {
	ctx := &core.Context{}
	sc := FromURICreator{}
	ioParams := &bql.IOParams{}
	Convey("Given a CaptureFromURI source with invalid URI", t, func() {
		params := data.Map{
			"uri": data.String("error uri"),
		}
		capture, err := sc.CreateSource(ctx, ioParams, params)
		So(err, ShouldBeNil)
		So(capture, ShouldNotBeNil)
		Convey("When generate stream", func() {
			Convey("Then error has occurred", func() {
				err := capture.GenerateStream(ctx, &dummyWriter{})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldStartWith, "error")
			})
		})
	})
}

type dummyWriter struct{}

func (w *dummyWriter) Write(ctx *core.Context, t *core.Tuple) error {
	return nil
}

func TestGetURISourceCreatorWithRawMode(t *testing.T) {
	ctx := &core.Context{}
	ioParams := &bql.IOParams{}
	Convey("Given a raw mode enabled capture source creator", t, func() {
		sc := FromURICreator{}
		Convey("When create source with not supported format", func() {
			params := data.Map{
				"uri":    data.String("/data/file.avi"),
				"format": data.String("4k"),
			}
			_, err := sc.createCaptureFromURI(ctx, ioParams, params)
			Convey("Then capture should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestGetURISourceCreator(t *testing.T) {
	cc := &core.ContextConfig{}
	ctx := core.NewContext(cc)
	ioParams := &bql.IOParams{}
	Convey("Given a CaptureFromURI creator", t, func() {
		sc := FromURICreator{}
		Convey("When create source with full parameters", func() {
			params := data.Map{
				"uri":              data.String("/data/file.avi"),
				"format":           data.String("cvmat"),
				"frame_skip":       data.Int(5),
				"next_frame_error": data.False,
			}
			Convey("Then creator should initialize capture source", func() {
				s, err := sc.createCaptureFromURI(ctx, ioParams, params)
				So(err, ShouldBeNil)
				capture, ok := s.(*captureFromURI)
				So(ok, ShouldBeTrue)
				So(capture.uri, ShouldEqual, "/data/file.avi")
				So(capture.frameSkip, ShouldEqual, 5)
				So(capture.endErrFlag, ShouldBeFalse)
			})
		})

		Convey("When create source with empty uri", func() {
			params := data.Map{
				"frame_skip":       data.Int(5),
				"next_frame_error": data.False,
			}
			Convey("Then creator should occur an error", func() {
				s, err := sc.createCaptureFromURI(ctx, ioParams, params)
				So(err, ShouldNotBeNil)
				So(s, ShouldBeNil)
			})
		})

		Convey("When create source with only uri", func() {
			params := data.Map{
				"uri": data.String("/data/file.avi"),
			}
			Convey("Then capture should set default values", func() {
				s, err := sc.createCaptureFromURI(ctx, ioParams, params)
				So(err, ShouldBeNil)
				capture, ok := s.(*captureFromURI)
				So(ok, ShouldBeTrue)
				So(capture.uri, ShouldEqual, "/data/file.avi")
				So(capture.frameSkip, ShouldEqual, 0)
				So(capture.endErrFlag, ShouldBeTrue)
			})
		})

		Convey("When create source with invalid uri", func() {
			params := data.Map{
				"uri": data.Null{},
			}
			Convey("Then create should occur an error", func() {
				s, err := sc.createCaptureFromURI(ctx, ioParams, params)
				So(err, ShouldNotBeNil)
				So(s, ShouldBeNil)
			})
		})

		Convey("When create source with invalid option parameters", func() {
			params := data.Map{
				"uri": data.String("/data/file.avi"),
			}
			testMap := data.Map{
				"format":           data.True,
				"frame_skip":       data.String("@"),
				"next_frame_error": data.String("True"),
			}
			for k, v := range testMap {
				v := v
				msg := fmt.Sprintf("with %v error", k)
				Convey("Then creator should occur a parse error on option parameters"+msg, func() {
					params[k] = v
					s, err := sc.createCaptureFromURI(ctx, ioParams, params)
					So(err, ShouldNotBeNil)
					So(s, ShouldBeNil)
				})
			}
		})

		Convey("When create source with only uri and rewindable", func() {
			params := data.Map{
				"uri":        data.String("/data/file.avi"),
				"rewindable": data.True,
			}
			Convey("Then rewindable capture should be created", func() {
				c := FromURICreator{}
				s, err := c.CreateSource(ctx, ioParams, params)
				So(err, ShouldBeNil)
				So(s, ShouldNotBeNil)
				_, ok := s.(core.RewindableSource)
				So(ok, ShouldBeTrue)
			})
		})

		Convey("When create source with only uri and deprecated rewind param name", func() {
			params := data.Map{
				"uri":    data.String("/data/file.avi"),
				"rewind": data.True,
			}
			Convey("Then rewindable capture should be created", func() {
				c := FromURICreator{}
				s, err := c.CreateSource(ctx, ioParams, params)
				So(err, ShouldBeNil)
				So(s, ShouldNotBeNil)
				_, ok := s.(core.RewindableSource)
				So(ok, ShouldBeTrue)
			})
		})
	})
}
