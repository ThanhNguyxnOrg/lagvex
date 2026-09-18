package client

import (
	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
)

// Type aliases for seamless access to protocol-level FEC primitives.
type (
	FECEncoderConfig      = protocol.FECEncoderConfig
	FECEncoder            = protocol.FECEncoder
	FECDecoder            = protocol.FECDecoder
	AdaptiveFECController = protocol.AdaptiveFECController
)

func DefaultFECEncoderConfig() FECEncoderConfig {
	return protocol.DefaultFECEncoderConfig()
}

func NewFECEncoder(cfg FECEncoderConfig) *FECEncoder {
	return protocol.NewFECEncoder(cfg)
}

func NewFECDecoder(bufferSize int) *FECDecoder {
	return protocol.NewFECDecoder(bufferSize)
}

func NewAdaptiveFECController() *AdaptiveFECController {
	return protocol.NewAdaptiveFECController()
}
