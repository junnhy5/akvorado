// SPDX-FileCopyrightText: 2022 Free Mobile
// SPDX-License-Identifier: AGPL-3.0-only

package flow

import (
	"time"

	"akvorado/inlet/flow/decoder"
	"akvorado/inlet/flow/decoder/netflow"
	"akvorado/inlet/flow/decoder/sflow"
)

// Message describes a decoded flow message.
type Message = decoder.FlowMessage

type wrappedDecoder struct {
	c                         *Component
	orig                      decoder.Decoder
	useSrcAddrForExporterAddr bool
}

// Decode decodes a flow while keeping some stats.
func (wd *wrappedDecoder) Decode(in decoder.RawFlow) []*Message {
	timeTrackStart := time.Now()
	decoded := wd.orig.Decode(in)
	timeTrackStop := time.Now()

	if decoded == nil {
		wd.c.metrics.decoderErrors.WithLabelValues(wd.orig.Name()).
			Inc()
		return nil
	}

	if wd.useSrcAddrForExporterAddr {
		for _, f := range decoded {
			f.ExporterAddress = in.Source.To16()
		}
	}

	wd.c.metrics.decoderTime.WithLabelValues(wd.orig.Name()).
		Observe(float64((timeTrackStop.Sub(timeTrackStart)).Nanoseconds()) / 1000 / 1000 / 1000)
	wd.c.metrics.decoderStats.WithLabelValues(wd.orig.Name()).
		Inc()
	return decoded
}

// Name returns the name of the original decoder.
func (wd *wrappedDecoder) Name() string {
	return wd.orig.Name()
}

// wrapDecoder wraps the provided decoders to get statistics from it.
func (c *Component) wrapDecoder(d decoder.Decoder, useSrcAddrForExporterAddr bool) decoder.Decoder {
	return &wrappedDecoder{
		c:                         c,
		orig:                      d,
		useSrcAddrForExporterAddr: useSrcAddrForExporterAddr,
	}
}

var decoders = map[string]decoder.NewDecoderFunc{
	"netflow": netflow.New,
	"sflow":   sflow.New,
}
