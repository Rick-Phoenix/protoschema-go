package protoschema

import (
	"database/sql"
	"fmt"
	"reflect"

	u "github.com/Rick-Phoenix/goutils"
	"github.com/Rick-Phoenix/protoschema/db"
	"github.com/Rick-Phoenix/protoschema/db/sqlgen"
	"github.com/labstack/gommon/log"
	_ "modernc.org/sqlite"
)

type Subquery struct {
	Method    string
	ParamName string
}

type SubqueryData struct {
	Method     string
	ParamName  string
	VarName    string
	ReturnType string
}

type Value struct {
	PkgPath string
}

type QuerySchema struct {
	Name       string
	Subqueries []Subquery
	OutType    any
}

type QueryData struct {
	Name            string
	FunctionParams  map[string]string
	OutType         string
	OutTypeFields   []string
	Subqueries      []SubqueryData
	MakeParamStruct bool
	FuncParamName   string
	FuncParamType   string
}

func (p *ProtoPackage) makeQuery() {
	tmpl := p.tmpl
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	qmap := make(map[string]string)
	qmap["GetUser"] = "userId"

	val := reflect.TypeOf(sqlgen.New(database))

	queryData := QueryData{Name: "GetUserWithPosts", FunctionParams: make(map[string]string)}

	subqueries := []Subquery{{"GetUser", "userId"}, {"GetPostsFromUserId", ""}}

	subQueriesData := []SubqueryData{}

	for _, subQ := range subqueries {
		subQData := SubqueryData{Method: subQ.Method}
		method, _ := val.MethodByName(subQ.Method)

		// fmt.Printf("Method Name: %+v\n", method.Name)
		if method.Type.NumIn() >= 3 {
			secondParam := method.Type.In(2)
			if secondParam.Kind() == reflect.Struct {
				subQData.ParamName = secondParam.Name()
				queryData.FunctionParams[secondParam.Name()] = secondParam.String()
			} else {
				if singleParamName, ok := qmap[method.Name]; ok {
					queryData.FunctionParams[singleParamName] = secondParam.Name()
					subQData.ParamName = singleParamName
				}
			}
		}

		if method.Type.NumOut() > 0 {
			out := method.Type.Out(0)
			outElem := out.Elem()
			outShortType := outElem.Name()
			outLongType := out.String()
			// outPkgPath := outElem.PkgPath()
			if out.Kind() == reflect.Slice {
				outShortType = outElem.Elem().Name() + "s"
				// outPkgPath = outElem.Elem().PkgPath()
			}
			outShortLower := u.Uncapitalize(outShortType)
			subQData.VarName = outShortLower
			subQData.ReturnType = outLongType
		}

		subQueriesData = append(subQueriesData, subQData)
	}

	queryData.Subqueries = subQueriesData

	outType := []db.UserWithPosts{}

	outModel := reflect.TypeOf(outType).Elem()

	if outModel.Kind() == reflect.Pointer {
		outModel = outModel.Elem()
	}

	queryData.OutType = reflect.TypeOf(outType).String()

	if outModel.Kind() != reflect.Struct {
		log.Fatalf("Not a struct")
	}

	for i := range outModel.NumField() {
		field := outModel.Field(i)
		queryData.OutTypeFields = append(queryData.OutTypeFields, field.Name)
	}

	if len(queryData.FunctionParams) > 1 {
		queryData.MakeParamStruct = true
		queryData.FuncParamName = "params"
		queryData.FuncParamType = queryData.Name + "Params"
	} else {
		for name, typ := range queryData.FunctionParams {
			queryData.FuncParamName = name
			queryData.FuncParamType = typ
		}
	}

	// fmt.Printf("DEBUG: %+v\n", queryData)

	outputPath := "gen/tttestquery.go"

	err = u.ExecTemplateAndFormat(tmpl, "multiQuery", outputPath, queryData)
	if err != nil {
		fmt.Print(err)
	}
}
