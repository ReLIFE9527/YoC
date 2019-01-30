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

const PackTailByte = '\n'

func (e *PacketError) Error() string {
	return e.Obj + " {" + e.Op + ": " + e.Err.Error() + "}"
}

func PackString(src string) (dst string) {
	dst = "PackHeader//Length:" + strconv.FormatInt(int64(len(src)), 10) + "//" + src + "//PackTail//" + string(PackTailByte)
	return dst
}

func DePackString(src string) (dst string, err error) {
	if int(src[len(src)-1]) == PackTailByte {
		src = src[:len(src)-1]
	}
	strArr := strings.Split(src, "//")
	for i := 0; i < len(strArr); i++ {
		if strArr[i] == "PackHeader" {
			length, err := getLength(strArr[i+1])
			if err != nil || len(strArr[i+2]) != int(length) || strArr[i+3] != "PackTail" {
				Log.Println(err)
				i += 3
				continue
			} else {
				dst = strArr[i+2]
				break
			}
		}
	}
	if dst == "" {
		err = &PacketError{src, "DePack", errors.New("can not dePack")}
	}
	return dst, err
}

func getLength(str string) (len int64, err error) {
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
		if !strings.Contains(stream, "\""+properties[i]+"\"") {
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
