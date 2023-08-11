package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"net/url"
	"text/template"
	"time"
)

//go:embed tpl.gohtml
var genOrm string

func gen(w io.Writer, srcFile string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, srcFile, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	s := &SingleFileEntryVisitor{}
	ast.Walk(s, f)
	file := s.Get()
	tpl := template.New("gen-orm")
	tpl, err = tpl.Parse(genOrm)
	if err != nil {
		return err
	}
	return tpl.Execute(w, Data{
		File: file,
		Ops:  []string{"LT", "GT", "EQ"},
	})
}

type Data struct {
	*File
	Ops []string
}

func main() {
	//src := os.Args[1]
	//dstDir := filepath.Dir(src)
	//fileName := filepath.Base(src)
	//idx := strings.LastIndexByte(fileName, '.')
	//dst := filepath.Join(dstDir, fileName[:idx]+".gen.go")
	//f, err := os.Create(dst)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//err = gen(f, src)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println("生成成功")
	aa, er := updateRespTime("http://localhost:9090", "geekbang_web_http_response{method=\"GET\",quantile=\"0.5\"}")
	fmt.Println(aa.Data.Result[0].Value[1])
	fmt.Println(er)
}

func updateRespTime(endpoint, query string) (*QueryInfo, error) {
	info := &QueryInfo{}
	ustr := endpoint + "/api/v1/query?query=" + query
	u, err := url.Parse(ustr)
	if err != nil {
		return info, err
	}
	u.RawQuery = u.Query().Encode()
	//http://your_prometheus.com/api/v1/query?query=xxx
	err = GetPromResult(u.String(), &info)
	if err != nil {
		return info, err
	}
	fmt.Println(info.Data.Result[0].Metric)
	return info, nil
}

func GetPromResult(url string, result interface{}) error {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	r, err := httpClient.Get(url)
	if err != nil {
		return err
	}

	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(result)
	if err != nil {
		return err
	}
	return nil
}

type ResultType struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}

type QueryData struct {
	ResultType string       `json:"resultType"`
	Result     []ResultType `json:"result"`
}

type QueryInfo struct {
	Status string    `json:"status"`
	Data   QueryData `json:"data"`
}
