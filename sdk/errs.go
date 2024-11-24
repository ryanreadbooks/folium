package sdk

import "fmt"

var (
	ErrGetIdFailed         = fmt.Errorf("get id failed")
	ErrWrongRequestFormat  = fmt.Errorf("wrong request format")
	ErrFolium              = fmt.Errorf("folium server error")
	ErrResultNotRecognized = fmt.Errorf("result format unrecognizable")
	ErrFoliumNotConnected  = fmt.Errorf("folium server not connected")
)
