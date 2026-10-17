package validate

import "encoding/binary"

type mp4Box struct {
	kind       string
	start, end int64
}

func (p *audioContainer) mp4Boxes(start, end int64, visit func(mp4Box) error) error {
	for offset := start; offset < end; {
		if err := p.step(); err != nil {
			return err
		}
		if end-offset < 8 {
			return ErrInvalidFileType
		}
		header, err := p.read(offset, 8)
		if err != nil {
			return err
		}
		size, headerSize := int64(binary.BigEndian.Uint32(header)), int64(8)
		if size == 1 {
			extended, err := p.read(offset+8, 8)
			if err != nil {
				return err
			}
			value := binary.BigEndian.Uint64(extended)
			if value > uint64(end-offset) {
				return ErrInvalidFileType
			}
			size, headerSize = int64(value), 16
		} else if size == 0 {
			size = end - offset
		}
		if size < headerSize || size > end-offset {
			return ErrInvalidFileType
		}
		if err := visit(mp4Box{kind: string(header[4:]), start: offset + headerSize, end: offset + size}); err != nil {
			return err
		}
		offset += size
	}
	return nil
}

func (p *audioContainer) mp4() error {
	var fileType, mediaData bool
	tracks := 0
	err := p.mp4Boxes(0, p.size, func(box mp4Box) error {
		switch box.kind {
		case "ftyp":
			fileType = box.end-box.start >= 8
		case "mdat":
			mediaData = mediaData || box.end > box.start
		case "moov":
			return p.mp4Boxes(box.start, box.end, func(child mp4Box) error {
				if child.kind != "trak" {
					return nil
				}
				tracks++
				return p.mp4AudioTrack(child)
			})
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !fileType || !mediaData || tracks != 1 {
		return ErrInvalidFileType
	}
	return nil
}

func (p *audioContainer) mp4AudioTrack(track mp4Box) error {
	var audio bool
	err := p.mp4Boxes(track.start, track.end, func(box mp4Box) error {
		if box.kind != "mdia" {
			return nil
		}
		return p.mp4Boxes(box.start, box.end, func(child mp4Box) error {
			switch child.kind {
			case "hdlr":
				if child.end-child.start < 12 {
					return ErrInvalidFileType
				}
				data, err := p.read(child.start+8, 4)
				if err != nil || string(data) != "soun" {
					return ErrInvalidFileType
				}
				audio = true
			case "mdhd":
				return p.mp4Duration(child)
			}
			return nil
		})
	})
	if err != nil {
		return err
	}
	if !audio {
		return ErrInvalidFileType
	}
	return nil
}

func (p *audioContainer) mp4Duration(box mp4Box) error {
	version, err := p.read(box.start, 1)
	if err != nil {
		return err
	}
	offset, length := int64(12), 8
	if version[0] == 1 {
		offset, length = 20, 12
	} else if version[0] != 0 {
		return ErrInvalidFileType
	}
	if box.end-box.start < offset+int64(length) {
		return ErrInvalidFileType
	}
	data, err := p.read(box.start+offset, length)
	if err != nil {
		return err
	}
	scale := binary.BigEndian.Uint32(data)
	if scale == 0 {
		return ErrInvalidFileType
	}
	duration := uint64(binary.BigEndian.Uint32(data[4:]))
	if version[0] == 1 {
		duration = binary.BigEndian.Uint64(data[4:])
	}
	// Fragmented browser MP4s may have a zero or unknown header duration.
	if duration == 0 || duration == ^uint64(0) || (version[0] == 0 && duration == uint64(^uint32(0))) {
		return nil
	}
	return p.checkDuration(float64(duration) / float64(scale) * 1000)
}
