// package zaplipgloss implements a zap encoder that formats its body using charmbracelet lipgloss.
package zaplipgloss

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type LipglossEncoder struct {
	cfg            zapcore.EncoderConfig
	buf            *buffer.Buffer
	openNamespaces int

	// for encoding generic values by reflection
	reflectBuf *buffer.Buffer
	reflectEnc zapcore.ReflectedEncoder
}

// NewLipglossEncoder creates a new Lipgloss Zap Encoder.
func NewLipglossEncoder(cfg zapcore.EncoderConfig) (*LipglossEncoder, error) {
	// If no EncoderConfig.NewReflectedEncoder is provided by the user, then use default
	if cfg.NewReflectedEncoder == nil {
		cfg.NewReflectedEncoder = func(w io.Writer) zapcore.ReflectedEncoder {
			return json.NewEncoder(w)
		}
	}

	buf := buffer.NewPool().Get()
	reflectBuf := buffer.NewPool().Get()

	return &LipglossEncoder{
		cfg:            cfg,
		buf:            buf,
		reflectBuf:     reflectBuf,
		openNamespaces: 0,
		reflectEnc:     cfg.NewReflectedEncoder(reflectBuf),
	}, nil
}

func (l *LipglossEncoder) addElementSeparator() {
	last := l.buf.Len() - 1

	if last < 0 {
		return
	}

	l.buf.AppendByte(' ')
}

func (l *LipglossEncoder) addKey(key string) {
	l.addElementSeparator()
	l.buf.WriteString(key)
	l.buf.WriteByte('=')
}

func (l *LipglossEncoder) AddInt(k string, v int)         { l.AddInt64(k, int64(v)) }
func (l *LipglossEncoder) AddInt32(k string, v int32)     { l.AddInt64(k, int64(v)) }
func (l *LipglossEncoder) AddInt16(k string, v int16)     { l.AddInt64(k, int64(v)) }
func (l *LipglossEncoder) AddInt8(k string, v int8)       { l.AddInt64(k, int64(v)) }
func (l *LipglossEncoder) AddUint(k string, v uint)       { l.AddUint64(k, uint64(v)) }
func (l *LipglossEncoder) AddUint32(k string, v uint32)   { l.AddUint64(k, uint64(v)) }
func (l *LipglossEncoder) AddUint16(k string, v uint16)   { l.AddUint64(k, uint64(v)) }
func (l *LipglossEncoder) AddUint8(k string, v uint8)     { l.AddUint64(k, uint64(v)) }
func (l *LipglossEncoder) AddUintptr(k string, v uintptr) { l.AddUint64(k, uint64(v)) }
func (l *LipglossEncoder) AppendComplex64(v complex64)    { l.appendComplex(complex128(v), 32) }
func (l *LipglossEncoder) AppendComplex128(v complex128)  { l.appendComplex(complex128(v), 64) }
func (l *LipglossEncoder) AppendFloat64(v float64)        { l.appendFloat(v, 64) }
func (l *LipglossEncoder) AppendFloat32(v float32)        { l.appendFloat(float64(v), 32) }
func (l *LipglossEncoder) AppendInt(v int)                { l.AppendInt64(int64(v)) }
func (l *LipglossEncoder) AppendInt32(v int32)            { l.AppendInt64(int64(v)) }
func (l *LipglossEncoder) AppendInt16(v int16)            { l.AppendInt64(int64(v)) }
func (l *LipglossEncoder) AppendInt8(v int8)              { l.AppendInt64(int64(v)) }
func (l *LipglossEncoder) AppendUint(v uint)              { l.AppendUint64(uint64(v)) }
func (l *LipglossEncoder) AppendUint32(v uint32)          { l.AppendUint64(uint64(v)) }
func (l *LipglossEncoder) AppendUint16(v uint16)          { l.AppendUint64(uint64(v)) }
func (l *LipglossEncoder) AppendUint8(v uint8)            { l.AppendUint64(uint64(v)) }
func (l *LipglossEncoder) AppendUintptr(v uintptr)        { l.AppendUint64(uint64(v)) }

// appendComplex appends the encoded form of the provided complex128 value.
// precision specifies the encoding precision for the real and imaginary
// components of the complex number.
func (l *LipglossEncoder) appendComplex(val complex128, precision int) {
	l.addElementSeparator()

	realComponent, imaginaryComponent := float64(real(val)), float64(imag(val))

	l.buf.AppendByte('"')
	l.buf.AppendFloat(realComponent, precision)

	if imaginaryComponent >= 0 {
		l.buf.AppendByte('+')
	}

	l.buf.AppendFloat(imaginaryComponent, precision)
	l.buf.AppendByte('i')
	l.buf.AppendByte('"')
}

// Logging-specific marshalers.
func (l *LipglossEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	l.addKey(key)

	return l.AppendArray(marshaler)
}

func (l *LipglossEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	l.addElementSeparator()
	l.buf.AppendByte('[')
	err := arr.MarshalLogArray(l)
	l.buf.AppendByte(']')

	if err != nil {
		return fmt.Errorf("marshal array: %w", err)
	}

	return nil
}

// Built-in types.
func (l *LipglossEncoder) AppendBool(val bool) {
	l.addElementSeparator()

	if val {
		l.buf.WriteString("true")
	} else {
		l.buf.WriteString("false")
	}
}

func (l *LipglossEncoder) AppendByteString(val []byte) {
	l.addElementSeparator()
	l.safeAddByteString(val)
}

func (l *LipglossEncoder) safeAddByteString(val []byte) {
	safeAppendStringLike(
		(*buffer.Buffer).AppendBytes,
		utf8.DecodeRune,
		l.buf,
		val,
	)
}

func (l *LipglossEncoder) appendFloat(val float64, bitSize int) {
	l.addElementSeparator()

	switch {
	case math.IsNaN(val):
		l.buf.AppendString(`"NaN"`)
	case math.IsInf(val, 1):
		l.buf.AppendString(`"+Inf"`)
	case math.IsInf(val, -1):
		l.buf.AppendString(`"-Inf"`)
	default:
		l.buf.AppendFloat(val, bitSize)
	}
}

// safeAddString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's encoder, it doesn't attempt to protect the
// user from browser vulnerabilities or JSONP-related problems.
func (l *LipglossEncoder) safeAddString(val string) {
	safeAppendStringLike(
		(*buffer.Buffer).AppendString,
		utf8.DecodeRuneInString,
		l.buf,
		val,
	)
}

func (l *LipglossEncoder) AppendInt64(val int64) {
	l.buf.AppendInt(val)
}

func (l *LipglossEncoder) AppendString(val string) {
	l.safeAddString(val)
}

func (l *LipglossEncoder) AppendUint64(val uint64) {
	l.buf.AppendUint(val)
}

// Time-related types.
func (l *LipglossEncoder) AppendDuration(val time.Duration) {
	l.buf.AppendString(val.String())
}

func (l *LipglossEncoder) AppendTime(val time.Time) {
	l.buf.AppendString(val.Format(time.DateTime))
}

func (l *LipglossEncoder) AppendObject(val zapcore.ObjectMarshaler) (_ error) {
	old := l.openNamespaces
	l.openNamespaces = 0
	l.addElementSeparator()
	l.buf.AppendByte('{')
	err := val.MarshalLogObject(l)
	l.buf.AppendByte('}')
	l.closeOpenNamespaces()
	l.openNamespaces = old

	if err != nil {
		return fmt.Errorf("marshal object: %w", err)
	}

	return nil
}

func (l *LipglossEncoder) closeOpenNamespaces() {
	for range l.openNamespaces {
		l.buf.AppendByte(')')
	}

	l.openNamespaces = 0
}

// AppendReflected uses reflection to serialize arbitrary objects, so it's
// slow and allocation-heavy.
func (l *LipglossEncoder) AppendReflected(value any) error {
	valueBytes, err := l.encodeReflected(value)
	if err != nil {
		return fmt.Errorf("encode reflected: %w", err)
	}

	l.addElementSeparator()

	_, err = l.buf.Write(valueBytes)
	if err != nil {
		return fmt.Errorf("encode reflected: %w", err)
	}

	return nil
}

func (l *LipglossEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	l.addKey(key)

	return l.AppendObject(marshaler)
}

// Built-in types.
func (l *LipglossEncoder) AddBinary(key string, value []byte) {
	l.AddString(key, base64.StdEncoding.EncodeToString(value))
}

func (l *LipglossEncoder) AddByteString(key string, value []byte) {
	l.addKey(key)
	l.AppendByteString(value)
}

func (l *LipglossEncoder) AddBool(key string, value bool) {
	l.addKey(key)
	l.AppendBool(value)
}

func (l *LipglossEncoder) AddComplex128(key string, value complex128) {
	l.addKey(key)
	l.AppendComplex128(value)
}

func (l *LipglossEncoder) AddComplex64(key string, value complex64) {
	l.addKey(key)
	l.AppendComplex64(value)
}

func (l *LipglossEncoder) AddDuration(key string, value time.Duration) {
	l.addKey(key)
	l.AppendDuration(value)
}

func (l *LipglossEncoder) AddFloat64(key string, value float64) {
	l.addKey(key)
	l.AppendFloat64(value)
}

func (l *LipglossEncoder) AddFloat32(key string, value float32) {
	l.addKey(key)
	l.AppendFloat32(value)
}

func (l *LipglossEncoder) AddInt64(key string, value int64) {
	l.addKey(key)
	l.AppendInt64(value)
}

func (l *LipglossEncoder) AddString(key string, value string) {
	l.addKey(key)
	l.AppendString(value)
}

func (l *LipglossEncoder) AddTime(key string, value time.Time) {
	l.addKey(key)
	l.AppendTime(value)
}

func (l *LipglossEncoder) AddUint64(key string, value uint64) {
	l.addKey(key)
	l.AppendUint64(value)
}

// AddReflected uses reflection to serialize arbitrary objects, so it can be
// slow and allocation-heavy.
func (l *LipglossEncoder) AddReflected(key string, value interface{}) error {
	valueBytes, err := l.encodeReflected(value)
	if err != nil {
		return err
	}

	l.addKey(key)

	_, err = l.buf.Write(valueBytes)
	if err != nil {
		return fmt.Errorf("add reflected: %w", err)
	}

	return nil
}

func (l *LipglossEncoder) encodeReflected(obj interface{}) ([]byte, error) {
	if obj == nil {
		return []byte("nil"), nil
	}

	l.resetReflectBuf()

	if err := l.reflectEnc.Encode(obj); err != nil {
		return nil, fmt.Errorf("encode reflected: %w", err)
	}

	l.reflectBuf.TrimNewline()

	return l.reflectBuf.Bytes(), nil
}

func (l *LipglossEncoder) resetReflectBuf() {
	if l.reflectBuf == nil {
		l.reflectBuf = buffer.NewPool().Get()
		l.reflectEnc = l.cfg.NewReflectedEncoder(l.reflectBuf)
	} else {
		l.reflectBuf.Reset()
	}
}

// OpenNamespace opens an isolated namespace where all subsequent fields will
// be added. Applications can use namespaces to prevent key collisions when
// injecting loggers into sub-components or third-party libraries.
func (l *LipglossEncoder) OpenNamespace(key string) {
	l.addKey(key)
	l.buf.WriteByte('(')

	l.openNamespaces++
}

// Clone copies the encoder, ensuring that adding fields to the copy doesn't
// affect the original.
//
//nolint:ireturn // Gosh, Clone returns cloned interfaces. Couldn't be more shocked.
func (l *LipglossEncoder) Clone() zapcore.Encoder {
	clone, _ := NewLipglossEncoder(l.cfg)
	clone.buf.Write(l.buf.Bytes())

	return clone
}

func (l *LipglossEncoder) clone() *LipglossEncoder {
	clone, _ := NewLipglossEncoder(l.cfg)

	return clone
}

// EncodeEntry encodes an entry and fields, along with any accumulated
// context, into a byte buffer and returns it. Any fields that are empty,
// including fields on the `Entry` type, should be omitted.
func (l *LipglossEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := l.clone()

	var style lipgloss.Style

	var icon string

	switch entry.Level {
	case zapcore.InvalidLevel:
		style = invalidStyle
		icon = invalidLevel
	case zap.InfoLevel:
		style = infoStyle
		icon = iconInfo
	case zap.DPanicLevel:
		style = dPanicStyle
		icon = iconDPanic
	case zap.PanicLevel:
		style = panicStyle
		icon = iconPanic
	case zap.WarnLevel:
		style = warnStyle
		icon = iconWarn
	case zap.DebugLevel:
		style = debugStyle
		icon = iconDebug
	case zap.ErrorLevel:
		style = errStyle
		icon = iconErr
	case zap.FatalLevel:
		style = fatalStyle
		icon = iconFatal
	}

	final.buf.WriteString(style.Render(icon, entry.Time.Format(time.TimeOnly)))
	final.buf.WriteString(invert(style).Render(pline))
	final.buf.WriteString(" ")
	final.buf.WriteString(entry.Message)

	for i := range fields {
		temp := final.clone()
		fields[i].AddTo(temp)
		final.buf.WriteString(" ")
		final.buf.WriteString(fieldstyle(i).Render(temp.buf.String()))
	}

	final.buf.WriteString("\n")

	return final.buf, nil
}

const _hex = "0123456789abcdef"

// safeAppendStringLike is a generic implementation of safeAddString and safeAddByteString.
// It appends a string or byte slice to the buffer, escaping all special characters.
func safeAppendStringLike[S []byte | string](
	appendTo func(*buffer.Buffer, S),
	decodeRune func(S) (rune, int),
	buf *buffer.Buffer,
	strLike S,
) {
	last := 0

	for idx := 0; idx < len(strLike); {
		if strLike[idx] >= utf8.RuneSelf {
			r, size := decodeRune(strLike[idx:])
			if r != utf8.RuneError || size != 1 {
				idx += size

				continue
			}

			appendTo(buf, strLike[last:idx])
			buf.AppendString(`\ufffd`)

			idx++
			last = idx
		} else {
			if strLike[idx] >= 0x20 && strLike[idx] != '\\' && strLike[idx] != '"' {
				idx++

				continue
			}

			appendTo(buf, strLike[last:idx])

			safeAppendStringLikeChar(idx, strLike, buf)

			idx++
			last = idx
		}
	}

	// add remaining
	appendTo(buf, strLike[last:])
}

func safeAppendStringLikeChar[S []byte | string](idx int, strLike S, buf *buffer.Buffer) {
	switch strLike[idx] {
	case '\\', '"':
		buf.AppendByte('\\')
		buf.AppendByte(strLike[idx])
	case '\n':
		buf.AppendByte('\\')
		buf.AppendByte('n')
	case '\r':
		buf.AppendByte('\\')
		buf.AppendByte('r')
	case '\t':
		buf.AppendByte('\\')
		buf.AppendByte('t')
	default:
		buf.AppendString(`\u00`)
		buf.AppendByte(_hex[strLike[idx]>>4])
		buf.AppendByte(_hex[strLike[idx]&0xF])
	}
}

//nolint:gochecknoinits // As required by upstream.
func init() {
	err := zap.RegisterEncoder("lipgloss", func(ec zapcore.EncoderConfig) (zapcore.Encoder, error) {
		return NewLipglossEncoder(ec)
	})
	if err != nil {
		panic(err)
	}
}
