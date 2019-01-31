package Pack

import (
	. "../Log"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type PacketError struct {
	Obj string
	Op  string
	Err error
}

const TailByte = '\n'

func (e *PacketError) Error() string {
	return e.Obj + " {" + e.Op + ": " + e.Err.Error() + "}"
}

func PackString(src string) (dst string) {
	dst = "//PackHeader//Length:" + strconv.FormatInt(int64(len(src)), 10) + "//" + src + "//PackTail//" + string(TailByte)
	return dst
}

func DePackString(src string) (dst string, err error) {
	if int(src[len(src)-1]) == TailByte {
		src = src[:len(src)-1]
	}
	strArr := strings.SplitAfter(src, "//")
	for i := 0; i < len(strArr); i++ {
		if strArr[i] == "PackHeader//" {
			length, err := getLength(strArr[i+1])
			if err != nil {
				return "", nil
			}
			var j = i + 3
			for ; j < len(strArr); j++ {
				if strArr[j] == "PackTail//" {
					break
				}
			}
			for k := i + 2; k < j; k++ {
				dst += strArr[k]
			}
			dst = dst[:len(dst)-2]
			if int64(len(dst)) != length {
				Log.Println("length err : ", src, " ", dst)
			}
		}
	}
	if dst == "" {
		err = &PacketError{src, "DePack", errors.New("can not dePack")}
	}
	return dst, err
}

func getLength(str string) (len int64, err error) {
	str = strings.Replace(str, "//", "", -1)
	strArr := strings.Split(str, ":")
	if strArr[0] == "Length" {
		length, err := strconv.ParseInt(strArr[1], 10, 64)
		if err != nil {
			return 0, &PacketError{str, "getLength", err}
		} else {
			return length, nil
		}
	} else {
		return 0, &PacketError{str, "getLength", errors.New("para error")}
	}
}

func IsStreamValid(properties []string, stream string) bool {
	for i := 0; i < len(properties); i++ {
		var prop = "\"" + properties[i] + "\""
		if !strings.Contains(stream, prop) {
			return false
		}
	}
	return true
}

func Convert2Map(str string) (dst *map[string]string) {
	dst = new(map[string]string)
	err := json.Unmarshal([]byte(str), dst)
	if err != nil {
		Log.Println(err)
	}
	return dst
}
