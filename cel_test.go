package schemabuilder_test

// func TestCheckCelExpression(t *testing.T) {
// 	env, err := cel.NewEnv(
// 		cel.Variable("value", cel.AnyType),
// 		cel.Variable("name", cel.StringType),
// 	)
// 	if err != nil {
// 		fmt.Printf("Error creating CEL environment: %v\n", err)
// 		return
// 	}
//
// 	expressions := []string{
// 		"value > 0",
// 		"name.startsWith('A')",
// 		"size(name) > 5 && value != null",
// 	}
//
// 	for _, expr := range expressions {
// 		fmt.Printf("\nExpression: \"%s\"\n", expr)
//
// 		ast, issues := env.Parse(expr)
// 		if issues != nil && issues.Err() != nil {
// 			fmt.Printf("  ❌ Parsing Error: %v\n", issues.Err())
// 			continue
// 		}
//
// 		checkedAST, issues := env.Check(ast)
// 		if issues != nil && issues.Err() != nil {
// 			fmt.Printf("  ❌ Compilation (Semantic) Error: %v\n", issues.Err())
// 			continue
// 		}
//
// 		parsedExpr, err := cel.AstToParsedExpr(checkedAST)
// 		if err != nil {
// 			fmt.Printf("Errors in parsedExpr: %v\n", err)
// 			continue
// 		}
//
// 		stringExpr := parsedExpr.String()
//
// 		fmt.Printf("DEBUG: %+v\n", stringExpr)
// 	}
// }
