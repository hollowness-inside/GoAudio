package wave

import (
	"encoding/binary"
	"io"
	"math"
	"os"
)

// Consts that appear in the .WAVE file format
var (
	ChunkID          = []byte{0x52, 0x49, 0x46, 0x46} // RIFF
	BigEndianChunkID = []byte{0x52, 0x49, 0x46, 0x58} // RIFX
	WaveID           = []byte{0x57, 0x41, 0x56, 0x45} // WAVE
	Format           = []byte{0x66, 0x6d, 0x74, 0x20} // FMT
	Subchunk2ID      = []byte{0x64, 0x61, 0x74, 0x61} // DATA
)

type appendIntFunc func(b []byte, i int) []byte

var (
	// appendIntFm to map X-bit int to function appending bytes to buffer
	//
	appendIntFm = map[int]appendIntFunc{
		16: appendInt16,
		32: appendInt32,
	}
)

// WriteFrames writes the slice to disk as a .wav file
// the WaveFmt metadata needs to be correct
// WaveData and WaveHeader are inferred from the samples however..
func WriteFrames(samples []Frame, wfmt WaveFmt, file string) error {
	return WriteWaveFile(samples, wfmt, file)
}

func WriteWaveFile(samples []Frame, wfmt WaveFmt, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	return WriteWaveToWriter(samples, wfmt, f)
}

func WriteWaveToWriter(samples []Frame, wfmt WaveFmt, writer io.Writer) error {
	wfb := fmtToBytes(wfmt)
	data, databits := framesToData(samples, wfmt)
	hdr := createHeader(data)

	_, err := writer.Write(hdr)
	if err != nil {
		return err
	}
	_, err = writer.Write(wfb)
	if err != nil {
		return err
	}
	_, err = writer.Write(databits)
	if err != nil {
		return err
	}

	return nil
}

func appendInt16(b []byte, i int) []byte {
	in := uint16(i)
	return binary.LittleEndian.AppendUint16(b, in)
}

func appendInt32(b []byte, i int) []byte {
	in := uint32(i)
	return binary.LittleEndian.AppendUint32(b, in)
}

func framesToData(frames []Frame, wfmt WaveFmt) (WaveData, []byte) {
	raw := samplesToRawData(frames, wfmt)

	// We receive frames but have to store the size of the samples
	// The size of the samples is frames / channels..
	subchunksize := (len(frames) * wfmt.NumChannels * wfmt.BitsPerSample) / 8

	// construct the data part..
	b := make([]byte, 0, 8+len(raw))
	b = append(b, Subchunk2ID...)
	b = appendInt32(b, subchunksize)
	b = append(b, raw...)

	wd := WaveData{
		Subchunk2ID:   Subchunk2ID,
		Subchunk2Size: subchunksize,
		RawData:       raw,
		Frames:        frames,
	}
	return wd, b
}

func floatToBytes(f float64, nBytes int) []byte {
	bits := math.Float64bits(f)
	bs := make([]byte, 0, 8)
	binary.LittleEndian.PutUint64(bs, bits)
	// trim padding
	switch nBytes {
	case 2:
		return bs[:2]
	case 4:
		return bs[:4]
	}
	return bs
}

// Turn the samples into raw data...
func samplesToRawData(samples []Frame, props WaveFmt) []byte {
	raw := []byte{}
	for _, s := range samples {
		// the samples are scaled - rescale them?
		rescaled := rescaleFrame(s, props.BitsPerSample)
		raw = appendIntFm[props.BitsPerSample](raw, rescaled)
	}
	return raw
}

// rescale frames back to the original values..
func rescaleFrame(s Frame, bits int) int {
	rescaled := float64(s) * float64(maxValues[bits])
	return int(rescaled)
}

func fmtToBytes(wfmt WaveFmt) []byte {
	b := []byte{}

	subchunksize := int32ToBytes(wfmt.Subchunk1Size)
	audioformat := int16ToBytes(wfmt.AudioFormat)
	numchans := int16ToBytes(wfmt.NumChannels)
	sr := int32ToBytes(wfmt.SampleRate)
	br := int32ToBytes(wfmt.ByteRate)
	blockalign := int16ToBytes(wfmt.BlockAlign)
	bitsPerSample := int16ToBytes(wfmt.BitsPerSample)

=======
	b := make([]byte, 0, 23)
>>>>>>> Stashed changes
	b = append(b, wfmt.Subchunk1ID...)
	b = appendInt32(b, wfmt.Subchunk1Size)
	b = appendInt16(b, wfmt.AudioFormat)
	b = appendInt16(b, wfmt.NumChannels)
	b = appendInt32(b, wfmt.SampleRate)
	b = appendInt32(b, wfmt.ByteRate)
	b = appendInt16(b, wfmt.BlockAlign)
	b = appendInt16(b, wfmt.BitsPerSample)

	return b
}

// turn the sample to a valid header
func createHeader(wd WaveData) []byte {
	// write chunkID

	chunksize := 36 + wd.Subchunk2Size

	bits := make([]byte, 0, 12)
	bits = append(bits, ChunkID...) // in theory switch on endianness..
	bits = appendInt32(bits, chunksize)
	bits = append(bits, WaveID...)

	return bits
}
