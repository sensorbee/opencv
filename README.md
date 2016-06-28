[![wercker status](https://app.wercker.com/status/4db748e2b586121e924f83ef991a5f7b/s "wercker status")](https://app.wercker.com/project/bykey/4db748e2b586121e924f83ef991a5f7b)

# OpenCV plugin for SensorBee

This is the [OpenCV](http://opencv.org) plugin for SensorBee.

This plugin currently supports following features of OpenCV:

* Inputting video stream from a file or a camera
* Encoding JPEG
* Cascade classifier

# Requirements

* OpenCV with ffmpeg enabled for video input sources
    * Example on Mac OS X: `brew install homebrew/science/opencv --with-ffmpeg`
* SensorBee
    * v0.5 or later

# Usage

## Plugin Registration

Add `gopkg.in/sensorbee/opencv.v0/plugin` to build.yaml for the `build_sensorbee` command.

### build.yaml

```yaml
plugins:
- gopkg.in/sensorbee/opencv.v0/plugin
```

## BQL examples

### Capturing frames from a video file

```sql
-- capturing
CREATE PAUSED SOURCE camera1_avi TYPE opencv_capture_from_uri WITH
    uri="video/camera1.avi",
    frame_skip=4, next_frame_error=false;
```

This source will start generating a stream from "video/camera1.avi" after executing `RESUME` query.

```
RESUME SOURCE camera1_avi;
```

Note that `PAUSED` should not be specified when capturing from a webcam, which
keeps generating a video stream.
