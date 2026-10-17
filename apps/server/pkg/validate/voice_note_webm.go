package validate

import (
	"encoding/binary"
	"math"
)

type ebmlElement struct {
	id         uint64
	start, end int64
	unknown    bool
}

func (p *audioContainer) ebmlInteger(offset int64, id bool) (uint64, int64, bool, error) {
	first, err := p.read(offset, 1)
	if err != nil || first[0] == 0 {
		return 0, 0, false, ErrInvalidFileType
	}
	width, mask := 1, byte(0x80)
	for first[0]&mask == 0 {
		width++
		mask >>= 1
	}
	if (id && width > 4) || width > 8 {
		return 0, 0, false, ErrInvalidFileType
	}
	data, err := p.read(offset, width)
	if err != nil {
		return 0, 0, false, err
	}
	value := uint64(data[0])
	if !id {
		value &= uint64(mask - 1)
	}
	for _, next := range data[1:] {
		value = value<<8 | uint64(next)
	}
	unknown := !id && value == (uint64(1)<<uint(7*width))-1
	return value, int64(width), unknown, nil
}

func (p *audioContainer) ebmlElements(start, end int64, visit func(ebmlElement) error) error {
	for offset := start; offset < end; {
		if err := p.step(); err != nil {
			return err
		}
		id, idWidth, _, err := p.ebmlInteger(offset, true)
		if err != nil {
			return err
		}
		size, sizeWidth, unknown, err := p.ebmlInteger(offset+idWidth, false)
		if err != nil {
			return err
		}
		dataStart := offset + idWidth + sizeWidth
		if dataStart > end || (!unknown && size > uint64(end-dataStart)) {
			return ErrInvalidFileType
		}
		dataEnd := end
		if !unknown {
			dataEnd = dataStart + int64(size)
		}
		if unknown && id != 0x18538067 && id != 0x1f43b675 {
			return ErrInvalidFileType
		}
		if err := visit(ebmlElement{id: id, start: dataStart, end: dataEnd, unknown: unknown}); err != nil {
			return err
		}
		offset = dataEnd
	}
	return nil
}

func (p *audioContainer) webm() error {
	var webm, audio, cluster bool
	err := p.ebmlElements(0, p.size, func(element ebmlElement) error {
		switch element.id {
		case 0x1a45dfa3: // EBML
			return p.ebmlElements(element.start, element.end, func(header ebmlElement) error {
				if header.id == 0x4282 { // DocType
					if header.end-header.start != 4 {
						return ErrInvalidFileType
					}
					data, err := p.read(header.start, 4)
					if err != nil {
						return err
					}
					webm = string(data) == "webm"
				}
				return nil
			})
		case 0x18538067: // Segment
			return p.ebmlElements(element.start, element.end, func(child ebmlElement) error {
				switch child.id {
				case 0x1654ae6b: // Tracks
					if audio {
						return ErrInvalidFileType
					}
					if err := p.webmAudioTracks(child); err != nil {
						return err
					}
					audio = true
				case 0x1549a966: // Info
					return p.webmDuration(child)
				case 0x1f43b675: // Cluster: skip encoded payload, including streaming unknown size.
					cluster = child.end > child.start
				}
				return nil
			})
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !webm || !audio || !cluster {
		return ErrInvalidFileType
	}
	return nil
}

func (p *audioContainer) webmAudioTracks(tracks ebmlElement) error {
	count := 0
	err := p.ebmlElements(tracks.start, tracks.end, func(track ebmlElement) error {
		if track.id != 0xae {
			return nil
		}
		count++
		var audio, opus bool
		err := p.ebmlElements(track.start, track.end, func(field ebmlElement) error {
			switch field.id {
			case 0x83: // TrackType
				if field.end-field.start != 1 {
					return ErrInvalidFileType
				}
				data, err := p.read(field.start, 1)
				if err != nil {
					return err
				}
				audio = data[0] == 2
			case 0x86: // CodecID
				if field.end-field.start != 6 {
					return ErrInvalidFileType
				}
				data, err := p.read(field.start, 6)
				if err != nil {
					return err
				}
				opus = string(data) == "A_OPUS"
			}
			return nil
		})
		if err != nil {
			return err
		}
		if !audio || !opus {
			return ErrInvalidFileType
		}
		return nil
	})
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrInvalidFileType
	}
	return nil
}

func (p *audioContainer) webmDuration(info ebmlElement) error {
	scale, duration := float64(1_000_000), float64(0)
	err := p.ebmlElements(info.start, info.end, func(field ebmlElement) error {
		switch field.id {
		case 0x2ad7b1: // TimestampScale, nanoseconds per tick.
			length := field.end - field.start
			if length < 1 || length > 8 {
				return ErrInvalidFileType
			}
			data, err := p.read(field.start, int(length))
			if err != nil {
				return err
			}
			value := uint64(0)
			for _, b := range data {
				value = value<<8 | uint64(b)
			}
			if value == 0 {
				return ErrInvalidFileType
			}
			scale = float64(value)
		case 0x4489: // Duration, a float measured in ticks.
			length := field.end - field.start
			if length != 4 && length != 8 {
				return ErrInvalidFileType
			}
			data, err := p.read(field.start, int(length))
			if err != nil {
				return err
			}
			if length == 4 {
				duration = float64(math.Float32frombits(binary.BigEndian.Uint32(data)))
			} else {
				duration = math.Float64frombits(binary.BigEndian.Uint64(data))
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return p.checkDuration(duration * scale / 1_000_000)
}
