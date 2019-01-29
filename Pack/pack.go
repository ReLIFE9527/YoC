package Pack

import (
	. "../Log"
	"errors"
	"strconv"
	"strings"
)

type PacketError struct {
	Obj string
	Op string
	Err error
}
func (e *PacketError) Error() string {
	return e.Obj + " " + e.Op + ": " + e.Err.Error()
}

func PackString(src string) (dst string) {
	dst = "PackHeader//Length:" + strconv.FormatInt(int64(len(src)), 10) + "//" + src + "//PackTail"
	return dst
}

func DePackString(src string)(dst []string,n int,err error) {
	strArr := strings.Split(src, "//")
	for i := 0; i < len(strArr); i++ {
		if strArr[i] == "PackHeader" {
			length, err := getLength(strArr[i+1])
			if err != nil || len(strArr[i+2]) != int(length) {
				Log.Println(err)
				i += 2
				continue
			}else {
				n++
				dst = append(dst, strArr[i+2])
			}
		}
	}
	if n<1 {
		err = &PacketError{src, "DePack", errors.New("can not dePack")}
	}
	return dst, n, err
}

func getLength(str string) (len int64,err error) {
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