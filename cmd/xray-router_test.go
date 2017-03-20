package cmd

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestZoomFactor(t *testing.T) {

	jsontext := `{
  "frame": {
    "id": "48",
    "format": "17",
    "width": "960",
    "height": "720",
    "rotation": "2",
    "timestamp": "2295"
  },
  "faces": [
    {
      "id": "1",
      "eulerY": "0.0",
      "eulerZ": "16.868137",
      "height": "305.15756",
      "width": "244.12605",
      "leftEyeOpen": "-1.0",
      "rightEyeOpen": "0.588354",
      "similing": "0.007244766",
      "facePt1": {
        "x": "636.2853",
        "y": "332.01703"
      },
      "facePt2": {
        "x": "880.4114",
        "y": "637.1746"
      }
    }
  ]
}`

	var fr frameRecord
	if err := json.Unmarshal([]byte(jsontext), &fr); err != nil {
		t.Errorf("TestZoomFactor(): failed to unmarshal JSON: %v", err)
	}

	boundingBox, err := fr.GetFullFrameRect()
	if err != nil {
		t.Errorf("TestZoomFactor() error: %v", err)
	}

	faces, err := fr.GetFaceRectangles()
	if err != nil {
		t.Errorf("TestZoomFactor() error: %v", err)
	}

	zoom := calculateOptimalZoomFactor(faces, boundingBox)
	fmt.Println(zoom)
}
