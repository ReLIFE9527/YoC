package Pack

import (
	. "../Debug"
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

type Packet string
type Stream string

const TailByte = '\n'

func (e *PacketError) Error() string {
	return e.Obj + " {" + e.Op + ": " + e.Err.Error() + "}"
}

func StreamPack(src Stream) (dst Packet) {
	str := string(src)
	dst = Packet("//PackHeader//Length:" + strconv.FormatInt(int64(len(str)), 10) + "//" + str + "//PackTail//" + string(TailByte))
	return dst
}

func DePack(src Packet) (dst Stream, err error) {
	if int(src[len(src)-1]) == TailByte {
		src = src[:len(src)-1]
	}
	strArr := strings.SplitAfter(string(src), "//")
	for i := 0; i < len(strArr); i++ {
		if strArr[i] == "PackHeader//" {
			length, err := getLength(Packet(strArr[i+1]))
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
				dst += Stream(strArr[k])
			}
			dst = dst[:len(dst)-2]
			if int64(len(dst)) != length {
				DebugLogger.Println("length err : ", src, " ", dst)
			}
		}
	}
	if dst == "" {
		err = &PacketError{string(src), "DePack", errors.New("can not dePack")}
	}
	return dst, err
}

func getLength(src Packet) (len int64, err error) {
	var str = string(src)
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

func IsStreamValid(src Stream, properties []string) bool {
	for i := 0; i < len(properties); i++ {
		var prop = "\"" + properties[i] + "\""
		if !strings.Contains(string(src), prop) {
			return false
		}
	}
	return true
}

func Convert2Map(str Stream) (dst *map[string]string) {
	dst = new(map[string]string)
	err := json.Unmarshal([]byte(str), dst)
	if err != nil {
		DebugLogger.Println(err)
	}
	return dst
}

func Convert2Stream(src *map[string]string) (dst Stream) {
	bytes, err := json.Marshal(src)
	if err != nil {
		DebugLogger.Println(err)
	}
	dst = Stream(bytes)
	return dst
}

func BuildBlock(src1, src2 string) string {
	return "\"" + src1 + "\":\"" + src2 + "\""
}

func Blocks2Stream(src string) Stream {
	src = "{" + src + "}"
	return Stream(src)
}
