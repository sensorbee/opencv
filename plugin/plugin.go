package plugin

import (
	"gopkg.in/sensorbee/opencv.v0"
	"gopkg.in/sensorbee/sensorbee.v0/bql"
	"gopkg.in/sensorbee/sensorbee.v0/bql/udf"
)

// initialize scouter components. this init method will be called by
// SensorBee customized main.go.
//
//  import(
//      _ "gopkg.in/sensorbee/opencv.v0/plugin"
//  )
func init() {
	// capture
	bql.MustRegisterGlobalSourceCreator("opencv_capture_from_uri",
		&opencv.FromURICreator{})
	bql.MustRegisterGlobalSourceCreator("opencv_capture_from_device",
		&opencv.FromDeviceCreator{})

	// cascade classifier
	udf.MustRegisterGlobalUDSCreator("opencv_cascade_classifier",
		udf.UDSCreatorFunc(opencv.NewCascadeClassifier))
	udf.MustRegisterGlobalUDF("opencv_detect_multi_scale",
		udf.MustConvertGeneric(opencv.DetectMultiScale))
	udf.MustRegisterGlobalUDF("opencv_draw_rects",
		udf.MustConvertGeneric(opencv.DrawRectsToImage))

	// mount image
	udf.MustRegisterGlobalUDSCreator("opencv_shared_image",
		udf.UDSCreatorFunc(opencv.NewSharedImage))
	udf.MustRegisterGlobalUDF("opencv_mount_image",
		udf.MustConvertGeneric(opencv.MountAlphaImage))
}
