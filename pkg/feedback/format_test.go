package feedback

import (
	"strconv"
	"testing"
)

var log string = "[FBSDK] makechan: chan=0xc000078090; elemsize=8; dataqsiz=5\n" +
	"[FBSDK] chansend: chan=0xc000078090; elemsize=8; dataqsiz=5; qcount=1\n" +
	"[FB] chan: obj=0xc000078090; id=841813590017;\n" +
	"[FBSDK] chanclose: chan=0xc000078090\n" +
	"[FB] chan: obj=0xc000078090; id=841813590018;\n" +
	"[FB] mutex: obj=0xc000010390; id=841813590020; locked=0\n" +
	"[FB] mutex: obj=0xc000010390; id=841813590019; locked=1\n"

func TestParseLog(t *testing.T) {
	res := ParseLog(log)
	for k, v := range res {
		print("k= ", strconv.FormatUint(k, 16), "\tv= ", v, "\n")
	}
}
