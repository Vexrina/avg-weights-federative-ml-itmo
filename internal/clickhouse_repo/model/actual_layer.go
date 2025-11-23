package model

// так хотелось бы, чтобы был кодген...

import (
	"avg_weights_fed_ml_itmo/internal/app/model"
	"time"

	"github.com/google/uuid"
)

type ActualLayers struct {
	ClientID  uuid.UUID `ch:"clientID"`
	CreatedAt time.Time `ch:"createdAt"`
	UpdatedAt time.Time `ch:"updatedAt"`
	Layer1    []float64 `ch:"layer1"`
	Layer2    []float64 `ch:"layer2"`
	Layer3    []float64 `ch:"layer3"`
	Layer4    []float64 `ch:"layer4"`
	Layer5    []float64 `ch:"layer5"`
}

func MapDomainReqToDB(dom model.AddMyWeightsDomainReq, updatedAt time.Time) ActualLayers {
	res := ActualLayers{
		ClientID:  dom.ClientID,
		UpdatedAt: updatedAt,
	}
	for layerName, layer := range dom.Layers {
		switch layerName {
		case "layer1":
			res.Layer1 = layer
		case "layer2":
			res.Layer2 = layer
		case "layer3":
			res.Layer3 = layer
		case "layer4":
			res.Layer4 = layer
		case "layer5":
			res.Layer5 = layer
		}
	}
	return res
}
