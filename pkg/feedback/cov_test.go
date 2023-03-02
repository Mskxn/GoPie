package feedback

import (
	"fmt"
	"testing"
	"toolkit/pkg/testdata"
)

func TestLog2Cov(t *testing.T) {
	_, orders := ParseLog(log)
	cov := Log2Cov(orders)
	print(cov.ToString())
}

func TestCovUpdate(t *testing.T) {
	_, orders := ParseLog(testdata.Log)
	c := Log2Cov(orders)
	fmt.Print(c.ToString())

	c.UpdateC(OpID(841813590024), ToStatus(ChanFull))
	c.UpdateC(OpID(1337), ToStatus(ChanEmpty))

	c.UpdateO(1, 2)

	c.UpdateT(OpID(841813590024), Chansend)
	fmt.Print(c.ToString())
}
