package db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"slices"
)

type PostWithUser struct {
	User *User
	Post *Post
}

type UserWithPosts struct {
	*User
	Posts []*Post
}

type PostsSlice struct {
	Posts []*Post
}

func (q *Queries) GetPosts(ctx context.Context, GetPostsFromUserIdParams GetPostsFromUserIdParams) (*PostsSlice, error) {
	posts, err := q.GetPostsFromUserId(ctx, GetPostsFromUserIdParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}

	return &PostsSlice{
		Posts: posts,
	}, nil
}

type QueryData struct {
	Name         string
	ParamName    string
	Params       map[string]string
	IsResult     bool
	IsErr        bool
	ReturnTypes  []string
	ReturnFields map[string]string
}

func (q *Queries) GetPkg() string {
	if q == nil {
		return ""
	}

	return reflect.TypeOf(q).Elem().PkgPath()
}

func (q *Queries) ExtractMethods() map[string]QueryData {
	output := make(map[string]QueryData)
	model := reflect.TypeOf(q)
	ignoredMethods := []string{"WithTx", "ExtractMethods", "GetPkg"}
	for i := range model.NumMethod() {
		method := model.Method(i)
		data := QueryData{
			Params:       make(map[string]string),
			ReturnFields: make(map[string]string),
		}
		if slices.Contains(ignoredMethods, method.Name) {
			continue
		}
		data.Name = method.Name
		if method.Type.NumOut() == 1 {
			data.IsErr = true
			data.ReturnTypes = append(data.ReturnTypes, "error")
		} else {
			firstReturn := method.Type.Out(0)

			if firstReturn == reflect.TypeOf((*sql.Result)(nil)).Elem() {
				data.IsResult = true
				data.ReturnTypes = append(data.ReturnTypes, "sql.Result")
			} else {
				var target reflect.Type
				if firstReturn.Kind() == reflect.Slice {
					target = firstReturn.Elem().Elem()
				} else if firstReturn.Kind() == reflect.Pointer {
					target = firstReturn.Elem()
				}

				if target != nil && target.Kind() == reflect.Struct {
					for i := range target.NumField() {
						field := target.Field(i)
						data.ReturnFields[field.Name] = field.Type.Name()
					}
				}

				data.ReturnTypes = append(data.ReturnTypes, target.Name())
				data.ReturnTypes = append(data.ReturnTypes, "error")
			}
		}

		if method.Type.NumIn() > 2 {
			queryParam := method.Type.In(2)
			data.ParamName = queryParam.Name()
			for i := range queryParam.NumField() {
				field := queryParam.Field(i)
				data.Params[field.Name] = field.Type.Name()
			}
		}
		output[data.Name] = data
	}

	fmt.Printf("DEBUG: %+v\n", output)

	return output
}
