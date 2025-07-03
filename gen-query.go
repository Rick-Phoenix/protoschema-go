package protoschema

import (
	"database/sql"
	"fmt"
	"path"
	"reflect"

	u "github.com/Rick-Phoenix/goutils"
	"github.com/Rick-Phoenix/protoschema/db"
	"github.com/Rick-Phoenix/protoschema/db/sqlgen"
	"github.com/labstack/gommon/log"
	_ "modernc.org/sqlite"
)

type Subquery struct {
	Method          string
	SingleParamName string
	QueryParamName  string
	NoReturn        bool
	Varname         string
	DiscardReturn   bool
}

type SubqueryData struct {
	Method        string
	ParamName     string
	VarName       string
	ReturnType    string
	NoReturn      bool
	DiscardReturn bool
	ParentContext *QueryData
}

type QuerySchema struct {
	Name       string
	Queries    []QueryGroup
	Subqueries []Subquery
	OutType    any
	Store      any
	Package    string
}

type QueryGroup struct {
	IsTx       bool
	Subqueries []Subquery
}

type QueryGroupData struct {
	IsTx       bool
	Subqueries []SubqueryData
}

type QueryData struct {
	Name            string
	FunctionParams  map[string]string
	OutType         string
	OutTypeFields   []string
	Queries         []QueryGroupData
	MakeParamStruct bool
	HasTx           bool
	FuncParamName   string
	FuncParamType   string
}

func (p *ProtoPackage) makeQuery() {
	tmpl := p.tmpl
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	querySchema := QuerySchema{
		Name:    "GetUserWithPosts",
		OutType: &db.PostWithUser{},
		Queries: []QueryGroup{
			{IsTx: true, Subqueries: []Subquery{{Method: "UpdatePost"}, {Method: "UpdateUser", NoReturn: true}}},
			{Subqueries: []Subquery{{Method: "GetUser", SingleParamName: "userId"}}},
		},
		// Subqueries: []Subquery{
		// {Method: "GetUser", QueryParamName: "GetPostsFromUserIdParams.UserId"},
		// {Method: "GetPostsFromUserId"},
		// },
		Store:   sqlgen.New(database),
		Package: "db",
	}

	store := reflect.TypeOf(querySchema.Store)

	if store.Elem().Kind() != reflect.Struct {
		log.Fatalf("Store %q is not a struct.", store.Name())
	}

	queryData := QueryData{Name: querySchema.Name, FunctionParams: make(map[string]string)}

	for _, queryGroup := range querySchema.Queries {

		if len(queryGroup.Subqueries) > 1 {
			queryData.HasTx = true
		}

		queryGroupData := QueryGroupData{IsTx: queryGroup.IsTx}

		for _, subQ := range queryGroup.Subqueries {
			subQData := SubqueryData{Method: subQ.Method, ParentContext: &queryData, VarName: subQ.Varname, NoReturn: subQ.NoReturn, DiscardReturn: subQ.DiscardReturn}
			method, ok := store.MethodByName(subQ.Method)

			if !ok {
				log.Fatalf("Could not find method %q in %q", subQ.Method, store.String())
			}

			if method.Type.NumIn() >= 3 {
				secondParam := method.Type.In(2)
				if secondParam.Kind() == reflect.Struct {
					subQData.ParamName = secondParam.Name()
					queryData.FunctionParams[secondParam.Name()] = getRealPkgPath(secondParam, querySchema.Package)

				} else if subQ.SingleParamName != "" {
					subQData.ParamName = subQ.SingleParamName
					queryData.FunctionParams[subQ.SingleParamName] = secondParam.Name()
				} else if subQ.QueryParamName != "" {
					subQData.ParamName = subQ.QueryParamName
				}
			}

			if len(queryData.FunctionParams) > 1 {
				queryData.MakeParamStruct = true
			}

			if !subQ.NoReturn && method.Type.NumOut() > 0 {
				out := method.Type.Out(0)
				outElem := out.Elem()
				outShortType := outElem.Name()
				outLongType := getRealPkgPath(out, querySchema.Package)
				if out.Kind() == reflect.Slice {
					outShortType = outElem.Elem().Name() + "s"
				}
				outShortLower := u.Uncapitalize(outShortType)
				if subQ.NoReturn {
					subQData.VarName = ""
				} else if subQ.DiscardReturn {
					subQData.VarName = "_"
				} else if subQ.Varname == "" {
					subQData.VarName = outShortLower
				}
				subQData.ReturnType = outLongType
			}

			queryGroupData.Subqueries = append(queryGroupData.Subqueries, subQData)
		}

		queryData.Queries = append(queryData.Queries, queryGroupData)
	}

	outModel := reflect.TypeOf(querySchema.OutType).Elem()

	if outModel.Kind() == reflect.Pointer {
		outModel = outModel.Elem()
	}

	queryData.OutType = getRealPkgPath(outModel, querySchema.Package)

	if outModel.Kind() != reflect.Struct {
		log.Fatalf("Not a struct")
	}

	for i := range outModel.NumField() {
		field := outModel.Field(i)
		queryData.OutTypeFields = append(queryData.OutTypeFields, field.Name)
	}

	if len(queryData.FunctionParams) > 1 {
		queryData.FuncParamName = "params"
		queryData.FuncParamType = queryData.Name + "Params"
	} else {
		for name, typ := range queryData.FunctionParams {
			queryData.FuncParamName = name
			queryData.FuncParamType = typ
		}
	}

	outputPath := "db/tttestquery.go"

	err = u.ExecTemplateAndFormat(tmpl, "multiQuery", outputPath, queryData)
	if err != nil {
		fmt.Print(err)
	}
}

func getRealPkgPath(model reflect.Type, pkg string) string {
	if path.Base(model.PkgPath()) == pkg {
		return model.Name()
	}

	return model.String()
}
