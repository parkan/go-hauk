package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/parkan/go-hauk/config"
)

type velocityUnit struct {
	MpsMultiplier float64 `json:"mpsMultiplier"`
	Unit          string  `json:"unit"`
}

func (s *Server) handleDynamic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")

	var velUnit velocityUnit
	switch s.cfg.VelocityUnit {
	case config.MilesPerHour:
		velUnit = velocityUnit{MpsMultiplier: 3.6 * 0.6213712, Unit: "mph"}
	case config.MetersPerSecond:
		velUnit = velocityUnit{MpsMultiplier: 1, Unit: "m/s"}
	default:
		velUnit = velocityUnit{MpsMultiplier: 3.6, Unit: "km/h"}
	}

	tileURI, _ := json.Marshal(s.cfg.MapTileURI)
	attribution, _ := json.Marshal(s.cfg.MapAttribution)
	defaultZoom, _ := json.Marshal(s.cfg.DefaultZoom)
	maxZoom, _ := json.Marshal(s.cfg.MaxZoom)
	maxPoints, _ := json.Marshal(s.cfg.MaxShownPts)
	velDelta, _ := json.Marshal(s.cfg.VelocityDataPts)
	trailColor, _ := json.Marshal(s.cfg.TrailColor)
	velUnitJSON, _ := json.Marshal(velUnit)
	offlineTimeout, _ := json.Marshal(s.cfg.OfflineTimeout)
	requestTimeout, _ := json.Marshal(s.cfg.RequestTimeout)

	fmt.Fprintf(w, "var TILE_URI = %s;\n", tileURI)
	fmt.Fprintf(w, "var ATTRIBUTION = %s;\n", attribution)
	fmt.Fprintf(w, "var DEFAULT_ZOOM = %s;\n", defaultZoom)
	fmt.Fprintf(w, "var MAX_ZOOM = %s;\n", maxZoom)
	fmt.Fprintf(w, "var MAX_POINTS = %s;\n", maxPoints)
	fmt.Fprintf(w, "var VELOCITY_DELTA_TIME = %s;\n", velDelta)
	fmt.Fprintf(w, "var TRAIL_COLOR = %s;\n", trailColor)
	fmt.Fprintf(w, "var VELOCITY_UNIT = %s;\n", velUnitJSON)
	fmt.Fprintf(w, "var OFFLINE_TIMEOUT = %s;\n", offlineTimeout)
	fmt.Fprintf(w, "var REQUEST_TIMEOUT = %s;\n", requestTimeout)
}
