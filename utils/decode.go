package utils

import (
	"fmt"
	"net/url"
	"strings"
	"unicode/utf16"
)

type Rc4StringDecoder struct {
	StringArray []string
	indexOffset int
	stringCache map[string]string
	firstTime   bool
}

func NewRc4StringDecoder(stringArray []string, indexOffset int) *Rc4StringDecoder {
	return &Rc4StringDecoder{
		StringArray: stringArray,
		indexOffset: indexOffset,
		stringCache: make(map[string]string),
		firstTime:   true,
	}
}

func (d *Rc4StringDecoder) Get(index int, key string) string {
	cacheKey := string(rune(index)) + d.StringArray[0]
	if cached, ok := d.stringCache[cacheKey]; ok {
		return cached
	}

	encoded := d.StringArray[index+d.indexOffset]
	str := rc4Decode(encoded, key)
	d.stringCache[cacheKey] = str
	//fmt.Println("encoded:", encoded)
	return str
}

func (d *Rc4StringDecoder) GetForRotate(index int, key string) (string, bool) {
	if d.firstTime {
		d.firstTime = false
		return "", true
	}
	//fmt.Println("called with", index, key, d.StringArray)
	return d.Get(index, key), false
}

func (d *Rc4StringDecoder) Shift() {
	d.StringArray = append(d.StringArray[1:], d.StringArray[0])
}

const Base64Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/="

func base64Transform(str string) string {
	var a []byte
	c := 0
	d := 0

	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		ch := string(runes[i])
		e := strings.Index(Base64Alphabet, ch)
		if e != -1 {
			if c%4 != 0 {
				d = d*64 + e
			} else {
				d = e
			}
			if c%4 != 0 {
				shift := (-2 * (c + 1)) & 6
				b := byte(255 & (d >> shift))
				a = append(a, b)
			}
			c++
		}
	}

	var encodedBuilder strings.Builder
	for _, b := range a {
		fmt.Fprintf(&encodedBuilder, "%%%02x", b)
	}
	decoded, err := url.PathUnescape(encodedBuilder.String())
	if err != nil {
		return string(a)
	}
	return decoded
}

func jsUTF16Units(s string) []uint16 {
	// JS strings are sequences of UTF-16 code units.
	// Convert Go runes -> UTF-16 code units to emulate charCodeAt / length behavior.
	rs := []rune(s)
	return utf16.Encode(rs)
}

func rc4Decode(input string, key string) string {
	s := make([]int, 256)
	for i := 0; i < 256; i++ {
		s[i] = i
	}
	j := 0
	decoded := ""

	str := base64Transform(input)

	keyUnits := jsUTF16Units(key)
	keyLen := len(keyUnits)
	if keyLen == 0 {
		keyLen = 1 // avoid modulo by zero; JS charCodeAt on empty string would be NaN but modulo won't be used in typical keys
	}

	for i := 0; i < 256; i++ {
		j = (j + s[i] + int(keyUnits[i%keyLen])) % 256
		s[i], s[j] = s[j], s[i]
	}

	i := 0
	j = 0

	strUnits := jsUTF16Units(str)
	decodedUnits := make([]uint16, 0, len(strUnits))

	for y := 0; y < len(strUnits); y++ {
		i = (i + 1) % 256
		j = (j + s[i]) % 256
		s[i], s[j] = s[j], s[i]
		t := (s[i] + s[j]) % 256
		xorVal := uint16(strUnits[y]) ^ uint16(s[t])
		decodedUnits = append(decodedUnits, xorVal)
	}

	runes := utf16.Decode(decodedUnits)
	decoded = string(runes)
	return decoded
}
