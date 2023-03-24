package data_analysis

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"log"
	"os"
	"strconv"
	"strings"
)

type tRes struct {
	name string
	cnt  int
	pass string
	raw  string
}

func filename(s string) string {
	a := strings.LastIndex(s, "/")
	b := strings.LastIndex(s, ".")
	if a != -1 && b != -1 && a+1 < b {
		return s[a+1 : b]
	}
	return ""
}

func strip(fn string) string {
	if strings.HasSuffix(fn, "_test") {
		return fn[0 : len(fn)-len("_test")]
	}
	return fn
}

func readfile(path string) ([]string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	res := strings.Split(string(bytes), "\n")
	return res, nil
}

func GenGraph() {
	bA, err := readfile("script/data_analysis/baselineA.txt")
	bB, err2 := readfile("script/data_analysis/baselineB.txt")
	bL, err3 := readfile("script/data_analysis/lite.txt")
	if err != nil || err2 != nil || err3 != nil {
		log.Fatalf("Error")
	}

	parseline := func(l string) *tRes {
		res := tRes{}
		res.raw = l
		ss := strings.Split(l, " ")
		fss := make([]string, 0)
		for _, s := range ss {
			if s != "" {
				fss = append(fss, s)
			}
		}
		if len(fss) != 3 {
			return nil
		}
		res.name = fss[0]
		res.pass = fss[1]
		cnt, _ := strconv.ParseInt(fss[2][0:len(fss[2])-1], 10, 32)
		res.cnt = int(cnt)
		return &res
	}

	parse := func(lines []string) map[string]*tRes {
		res := make(map[string]*tRes, 0)
		for _, line := range lines {
			t := parseline(line)
			if t != nil {
				res[t.name] = t
			}
		}
		return res
	}

	res1 := parse(bA)
	res2 := parse(bB)
	res3 := parse(bL)
	axis := make([]string, 0)
	item1 := make([]opts.BarData, 0)
	item2 := make([]opts.BarData, 0)
	item3 := make([]opts.BarData, 0)
	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithYAxisOpts(opts.YAxis{Type: "log"}))
	for k, v := range res1 {
		if v2, ok := res2[k]; ok {
			if v3, ok := res3[k]; ok {
				axis = append(axis, k)
				item1 = append(item1, opts.BarData{Value: v.cnt})
				item2 = append(item2, opts.BarData{Value: v2.cnt})
				item3 = append(item3, opts.BarData{Value: v3.cnt})
			}
		}
	}
	t := bar.SetXAxis(axis)
	// t.AddSeries("A", item1)
	t.AddSeries("B", item2)
	t.AddSeries("L", item3)
	f, _ := os.Create("script/data_analysis/bar.html")
	bar.Render(f)
}
