package feedback

import (
	"fmt"
	"testing"
)

var log string = "[FBSDK] makechan: chan=9; elemsize=8; dataqsiz=5\n" +
	"[FBSDK] Chansend: chan=9; elemsize=8; dataqsiz=5; qcount=1; gid=6\n" +
	"[FB] chan: obj=0xc000120000; id=841813590018;\n" +
	"[FBSDK] Chansend: chan=9; elemsize=8; dataqsiz=5; qcount=2; gid=6\n" +
	"[FB] chan: obj=0xc000120000; id=841813590019;\n" +
	"[FBSDK] chanrecv: chan=9; elemsize=8; dataqsiz=5; qcount=1; gid=6\n" +
	"[FB] chan: obj=0xc000120000; id=841813590020;\n" +
	"[FBSDK] chanclose: chan=9; gid=6\n" +
	"[FB] chan: obj=0xc000120000; id=841813590021;\n" +
	"[FB] mutex: obj=1; id=841813590024; locked=1; gid=6\n" +
	"[FB] mutex: obj=1; id=841813590025; locked=0; gid=6\n" +
	"[FB] mutex: obj=1; id=841813590022; locked=1; gid=8\n" +
	"[FB] mutex: obj=1; id=841813590023; locked=0; gid=8\n" +
	"[FBSDK] chanrecv: chan=9; elemsize=8; dataqsiz=5; qcount=0; gid=7\n" +
	"[FB] chan: obj=0xc000120000; id=841813590017;\n" +
	"[FB] mutex: obj=2; id=841813590028; locked=1; gid=6\n" +
	"[FB] mutex: obj=2; id=841813590029; locked=0; gid=6\n"

func TestParseLog(t *testing.T) {
	res, _ := ParseLog(log)
	for k, v := range res {
		print("oid = ", k, "\n")
		for i, o := range v {
			var f string
			if o.status.IsCritical() != 0 {
				f = "\033[1;31;40m[%v] %s\033[0m\n"
			} else {
				f = "[%v] %s\n"
			}
			s := o.ToString()
			fmt.Printf(f, i, s)
		}
		print("\n")
	}
}

func TestLog2Cov(t *testing.T) {
	_, orders := ParseLog(log)
	cov := Log2Cov(orders)
	print(cov.ToString())
}

func TestSRDOAnalysis(t *testing.T) {
	res, _ := ParseLog(log)
	ps := SRDOAnalysis(res)
	for i, p := range ps {
		print("[", i, "] ", p.ToString(), "\n")
	}
}

func TestSODRAnalysis(t *testing.T) {
	res, _ := ParseLog(log)
	ps := SODRAnalysis(res)
	for i, p := range ps {
		print("[", i, "] ", p.ToString(), "\n")
	}
}