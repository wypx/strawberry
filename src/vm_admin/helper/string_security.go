package helper

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
)

type encryption struct {
	text string
}

// 字符串转MD5
func (s *encryption) EncryptMD5() (passwd string) {
	ctx := md5.New()
	ctx.Write([]byte(s.text))
	return hex.EncodeToString(ctx.Sum(nil))
}

// 字符串转SHA1
func (s *encryption) EncryptSHA1() (passwd string) {
	ctx := sha1.New()
	ctx.Write([]byte(s.text))
	return hex.EncodeToString(ctx.Sum(nil))
}

// 字符串转SHA256
func (s *encryption) EncryptSHA256() (passwd string) {
	ctx := sha256.New()
	ctx.Write([]byte(s.text))
	return hex.EncodeToString(ctx.Sum(nil))
}
