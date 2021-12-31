package config

import (
	"net"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestListFlags(t *testing.T) {
	expected := map[string]interface{}{
		"bool":           true,
		"int":            int(1),
		"intSlice":       []int{1, 2, 3},
		"int8":           int8(2),
		"int16":          int16(3),
		"int32":          int32(4),
		"int32Slice":     []int32{4, 5, 6},
		"int64":          int64(5),
		"int64Slice":     []int64{7, 8, 9},
		"uint":           uint(6),
		"uintSlice":      []uint{1, 2, 3},
		"uint8":          uint8(7),
		"uint16":         uint16(8),
		"uint32":         uint32(9),
		"uint64":         uint64(10),
		"float32":        float32(1.0),
		"float32Slice":   []float32{1, 2},
		"float64":        float64(2.0),
		"float64Slice":   []float64{3, 4},
		"stringSlice":    []string{"a", "b", "c"},
		"stringArray":    []string{"a", "b", "c"},
		"stringToInt":    map[string]int{"a": 1},
		"stringToInt64":  map[string]int64{"a": 2},
		"stringToString": map[string]string{"a": "b"},
		"bytesBase64":    []byte{49},
		"bytesHex":       []byte{255},
		"count":          int(1),
		"duration":       5 * time.Second,
		"durationSlice":  []time.Duration{2 * time.Minute},
		"ip":             net.IPv4(127, 0, 0, 1),
		"ipMask":         net.IPv4Mask(255, 255, 255, 0),
		"ipNet":          net.IPNet{IP: net.IPv4(192, 168, 0, 0).To4(), Mask: net.IPv4Mask(255, 255, 0, 0)},
		"ipSlice":        []net.IP{net.IPv4(1, 2, 3, 4)},
	}
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	f.Bool("bool", true, "")
	f.Int("int", 1, "")
	f.IntSlice("intSlice", []int{1, 2, 3}, "")
	f.Int8("int8", 2, "")
	f.Int16("int16", 3, "")
	f.Int32("int32", 4, "")
	f.Int32Slice("int32Slice", []int32{4, 5, 6}, "")
	f.Int64("int64", 5, "")
	f.Int64Slice("int64Slice", []int64{7, 8, 9}, "")
	f.Uint("uint", 6, "")
	f.UintSlice("uintSlice", []uint{1, 2, 3}, "")
	f.Uint8("uint8", 7, "")
	f.Uint16("uint16", 8, "")
	f.Uint32("uint32", 9, "")
	f.Uint64("uint64", 10, "")
	f.Float32("float32", 1.0, "")
	f.Float32Slice("float32Slice", []float32{1, 2}, "")
	f.Float64("float64", 2.0, "")
	f.Float64Slice("float64Slice", []float64{3, 4}, "")
	f.StringSlice("stringSlice", []string{"a", "b", "c"}, "")
	f.StringArray("stringArray", []string{"a", "b", "c"}, "")
	f.StringToInt("stringToInt", map[string]int{"a": 1}, "")
	f.StringToInt64("stringToInt64", map[string]int64{"a": 2}, "")
	f.StringToString("stringToString", map[string]string{"a": "b"}, "")
	f.BytesBase64("bytesBase64", []byte{49}, "")
	f.BytesHex("bytesHex", []byte{255}, "")
	f.Count("count", "")
	f.Duration("duration", 5*time.Second, "")
	f.DurationSlice("durationSlice", []time.Duration{2 * time.Minute}, "")
	f.IP("ip", net.IPv4(127, 0, 0, 1), "")
	f.IPMask("ipMask", net.IPv4Mask(255, 255, 255, 0), "")
	f.IPNet("ipNet", net.IPNet{IP: net.IPv4(192, 168, 0, 0).To4(), Mask: net.IPv4Mask(255, 255, 0, 0)}, "")
	f.IPSlice("ipSlice", []net.IP{net.IPv4(1, 2, 3, 4)}, "")
	_ = f.Parse([]string{"--count"})
	result := listFlags(f, true)
	assert.Equal(t, expected, result)
}
