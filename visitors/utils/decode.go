package utils

import (
	"net/url"
	"strings"
)

type Rc4StringDecoder struct {
	stringArray []string
	indexOffset int
	stringCache map[string]string
	firstTime   bool
}

func NewRc4StringDecoder(stringArray []string, indexOffset int) *Rc4StringDecoder {
	return &Rc4StringDecoder{
		stringArray: stringArray,
		indexOffset: indexOffset,
		stringCache: make(map[string]string),
		firstTime:   true,
	}
}

func (d *Rc4StringDecoder) Get(index int, key string) string {
	cacheKey := string(rune(index)) + d.stringArray[0]
	if cached, ok := d.stringCache[cacheKey]; ok {
		return cached
	}

	encoded := d.stringArray[index+d.indexOffset]
	str := d.rc4Decode(encoded, key)
	d.stringCache[cacheKey] = str
	//fmt.Println("encoded:", encoded)
	return str
}

func (d *Rc4StringDecoder) GetForRotate(index int, key string) (string, bool) {
	if d.firstTime {
		d.firstTime = false
		return "", true
	}
	//fmt.Println("called with", index, key, d.stringArray)
	return d.Get(index, key), false
}

func (d *Rc4StringDecoder) rc4Decode(str string, key string) string {
	str = Base64Transform(str)

	s := make([]int, 256)
	for i := 0; i < 256; i++ {
		s[i] = i
	}

	// KSA
	j := 0
	for i := 0; i < 256; i++ {
		j = (j + s[i] + int(key[i%len(key)])) % 256
		s[i], s[j] = s[j], s[i]
	}

	// PRGA
	i := 0
	j = 0
	result := make([]byte, len(str))

	for y := 0; y < len(str); y++ {
		i = (i + 1) % 256
		j = (j + s[i]) % 256
		s[i], s[j] = s[j], s[i]
		result[y] = str[y] ^ byte(s[(s[i]+s[j])%256])
	}

	return string(result)
}

const Base64Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/="

func Base64Transform(str string) string {
	var a strings.Builder
	c := 0
	d := 0

	for i := 0; i < len(str); i++ {
		e := str[i]
		idx := strings.IndexByte(Base64Alphabet, e)

		if idx != -1 {
			if c%4 != 0 {
				d = d*64 + idx
			} else {
				d = idx
			}

			c++
			if c%4 != 0 {
				shift := (-2 * c) & 6
				char := byte(255 & (d >> shift))
				a.WriteByte(char)
			}
		}
	}

	// Convert to percent-encoded string
	result := a.String()
	var encoded strings.Builder
	for i := 0; i < len(result); i++ {
		encoded.WriteString("%")
		b := result[i]
		if b < 16 {
			encoded.WriteString("0")
		}
		encoded.WriteString(strings.ToLower(string("0123456789abcdef"[b>>4])))
		encoded.WriteString(strings.ToLower(string("0123456789abcdef"[b&0xf])))
	}

	// Decode the percent-encoded string
	decoded, err := url.QueryUnescape(encoded.String())
	if err != nil {
		return ""
	}

	return decoded
}
