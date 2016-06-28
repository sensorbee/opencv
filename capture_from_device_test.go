package opencv

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/sensorbee/sensorbee.v0/bql"
	"gopkg.in/sensorbee/sensorbee.v0/core"
	"gopkg.in/sensorbee/sensorbee.v0/data"
	"testing"
)

func TestGenerateStreamDeviceError(t *testing.T) {
	ctx := &core.Context{}
	sc := FromDeviceCreator{}
	ioParams := &bql.IOParams{}
	Convey("Given a CaptureFromDevice source with invalid device ID", t, func() {
		params := data.Map{
			"device_id": data.Int(999999), // invalid device ID
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

func TestGetDeviceSourceCreator(t *testing.T) {
	ctx := &core.Context{}
	ioParams := &bql.IOParams{}
	Convey("Given a CaptureFromDevice creator", t, func() {
		sc := FromDeviceCreator{}
		Convey("When create source with full parameters", func() {
			params := data.Map{
				"device_id": data.Int(0),
				"format":    data.String("cvmat"),
				"width":     data.Int(500),
				"height":    data.Int(600),
				"fps":       data.Int(25),
			}
			Convey("Then creator should initialize capture source", func() {
				s, err := sc.createCaptureFromDevice(ctx, ioParams, params)
				So(err, ShouldBeNil)
				capture, ok := s.(*captureFromDevice)
				So(ok, ShouldBeTrue)
				So(capture.deviceID, ShouldEqual, 0)
				So(capture.width, ShouldEqual, 500)
				So(capture.height, ShouldEqual, 600)
				So(capture.fps, ShouldEqual, 25)
			})
		})

		Convey("When create source with empty device ID", func() {
			params := data.Map{
				"width":  data.Int(500),
				"height": data.Int(600),
				"fps":    data.Int(25),
			}
			Convey("Then creator should occur an error", func() {
				s, err := sc.CreateSource(ctx, ioParams, params)
				So(err, ShouldNotBeNil)
				So(s, ShouldBeNil)
			})
		})

		Convey("When create source with only device ID", func() {
			params := data.Map{
				"device_id": data.Int(0),
			}
			Convey("Then capture should set default values", func() {
				s, err := sc.createCaptureFromDevice(ctx, ioParams, params)
				So(err, ShouldBeNil)
				capture, ok := s.(*captureFromDevice)
				So(ok, ShouldBeTrue)
				So(capture.deviceID, ShouldEqual, 0)
				So(capture.width, ShouldEqual, 0)
				So(capture.height, ShouldEqual, 0)
				So(capture.fps, ShouldEqual, 0)
			})
		})

		Convey("When create source with invalid device ID", func() {
			params := data.Map{
				"device_id": data.String("a"),
			}
			Convey("Then creator should occur parse errors", func() {
				s, err := sc.CreateSource(ctx, ioParams, params)
				So(err, ShouldNotBeNil)
				So(s, ShouldBeNil)
			})
		})

		Convey("When create source with invalid option parameters", func() {
			params := data.Map{
				"device_id": data.Int(0),
			}
			testMap := data.Map{
				"format": data.False,
				"width":  data.String("a"),
				"height": data.String("b"),
				"fps":    data.String("@"),
			}
			for k, v := range testMap {
				v := v
				msg := fmt.Sprintf("with %v error", k)
				Convey("Then creator should occur a parse error on option parameters "+msg,
					func() {
						params[k] = v
						s, err := sc.CreateSource(ctx, ioParams, params)
						So(err, ShouldNotBeNil)
						So(s, ShouldBeNil)
					})
			}
		})
	})
}

func TestGetDeviceSourceCreatorWithRawMode(t *testing.T) {
	ctx := &core.Context{}
	ioParams := &bql.IOParams{}
	Convey("Given a raw mode enabled capture source creator", t, func() {
		sc := FromDeviceCreator{}
		Convey("When create source with not supported format", func() {
			params := data.Map{
				"device_id": data.Int(0),
				"format":    data.String("4k"),
			}
			_, err := sc.createCaptureFromDevice(ctx, ioParams, params)
			Convey("Then capture should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
