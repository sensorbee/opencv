[![wercker status](https://app.wercker.com/status/4db748e2b586121e924f83ef991a5f7b/s "wercker status")](https://app.wercker.com/project/bykey/4db748e2b586121e924f83ef991a5f7b)

# OpenCV plug-in for SensorBee

This plug-in is a library to use [OpenCV](http://opencv.org) library, User can use a part of OpenCV functions. For example user can create source component to generate stream video capturing.

# Require

* OpenCV
    * attention that ffmpeg version.
    * ex) Mac OS X `brew install opencv --with-ffmpeg`
* SensorBee
    * later v0.5

# Usage

## Registering plug-in

`build_sensorbee` with build.yaml set `gopkg.in/sensorbee/opencv.v0/plugin`

### build.yaml

```yaml
plugins:
- gopkg.in/sensorbee/opencv.v0/plugin
```

## Using from BQLs sample

### Capturing video source and streaming frames

```sql
-- capturing
CREATE PAUSED SOURCE camera1_avi TYPE opencv_capture_from_uri WITH
    uri="video/camera1.avi",
    frame_skip=4, next_frame_error=false;
```

will start generating stream from "video/camera1.avi" after execute `RESUME` query.

```
RESUME SOURCE camera1_avi;
```
