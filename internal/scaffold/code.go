package scaffold

import (
	"context"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cimoing/gos/internal/filesystem"
	"github.com/cimoing/gos/internal/generator"
	"github.com/cimoing/gos/internal/naming"
)

type CodeGenerator struct {
	engine *generator.Engine
	now    func() time.Time
}

type MakeUsecaseOptions struct {
	TargetDir string
	Name      string
	Force     bool
	DryRun    bool
}

type MakeMigrationOptions struct {
	TargetDir string
	Name      string
	Dir       string
	Force     bool
	DryRun    bool
}

type MakeHandlerOptions struct {
	TargetDir  string
	Name       string
	ModulePath string
	Register   bool
	OpenAPI    bool
	Force      bool
	DryRun     bool
}

type MakeRepositoryOptions struct {
	TargetDir     string
	Name          string
	ModulePath    string
	DB            string
	TableName     string
	Fields        string
	WithMigration bool
	MigrationDir  string
	Register      bool
	OpenAPI       bool
	Force         bool
	DryRun        bool
}

type MakeModelOptions struct {
	TargetDir string
	Name      string
	Fields    string
	OpenAPI   bool
	Force     bool
	DryRun    bool
}

type MakeTestOptions struct {
	TargetDir  string
	Kind       string
	Name       string
	ModulePath string
	Force      bool
	DryRun     bool
}

type MakeCommandOptions struct {
	TargetDir  string
	Name       string
	ModulePath string
	Register   bool
	Force      bool
	DryRun     bool
}

func NewCodeGenerator(engine *generator.Engine) *CodeGenerator {
	return &CodeGenerator{
		engine: engine,
		now:    time.Now,
	}
}

func (g *CodeGenerator) GenerateUsecase(ctx context.Context, opts MakeUsecaseOptions) (*filesystem.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	module, action, err := parseUsecaseName(opts.Name)
	if err != nil {
		return nil, err
	}

	data := usecaseTemplateData{
		PackageName: naming.ToSnake(module),
		TypeName:    naming.ToPascal(action),
		ActionSnake: naming.ToSnake(action),
	}

	source, err := g.engine.Templates.RenderString("usecase.go", usecaseGoTemplate, data)
	if err != nil {
		return nil, err
	}
	testSource, err := g.engine.Templates.RenderString("usecase_test.go", usecaseTestTemplate, data)
	if err != nil {
		return nil, err
	}

	moduleDir := naming.ToSnake(module)
	files := []filesystem.File{
		{
			Path:    filepath.ToSlash(filepath.Join("internal", "usecase", moduleDir, data.ActionSnake+".go")),
			Content: source,
		},
		{
			Path:    filepath.ToSlash(filepath.Join("internal", "usecase", moduleDir, data.ActionSnake+"_test.go")),
			Content: testSource,
		},
	}

	return g.engine.Writer.Write(files, filesystem.WriteOptions{
		Root:   opts.TargetDir,
		Force:  opts.Force,
		DryRun: opts.DryRun,
	})
}

func (g *CodeGenerator) GenerateMigration(ctx context.Context, opts MakeMigrationOptions) (*filesystem.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	name := naming.ToSnake(opts.Name)
	if name == "" {
		return nil, fmt.Errorf("migration name is required")
	}
	if opts.Dir == "" {
		opts.Dir = "migrations"
	}

	timestamp := g.now().Format("20060102150405")
	base := timestamp + "_" + name
	files := []filesystem.File{
		{
			Path:    filepath.ToSlash(filepath.Join(opts.Dir, base+".up.sql")),
			Content: []byte("-- Write your migration SQL here.\n"),
		},
		{
			Path:    filepath.ToSlash(filepath.Join(opts.Dir, base+".down.sql")),
			Content: []byte("-- Write your rollback SQL here.\n"),
		},
	}

	return g.engine.Writer.Write(files, filesystem.WriteOptions{
		Root:   opts.TargetDir,
		Force:  opts.Force,
		DryRun: opts.DryRun,
	})
}

func (g *CodeGenerator) GenerateHandler(ctx context.Context, opts MakeHandlerOptions) (*filesystem.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(opts.ModulePath) == "" {
		return nil, fmt.Errorf("module path is required")
	}

	module := naming.ToSnake(opts.Name)
	if module == "" {
		return nil, fmt.Errorf("handler module is required")
	}

	data := handlerTemplateData{
		ModulePath:   opts.ModulePath,
		PackageName:  "handler",
		TypeName:     naming.ToPascal(module),
		VariableName: naming.ToCamel(module) + "Handler",
		RoutePath:    "/" + naming.ToKebab(module) + "s",
	}

	source, err := g.engine.Templates.RenderString("handler.go", handlerGoTemplate, data)
	if err != nil {
		return nil, err
	}
	testSource, err := g.engine.Templates.RenderString("handler_test.go", handlerTestTemplate, data)
	if err != nil {
		return nil, err
	}

	files := []filesystem.File{
		{
			Path:    filepath.ToSlash(filepath.Join("internal", "interfaces", "http", "handler", module+"_handler.go")),
			Content: source,
		},
		{
			Path:    filepath.ToSlash(filepath.Join("internal", "interfaces", "http", "handler", module+"_handler_test.go")),
			Content: testSource,
		},
	}

	var skipped []string
	if opts.Register {
		routerSource, changed, err := registerHandlerInRouter(opts.TargetDir, data)
		if err != nil {
			skipped = append(skipped, filepath.ToSlash(filepath.Join("internal", "interfaces", "http", "router.go")))
		} else if changed {
			files = append(files, filesystem.File{
				Path:      filepath.ToSlash(filepath.Join("internal", "interfaces", "http", "router.go")),
				Content:   routerSource,
				Overwrite: true,
			})
		}
	}
	if opts.OpenAPI {
		openAPISource, changed, err := registerHandlerInOpenAPI(opts.TargetDir, data)
		if err != nil {
			skipped = append(skipped, filepath.ToSlash(filepath.Join("api", "openapi.yaml")))
		} else if changed {
			files = append(files, filesystem.File{
				Path:      filepath.ToSlash(filepath.Join("api", "openapi.yaml")),
				Content:   openAPISource,
				Overwrite: true,
			})
		}
	}

	result, err := g.engine.Writer.Write(files, filesystem.WriteOptions{
		Root:   opts.TargetDir,
		Force:  opts.Force,
		DryRun: opts.DryRun,
	})
	if err != nil {
		return nil, err
	}
	result.Skipped = append(result.Skipped, skipped...)
	return result, nil
}

func (g *CodeGenerator) GenerateRepository(ctx context.Context, opts MakeRepositoryOptions) (*filesystem.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(opts.ModulePath) == "" {
		return nil, fmt.Errorf("module path is required")
	}
	if opts.DB == "" {
		opts.DB = "mysql"
	}
	if opts.DB != "mysql" {
		return nil, fmt.Errorf("unsupported repository db %q", opts.DB)
	}

	module := naming.ToSnake(opts.Name)
	if module == "" {
		return nil, fmt.Errorf("repository module is required")
	}
	tableName := opts.TableName
	if tableName == "" {
		tableName = pluralTableName(module)
	}
	if !isSQLIdentifier(tableName) {
		return nil, fmt.Errorf("invalid table name %q", tableName)
	}
	fields, err := parseFields(opts.Fields)
	if err != nil {
		return nil, err
	}

	data := repositoryTemplateData{
		ModulePath:              opts.ModulePath,
		PackageName:             module,
		TypeName:                naming.ToPascal(module),
		ReceiverName:            naming.ToCamel(module),
		FieldName:               naming.ToPascal(module) + "Repository",
		VariableName:            naming.ToCamel(module) + "Repository",
		TableName:               tableName,
		Fields:                  fields,
		HasFields:               len(fields) > 0,
		EntityImports:           entityImports(fields),
		SelectColumns:           selectColumns(fields),
		ScanTargets:             scanTargets(naming.ToCamel(module), fields),
		InsertColumns:           insertColumns(fields),
		InsertValues:            insertValues(fields),
		InsertArgs:              insertArgs(naming.ToCamel(module), fields),
		UpdateSet:               updateSet(fields),
		UpdateArgs:              updateArgs(naming.ToCamel(module), fields),
		MigrationColumns:        migrationColumns(fields),
		IntegrationImports:      integrationImports(fields),
		IntegrationEntityFields: integrationEntityFields(fields),
		IntegrationAssertions:   integrationAssertions(naming.ToCamel(module), fields),
	}

	entitySource, err := g.engine.Templates.RenderString("entity.go", domainEntityTemplate, data)
	if err != nil {
		return nil, err
	}
	contractSource, err := g.engine.Templates.RenderString("repository.go", domainRepositoryTemplate, data)
	if err != nil {
		return nil, err
	}
	repositorySource, err := g.engine.Templates.RenderString("mysql_repository.go", mysqlRepositoryTemplate, data)
	if err != nil {
		return nil, err
	}
	testSource, err := g.engine.Templates.RenderString("mysql_repository_test.go", mysqlRepositoryTestTemplate, data)
	if err != nil {
		return nil, err
	}
	integrationTestSource, err := g.engine.Templates.RenderString("mysql_repository_integration_test.go", mysqlRepositoryIntegrationTestTemplate, data)
	if err != nil {
		return nil, err
	}

	files := []filesystem.File{
		{
			Path:         filepath.ToSlash(filepath.Join("internal", "domain", module, "entity.go")),
			Content:      entitySource,
			SkipIfExists: true,
		},
		{
			Path:         filepath.ToSlash(filepath.Join("internal", "domain", module, "repository.go")),
			Content:      contractSource,
			SkipIfExists: true,
		},
		{
			Path:    filepath.ToSlash(filepath.Join("internal", "infrastructure", "persistence", "mysql", module+"_repository.go")),
			Content: repositorySource,
		},
		{
			Path:    filepath.ToSlash(filepath.Join("internal", "infrastructure", "persistence", "mysql", module+"_repository_test.go")),
			Content: testSource,
		},
		{
			Path:    filepath.ToSlash(filepath.Join("internal", "infrastructure", "persistence", "mysql", module+"_repository_integration_test.go")),
			Content: integrationTestSource,
		},
	}

	if opts.WithMigration {
		if opts.MigrationDir == "" {
			opts.MigrationDir = "migrations"
		}
		timestamp := g.now().Format("20060102150405")
		migrationBase := timestamp + "_create_" + tableName + "_table"
		files = append(files,
			filesystem.File{
				Path:    filepath.ToSlash(filepath.Join(opts.MigrationDir, migrationBase+".up.sql")),
				Content: []byte(fmt.Sprintf("CREATE TABLE %s (\n    id BIGINT PRIMARY KEY AUTO_INCREMENT%s\n);\n", tableName, data.MigrationColumns)),
			},
			filesystem.File{
				Path:    filepath.ToSlash(filepath.Join(opts.MigrationDir, migrationBase+".down.sql")),
				Content: []byte(fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", tableName)),
			},
		)
	}

	var skipped []string
	if opts.Register {
		assemblySource, changed, err := registerRepositoryInAssembly(opts.TargetDir, data)
		if err != nil {
			skipped = append(skipped, filepath.ToSlash(filepath.Join("internal", "app", "assembly.go")))
		} else if changed {
			files = append(files, filesystem.File{
				Path:      filepath.ToSlash(filepath.Join("internal", "app", "assembly.go")),
				Content:   assemblySource,
				Overwrite: true,
			})
		}
	}
	if opts.OpenAPI {
		openAPISource, changed, err := registerSchemaInOpenAPI(opts.TargetDir, data.TypeName, fields)
		if err != nil {
			skipped = append(skipped, filepath.ToSlash(filepath.Join("api", "openapi.yaml")))
		} else if changed {
			files = append(files, filesystem.File{
				Path:      filepath.ToSlash(filepath.Join("api", "openapi.yaml")),
				Content:   openAPISource,
				Overwrite: true,
			})
		}
	}

	result, err := g.engine.Writer.Write(files, filesystem.WriteOptions{
		Root:   opts.TargetDir,
		Force:  opts.Force,
		DryRun: opts.DryRun,
	})
	if err != nil {
		return nil, err
	}
	result.Skipped = append(result.Skipped, skipped...)
	return result, nil
}

func (g *CodeGenerator) GenerateModel(ctx context.Context, opts MakeModelOptions) (*filesystem.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	module := naming.ToSnake(opts.Name)
	if module == "" {
		return nil, fmt.Errorf("model module is required")
	}
	fields, err := parseFields(opts.Fields)
	if err != nil {
		return nil, err
	}

	data := repositoryTemplateData{
		PackageName:   module,
		TypeName:      naming.ToPascal(module),
		Fields:        fields,
		EntityImports: entityImports(fields),
	}

	entitySource, err := g.engine.Templates.RenderString("entity.go", domainEntityTemplate, data)
	if err != nil {
		return nil, err
	}

	files := []filesystem.File{
		{
			Path:    filepath.ToSlash(filepath.Join("internal", "domain", module, "entity.go")),
			Content: entitySource,
		},
	}

	var skipped []string
	if opts.OpenAPI {
		openAPISource, changed, err := registerSchemaInOpenAPI(opts.TargetDir, data.TypeName, fields)
		if err != nil {
			skipped = append(skipped, filepath.ToSlash(filepath.Join("api", "openapi.yaml")))
		} else if changed {
			files = append(files, filesystem.File{
				Path:      filepath.ToSlash(filepath.Join("api", "openapi.yaml")),
				Content:   openAPISource,
				Overwrite: true,
			})
		}
	}

	result, err := g.engine.Writer.Write(files, filesystem.WriteOptions{
		Root:   opts.TargetDir,
		Force:  opts.Force,
		DryRun: opts.DryRun,
	})
	if err != nil {
		return nil, err
	}
	result.Skipped = append(result.Skipped, skipped...)
	return result, nil
}

func (g *CodeGenerator) GenerateTest(ctx context.Context, opts MakeTestOptions) (*filesystem.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(opts.ModulePath) == "" {
		return nil, fmt.Errorf("module path is required")
	}

	var file filesystem.File
	var err error
	switch opts.Kind {
	case "usecase":
		file, err = g.makeUsecaseTestFile(opts.Name)
	case "handler":
		file, err = g.makeHandlerTestFile(opts.Name)
	case "repository":
		file, err = g.makeRepositoryTestFile(opts.Name)
	default:
		return nil, fmt.Errorf("unsupported test kind %q", opts.Kind)
	}
	if err != nil {
		return nil, err
	}

	return g.engine.Writer.Write([]filesystem.File{file}, filesystem.WriteOptions{
		Root:   opts.TargetDir,
		Force:  opts.Force,
		DryRun: opts.DryRun,
	})
}

func (g *CodeGenerator) GenerateCommand(ctx context.Context, opts MakeCommandOptions) (*filesystem.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(opts.ModulePath) == "" {
		return nil, fmt.Errorf("module path is required")
	}

	commandName := naming.ToKebab(opts.Name)
	if commandName == "" {
		return nil, fmt.Errorf("command name is required")
	}
	switch commandName {
	case "serve", "schedule", "queue", "help":
		return nil, fmt.Errorf("command name %q is reserved", commandName)
	}

	data := commandTemplateData{
		ModulePath:   opts.ModulePath,
		PackageName:  "command",
		CommandName:  commandName,
		FunctionName: "New" + naming.ToPascal(opts.Name) + "Command",
	}

	source, err := g.engine.Templates.RenderString("command.go", commandGoTemplate, data)
	if err != nil {
		return nil, err
	}
	testSource, err := g.engine.Templates.RenderString("command_test.go", commandTestTemplate, data)
	if err != nil {
		return nil, err
	}

	fileName := naming.ToSnake(opts.Name)
	files := []filesystem.File{
		{
			Path:    filepath.ToSlash(filepath.Join("internal", "command", fileName+".go")),
			Content: source,
		},
		{
			Path:    filepath.ToSlash(filepath.Join("internal", "command", fileName+"_test.go")),
			Content: testSource,
		},
	}

	var skipped []string
	if opts.Register {
		mainSource, changed, err := registerCommandInMain(opts.TargetDir, data)
		if err != nil {
			skipped = append(skipped, filepath.ToSlash(filepath.Join("cmd", "api", "main.go")))
		} else if changed {
			files = append(files, filesystem.File{
				Path:      filepath.ToSlash(filepath.Join("cmd", "api", "main.go")),
				Content:   mainSource,
				Overwrite: true,
			})
		}
	}

	result, err := g.engine.Writer.Write(files, filesystem.WriteOptions{
		Root:   opts.TargetDir,
		Force:  opts.Force,
		DryRun: opts.DryRun,
	})
	if err != nil {
		return nil, err
	}
	result.Skipped = append(result.Skipped, skipped...)
	return result, nil
}

func (g *CodeGenerator) makeUsecaseTestFile(name string) (filesystem.File, error) {
	module, action, err := parseUsecaseName(name)
	if err != nil {
		return filesystem.File{}, err
	}

	data := usecaseTemplateData{
		PackageName: naming.ToSnake(module),
		TypeName:    naming.ToPascal(action),
		ActionSnake: naming.ToSnake(action),
	}

	source, err := g.engine.Templates.RenderString("usecase_test.go", usecaseTestTemplate, data)
	if err != nil {
		return filesystem.File{}, err
	}

	moduleDir := naming.ToSnake(module)
	return filesystem.File{
		Path:    filepath.ToSlash(filepath.Join("internal", "usecase", moduleDir, data.ActionSnake+"_test.go")),
		Content: source,
	}, nil
}

func (g *CodeGenerator) makeHandlerTestFile(name string) (filesystem.File, error) {
	module := naming.ToSnake(name)
	if module == "" {
		return filesystem.File{}, fmt.Errorf("handler module is required")
	}

	data := handlerTemplateData{
		PackageName: "handler",
		TypeName:    naming.ToPascal(module),
		RoutePath:   "/" + naming.ToKebab(module) + "s",
	}

	source, err := g.engine.Templates.RenderString("handler_test.go", handlerTestTemplate, data)
	if err != nil {
		return filesystem.File{}, err
	}

	return filesystem.File{
		Path:    filepath.ToSlash(filepath.Join("internal", "interfaces", "http", "handler", module+"_handler_test.go")),
		Content: source,
	}, nil
}

func (g *CodeGenerator) makeRepositoryTestFile(name string) (filesystem.File, error) {
	module := naming.ToSnake(name)
	if module == "" {
		return filesystem.File{}, fmt.Errorf("repository module is required")
	}

	data := repositoryTemplateData{
		TypeName: naming.ToPascal(module),
	}

	source, err := g.engine.Templates.RenderString("mysql_repository_test.go", mysqlRepositoryTestTemplate, data)
	if err != nil {
		return filesystem.File{}, err
	}

	return filesystem.File{
		Path:    filepath.ToSlash(filepath.Join("internal", "infrastructure", "persistence", "mysql", module+"_repository_test.go")),
		Content: source,
	}, nil
}

func parseUsecaseName(name string) (string, string, error) {
	parts := strings.Split(strings.Trim(name, "/"), "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("usecase name must be <module>/<action>, got %q", name)
	}
	return parts[0], parts[1], nil
}

func pluralTableName(name string) string {
	if strings.HasSuffix(name, "y") && len(name) > 1 {
		return strings.TrimSuffix(name, "y") + "ies"
	}
	if strings.HasSuffix(name, "s") {
		return name + "es"
	}
	return name + "s"
}

func isSQLIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for i, r := range value {
		if r == '_' || unicode.IsLetter(r) || (i > 0 && unicode.IsDigit(r)) {
			continue
		}
		return false
	}
	return true
}

func parseFields(value string) ([]fieldSpec, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parts, err := splitFieldSpecs(value)
	if err != nil {
		return nil, err
	}
	fields := make([]fieldSpec, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		fieldParts := strings.SplitN(strings.TrimSpace(part), ":", 3)
		if len(fieldParts) < 2 {
			return nil, fmt.Errorf("field must be name:type, got %q", part)
		}
		name := naming.ToSnake(fieldParts[0])
		typ := strings.ToLower(strings.TrimSpace(fieldParts[1]))
		if name == "" || !isSQLIdentifier(name) {
			return nil, fmt.Errorf("invalid field name %q", name)
		}
		if name == "id" {
			return nil, fmt.Errorf("field name %q is reserved", name)
		}
		if seen[name] {
			return nil, fmt.Errorf("duplicate field name %q", name)
		}
		seen[name] = true

		field, err := newFieldSpec(name, typ)
		if err != nil {
			return nil, err
		}
		if len(fieldParts) == 3 {
			if err := applyFieldOptions(&field, fieldParts[2]); err != nil {
				return nil, err
			}
		}
		finalizeFieldSpec(&field)
		if err := validateFieldSpec(field); err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func splitFieldSpecs(value string) ([]string, error) {
	rawParts := strings.Split(value, ",")
	fields := make([]string, 0, len(rawParts))
	for _, rawPart := range rawParts {
		part := strings.TrimSpace(rawPart)
		if part == "" {
			continue
		}
		if strings.Contains(part, ":") {
			fields = append(fields, part)
			continue
		}
		if len(fields) == 0 {
			return nil, fmt.Errorf("field option %q has no field", part)
		}
		fields[len(fields)-1] += "," + part
	}
	return fields, nil
}

func newFieldSpec(name string, typ string) (fieldSpec, error) {
	field := fieldSpec{
		Name:   name,
		GoName: naming.ToPascal(name),
		Column: name,
	}

	switch typ {
	case "string":
		field.GoType = "string"
		field.SQLType = "VARCHAR(255)"
		field.ZeroSQL = "''"
	case "int":
		field.GoType = "int"
		field.SQLType = "INT"
		field.ZeroSQL = "0"
	case "int64":
		field.GoType = "int64"
		field.SQLType = "BIGINT"
		field.ZeroSQL = "0"
	case "bool":
		field.GoType = "bool"
		field.SQLType = "BOOLEAN"
		field.ZeroSQL = "FALSE"
	case "time":
		field.GoType = "time.Time"
		field.SQLType = "TIMESTAMP"
		field.ZeroSQL = "CURRENT_TIMESTAMP"
		field.NeedsTime = true
	default:
		return fieldSpec{}, fmt.Errorf("unsupported field type %q", typ)
	}
	return field, nil
}

func applyFieldOptions(field *fieldSpec, value string) error {
	options, err := splitFieldOptions(value)
	if err != nil {
		return err
	}
	for _, option := range options {
		option = strings.TrimSpace(option)
		if option == "" {
			continue
		}

		switch {
		case option == "required":
			field.Nullable = false
		case option == "nullable":
			field.Nullable = true
		case option == "unique":
			field.Unique = true
		case option == "index":
			field.Index = true
		case strings.HasPrefix(option, "size="):
			sizeText := strings.TrimPrefix(option, "size=")
			size, err := strconv.Atoi(sizeText)
			if err != nil || size <= 0 {
				return fmt.Errorf("invalid size option %q for field %s", option, field.Name)
			}
			if field.GoType != "string" {
				return fmt.Errorf("size option is only supported for string fields")
			}
			field.SQLType = fmt.Sprintf("VARCHAR(%d)", size)
		case strings.HasPrefix(option, "default="):
			defaultSQL, err := defaultSQLForField(field, strings.TrimPrefix(option, "default="))
			if err != nil {
				return err
			}
			field.DefaultSQL = defaultSQL
			field.HasDefault = true
		case strings.HasPrefix(option, "sql="):
			sqlType := strings.TrimSpace(strings.TrimPrefix(option, "sql="))
			if !isSafeSQLType(sqlType) {
				return fmt.Errorf("invalid sql option %q for field %s", option, field.Name)
			}
			field.SQLType = sqlType
		case strings.HasPrefix(option, "json="):
			jsonName := strings.TrimSpace(strings.TrimPrefix(option, "json="))
			if !isJSONFieldName(jsonName) {
				return fmt.Errorf("invalid json option %q for field %s", option, field.Name)
			}
			field.JSONName = jsonName
		default:
			return fmt.Errorf("unsupported field option %q", option)
		}
	}
	return nil
}

func splitFieldOptions(value string) ([]string, error) {
	var options []string
	var current strings.Builder
	depth := 0
	for _, r := range value {
		switch r {
		case '(':
			depth++
		case ')':
			if depth == 0 {
				return nil, fmt.Errorf("invalid field option list %q", value)
			}
			depth--
		case ',':
			if depth == 0 {
				options = append(options, current.String())
				current.Reset()
				continue
			}
		}
		current.WriteRune(r)
	}
	if depth != 0 {
		return nil, fmt.Errorf("invalid field option list %q", value)
	}
	options = append(options, current.String())
	return options, nil
}

func finalizeFieldSpec(field *fieldSpec) {
	if field.JSONName == "" {
		field.JSONName = field.Name
	}
	field.Tag = fmt.Sprintf(" `json:%q`", field.JSONName)
}

func validateFieldSpec(field fieldSpec) error {
	if field.HasDefault && field.DefaultSQL == "NULL" && !field.Nullable {
		return fmt.Errorf("field %s uses default=null but is not nullable", field.Name)
	}
	return nil
}

func defaultSQLForField(field *fieldSpec, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("default option for field %s cannot be empty", field.Name)
	}
	if strings.EqualFold(value, "null") {
		return "NULL", nil
	}

	switch field.GoType {
	case "string":
		return "'" + strings.ReplaceAll(value, "'", "''") + "'", nil
	case "int", "int64":
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return "", fmt.Errorf("invalid numeric default %q for field %s", value, field.Name)
		}
		return value, nil
	case "bool":
		switch strings.ToLower(value) {
		case "true":
			return "TRUE", nil
		case "false":
			return "FALSE", nil
		default:
			return "", fmt.Errorf("invalid bool default %q for field %s", value, field.Name)
		}
	case "time.Time":
		if strings.EqualFold(value, "now") || strings.EqualFold(value, "current_timestamp") {
			return "CURRENT_TIMESTAMP", nil
		}
		return "", fmt.Errorf("time default for field %s must be now or CURRENT_TIMESTAMP", field.Name)
	default:
		return "", fmt.Errorf("unsupported default field type %s", field.GoType)
	}
}

func isSafeSQLType(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		switch r {
		case '_', ' ', '(', ')', ',':
			continue
		default:
			return false
		}
	}
	return true
}

func isJSONFieldName(value string) bool {
	if value == "-" {
		return true
	}
	if value == "" {
		return false
	}
	for _, r := range value {
		if r == '_' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return false
	}
	return true
}

func entityImports(fields []fieldSpec) string {
	for _, field := range fields {
		if field.NeedsTime {
			return "\nimport \"time\"\n"
		}
	}
	return ""
}

func selectColumns(fields []fieldSpec) string {
	columns := []string{"id"}
	for _, field := range fields {
		columns = append(columns, field.Column)
	}
	return strings.Join(columns, ", ")
}

func scanTargets(receiver string, fields []fieldSpec) string {
	targets := []string{"&" + receiver + ".ID"}
	for _, field := range fields {
		targets = append(targets, "&"+receiver+"."+field.GoName)
	}
	return strings.Join(targets, ", ")
}

func insertColumns(fields []fieldSpec) string {
	columns := make([]string, 0, len(fields))
	for _, field := range fields {
		columns = append(columns, field.Column)
	}
	return strings.Join(columns, ", ")
}

func insertValues(fields []fieldSpec) string {
	values := make([]string, 0, len(fields))
	for range fields {
		values = append(values, "?")
	}
	return strings.Join(values, ", ")
}

func insertArgs(receiver string, fields []fieldSpec) string {
	args := make([]string, 0, len(fields))
	for _, field := range fields {
		args = append(args, receiver+"."+field.GoName)
	}
	return strings.Join(args, ", ")
}

func updateSet(fields []fieldSpec) string {
	if len(fields) == 0 {
		return "id = id"
	}

	sets := make([]string, 0, len(fields))
	for _, field := range fields {
		sets = append(sets, field.Column+" = ?")
	}
	return strings.Join(sets, ", ")
}

func updateArgs(receiver string, fields []fieldSpec) string {
	args := make([]string, 0, len(fields)+1)
	for _, field := range fields {
		args = append(args, receiver+"."+field.GoName)
	}
	args = append(args, receiver+".ID")
	return strings.Join(args, ", ")
}

func migrationColumns(fields []fieldSpec) string {
	if len(fields) == 0 {
		return ""
	}

	lines := make([]string, 0, len(fields))
	for _, field := range fields {
		nullability := "NOT NULL"
		defaultValue := field.ZeroSQL
		if field.HasDefault {
			defaultValue = field.DefaultSQL
		}
		defaultSQL := " DEFAULT " + defaultValue
		if field.Nullable {
			nullability = "NULL"
			if field.HasDefault {
				defaultSQL = " DEFAULT " + defaultValue
			} else {
				defaultSQL = ""
			}
		}
		lines = append(lines, fmt.Sprintf(",\n    %s %s %s%s", field.Column, field.SQLType, nullability, defaultSQL))
	}
	for _, field := range fields {
		if field.Unique {
			lines = append(lines, fmt.Sprintf(",\n    UNIQUE KEY uk_%s (%s)", field.Column, field.Column))
		}
	}
	for _, field := range fields {
		if field.Index && !field.Unique {
			lines = append(lines, fmt.Sprintf(",\n    KEY idx_%s (%s)", field.Column, field.Column))
		}
	}
	return strings.Join(lines, "")
}

func integrationImports(fields []fieldSpec) string {
	for _, field := range fields {
		if field.GoType == "time.Time" {
			return "\n\t\"time\""
		}
	}
	return ""
}

func integrationEntityFields(fields []fieldSpec) string {
	if len(fields) == 0 {
		return ""
	}

	lines := make([]string, 0, len(fields))
	for _, field := range fields {
		lines = append(lines, fmt.Sprintf("\n\t\t%s: %s,", field.GoName, integrationLiteral(field)))
	}
	return strings.Join(lines, "")
}

func integrationLiteral(field fieldSpec) string {
	switch field.GoType {
	case "string":
		return fmt.Sprintf("%q", "test-"+field.Name)
	case "int":
		return "123"
	case "int64":
		return "123"
	case "bool":
		return "true"
	case "time.Time":
		return "time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)"
	default:
		return field.ZeroSQL
	}
}

func integrationAssertions(receiver string, fields []fieldSpec) string {
	if len(fields) == 0 {
		return ""
	}

	lines := make([]string, 0, len(fields))
	for _, field := range fields {
		switch field.GoType {
		case "time.Time":
			lines = append(lines, fmt.Sprintf("\n\tif !found.%s.Equal(%s.%s) {\n\t\tt.Fatalf(%q, %s.%s, found.%s)\n\t}", field.GoName, receiver, field.GoName, "expected "+field.GoName+" %v, got %v", receiver, field.GoName, field.GoName))
		default:
			lines = append(lines, fmt.Sprintf("\n\tif found.%s != %s.%s {\n\t\tt.Fatalf(%q, %s.%s, found.%s)\n\t}", field.GoName, receiver, field.GoName, "expected "+field.GoName+" %v, got %v", receiver, field.GoName, field.GoName))
		}
	}
	return strings.Join(lines, "")
}

type usecaseTemplateData struct {
	PackageName string
	TypeName    string
	ActionSnake string
}

type handlerTemplateData struct {
	ModulePath   string
	PackageName  string
	TypeName     string
	VariableName string
	RoutePath    string
}

type commandTemplateData struct {
	ModulePath   string
	PackageName  string
	CommandName  string
	FunctionName string
}

type repositoryTemplateData struct {
	ModulePath              string
	PackageName             string
	TypeName                string
	ReceiverName            string
	FieldName               string
	VariableName            string
	TableName               string
	Fields                  []fieldSpec
	HasFields               bool
	EntityImports           string
	SelectColumns           string
	ScanTargets             string
	InsertColumns           string
	InsertValues            string
	InsertArgs              string
	UpdateSet               string
	UpdateArgs              string
	MigrationColumns        string
	IntegrationImports      string
	IntegrationEntityFields string
	IntegrationAssertions   string
}

type fieldSpec struct {
	Name       string
	GoName     string
	GoType     string
	SQLType    string
	ZeroSQL    string
	DefaultSQL string
	Column     string
	JSONName   string
	Tag        string
	Nullable   bool
	Unique     bool
	Index      bool
	HasDefault bool
	NeedsTime  bool
}

const usecaseGoTemplate = `package {{ .PackageName }}

import "context"

type {{ .TypeName }}Input struct {
}

type {{ .TypeName }}Output struct {
}

type {{ .TypeName }}Usecase struct {
}

func New{{ .TypeName }}Usecase() *{{ .TypeName }}Usecase {
	return &{{ .TypeName }}Usecase{}
}

func (uc *{{ .TypeName }}Usecase) Execute(ctx context.Context, input {{ .TypeName }}Input) (*{{ .TypeName }}Output, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return &{{ .TypeName }}Output{}, nil
}
`

const usecaseTestTemplate = `package {{ .PackageName }}

import (
	"context"
	"testing"
)

func Test{{ .TypeName }}UsecaseExecute(t *testing.T) {
	uc := New{{ .TypeName }}Usecase()

	out, err := uc.Execute(context.Background(), {{ .TypeName }}Input{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatalf("expected output")
	}
}
`

const handlerGoTemplate = `package {{ .PackageName }}

import (
	"encoding/json"
	"net/http"

	"{{ .ModulePath }}/internal/pkg/response"
)

type {{ .TypeName }}Handler struct {
}

type Create{{ .TypeName }}Request struct {
}

func New{{ .TypeName }}Handler() *{{ .TypeName }}Handler {
	return &{{ .TypeName }}Handler{}
}

func (h *{{ .TypeName }}Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET {{ .RoutePath }}", h.List)
	mux.HandleFunc("POST {{ .RoutePath }}", h.Create)
}

func (h *{{ .TypeName }}Handler) List(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, response.Success([]any{}))
}

func (h *{{ .TypeName }}Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input Create{{ .TypeName }}Request
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.JSON(w, http.StatusBadRequest, response.Error("BAD_REQUEST", "invalid request"))
		return
	}

	response.JSON(w, http.StatusCreated, response.Success(map[string]any{}))
}
`

func registerHandlerInRouter(root string, data handlerTemplateData) ([]byte, bool, error) {
	routerPath := filepath.Join(root, "internal", "interfaces", "http", "router.go")
	source, err := os.ReadFile(routerPath)
	if err != nil {
		return nil, false, fmt.Errorf("read router for handler registration: %w", err)
	}

	text := string(source)
	if strings.Contains(text, "New"+data.TypeName+"Handler") {
		return source, false, nil
	}

	importPath := data.ModulePath + "/internal/interfaces/http/handler"
	if !strings.Contains(text, importPath) {
		importMarker := "\n\t\"" + data.ModulePath + "/internal/interfaces/http/middleware\""
		if !strings.Contains(text, importMarker) {
			return nil, false, fmt.Errorf("router import block is not in the expected api-clean format")
		}
		text = strings.Replace(text, importMarker, "\n\t\""+importPath+"\""+importMarker, 1)
	}

	registrationMarker := "\n\thandler := middleware.Chain("
	if !strings.Contains(text, registrationMarker) {
		return nil, false, fmt.Errorf("router NewRouter function is not in the expected api-clean format")
	}

	registration := fmt.Sprintf("\n\t%s := handler.New%sHandler()\n\t%s.RegisterRoutes(mux)\n", data.VariableName, data.TypeName, data.VariableName)
	text = strings.Replace(text, registrationMarker, registration+registrationMarker, 1)

	formatted, err := format.Source([]byte(text))
	if err != nil {
		return nil, false, fmt.Errorf("format registered router: %w", err)
	}
	return formatted, true, nil
}

func registerHandlerInOpenAPI(root string, data handlerTemplateData) ([]byte, bool, error) {
	openAPIPath := filepath.Join(root, "api", "openapi.yaml")
	source, err := os.ReadFile(openAPIPath)
	if err != nil {
		return nil, false, fmt.Errorf("read OpenAPI for handler registration: %w", err)
	}

	text := string(source)
	pathLine := "  " + data.RoutePath + ":"
	if strings.Contains(text, pathLine) {
		return source, false, nil
	}
	if !strings.Contains(text, "\npaths:\n") {
		return nil, false, fmt.Errorf("OpenAPI file is missing paths section")
	}

	marker := "\ncomponents:"
	if !strings.Contains(text, marker) {
		return nil, false, fmt.Errorf("OpenAPI file is not in the expected api-clean format")
	}

	snippet := fmt.Sprintf(`
  %s:
    get:
      tags:
        - %ss
      summary: List %ss
      operationId: list%ss
      responses:
        "200":
          description: List %ss
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ListResponse"
              examples:
                success:
                  value:
                    code: OK
                    message: success
                    data: []
        "400":
          $ref: "#/components/responses/BadRequest"
        "500":
          $ref: "#/components/responses/InternalServerError"
    post:
      tags:
        - %ss
      summary: Create %s
      operationId: create%s
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Create%sRequest"
            examples:
              create:
                value: {}
      responses:
        "201":
          description: Created %s
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/SuccessResponse"
              examples:
                created:
                  value:
                    code: OK
                    message: success
                    data: {}
        "400":
          $ref: "#/components/responses/BadRequest"
        "500":
          $ref: "#/components/responses/InternalServerError"
`, data.RoutePath, data.TypeName, data.TypeName, data.TypeName, data.TypeName, data.TypeName, data.TypeName, data.TypeName, data.TypeName, data.TypeName)
	text = strings.Replace(text, marker, snippet+marker, 1)

	schemaMarker := "\n    ErrorResponse:"
	if !strings.Contains(text, schemaMarker) {
		return nil, false, fmt.Errorf("OpenAPI file is not in the expected api-clean schema format")
	}
	requestSchema := fmt.Sprintf(`
    Create%sRequest:
      type: object
      additionalProperties: true
`, data.TypeName)
	if !strings.Contains(text, "\n    Create"+data.TypeName+"Request:\n") {
		text = strings.Replace(text, schemaMarker, requestSchema+schemaMarker, 1)
	}
	return []byte(text), true, nil
}

func registerSchemaInOpenAPI(root string, typeName string, fields []fieldSpec) ([]byte, bool, error) {
	openAPIPath := filepath.Join(root, "api", "openapi.yaml")
	source, err := os.ReadFile(openAPIPath)
	if err != nil {
		return nil, false, fmt.Errorf("read OpenAPI for schema registration: %w", err)
	}

	text := string(source)
	schemaLine := "    " + typeName + ":"
	if strings.Contains(text, "\n"+schemaLine+"\n") {
		return source, false, nil
	}
	if !strings.Contains(text, "\n  schemas:\n") {
		return nil, false, fmt.Errorf("OpenAPI file is missing schemas section")
	}

	marker := "\n    ErrorResponse:"
	if !strings.Contains(text, marker) {
		return nil, false, fmt.Errorf("OpenAPI file is not in the expected api-clean schema format")
	}

	snippet := openAPISchemaSnippet(typeName, fields)
	text = strings.Replace(text, marker, "\n"+snippet+marker, 1)
	return []byte(text), true, nil
}

func openAPISchemaSnippet(typeName string, fields []fieldSpec) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "    %s:\n", typeName)
	builder.WriteString("      type: object\n")
	builder.WriteString("      required:\n")
	builder.WriteString("        - id\n")
	for _, field := range fields {
		if field.Nullable || field.JSONName == "-" {
			continue
		}
		fmt.Fprintf(&builder, "        - %s\n", field.JSONName)
	}
	builder.WriteString("      properties:\n")
	builder.WriteString("        id:\n")
	builder.WriteString("          type: integer\n")
	builder.WriteString("          format: int64\n")
	builder.WriteString("          example: 1\n")
	for _, field := range fields {
		if field.JSONName == "-" {
			continue
		}
		fmt.Fprintf(&builder, "        %s:\n", field.JSONName)
		for _, line := range openAPIFieldSchemaLines(field) {
			fmt.Fprintf(&builder, "          %s\n", line)
		}
	}
	return builder.String()
}

func openAPIFieldSchemaLines(field fieldSpec) []string {
	var lines []string
	switch field.GoType {
	case "string":
		lines = append(lines, "type: string")
		if maxLength := openAPIStringMaxLength(field.SQLType); maxLength > 0 {
			lines = append(lines, fmt.Sprintf("maxLength: %d", maxLength))
		}
	case "int":
		lines = append(lines, "type: integer", "format: int32")
	case "int64":
		lines = append(lines, "type: integer", "format: int64")
	case "bool":
		lines = append(lines, "type: boolean")
	case "time.Time":
		lines = append(lines, "type: string", "format: date-time")
	default:
		lines = append(lines, "type: string")
	}
	if field.Nullable {
		lines = append(lines, "nullable: true")
	}
	return lines
}

func openAPIStringMaxLength(sqlType string) int {
	upper := strings.ToUpper(strings.TrimSpace(sqlType))
	if !strings.HasPrefix(upper, "VARCHAR(") || !strings.HasSuffix(upper, ")") {
		return 0
	}
	sizeText := strings.TrimSuffix(strings.TrimPrefix(upper, "VARCHAR("), ")")
	size, err := strconv.Atoi(sizeText)
	if err != nil || size <= 0 {
		return 0
	}
	return size
}

func registerRepositoryInAssembly(root string, data repositoryTemplateData) ([]byte, bool, error) {
	assemblyPath := filepath.Join(root, "internal", "app", "assembly.go")
	source, err := os.ReadFile(assemblyPath)
	if err != nil {
		return nil, false, fmt.Errorf("read assembly for repository registration: %w", err)
	}

	text := string(source)
	if strings.Contains(text, data.FieldName+" *mysqlrepo."+data.TypeName+"Repository") {
		return source, false, nil
	}

	importPath := data.ModulePath + "/internal/infrastructure/persistence/mysql"
	if !strings.Contains(text, importPath) {
		importMarker := "\n\t\"" + data.ModulePath + "/internal/config\""
		if !strings.Contains(text, importMarker) {
			return nil, false, fmt.Errorf("assembly import block is not in the expected api-clean format")
		}
		text = strings.Replace(text, importMarker, "\n\tmysqlrepo \""+importPath+"\""+importMarker, 1)
	}

	dependenciesMarker := "\t// gos:dependencies"
	if !strings.Contains(text, dependenciesMarker) {
		return nil, false, fmt.Errorf("assembly dependencies marker not found")
	}
	dependencyField := fmt.Sprintf("\t%s *mysqlrepo.%sRepository\n", data.FieldName, data.TypeName)
	text = strings.Replace(text, dependenciesMarker, dependencyField+dependenciesMarker, 1)

	assembleMarker := "\t// gos:assemble"
	if !strings.Contains(text, assembleMarker) {
		return nil, false, fmt.Errorf("assembly construction marker not found")
	}
	construction := fmt.Sprintf("\t%s := mysqlrepo.New%sRepository(db)\n", data.VariableName, data.TypeName)
	text = strings.Replace(text, assembleMarker, construction+assembleMarker, 1)

	returnMarker := "\t\t// gos:return-dependencies"
	if !strings.Contains(text, returnMarker) {
		return nil, false, fmt.Errorf("assembly return marker not found")
	}
	returnField := fmt.Sprintf("\t\t%s: %s,\n", data.FieldName, data.VariableName)
	text = strings.Replace(text, returnMarker, returnField+returnMarker, 1)

	formatted, err := format.Source([]byte(text))
	if err != nil {
		return nil, false, fmt.Errorf("format registered assembly: %w", err)
	}
	return formatted, true, nil
}

func registerCommandInMain(root string, data commandTemplateData) ([]byte, bool, error) {
	mainPath := filepath.Join(root, "cmd", "api", "main.go")
	source, err := os.ReadFile(mainPath)
	if err != nil {
		return nil, false, fmt.Errorf("read main for command registration: %w", err)
	}

	text := string(source)
	if strings.Contains(text, "appcommand."+data.FunctionName+"()") {
		return source, false, nil
	}

	importPath := data.ModulePath + "/internal/command"
	if !strings.Contains(text, importPath) {
		importMarker := "\n\t// gos:command-imports"
		if !strings.Contains(text, importMarker) {
			return nil, false, fmt.Errorf("command import marker not found")
		}
		text = strings.Replace(text, importMarker, "\n\tappcommand \""+importPath+"\""+importMarker, 1)
	}

	commandMarker := "\n\t// gos:commands"
	if !strings.Contains(text, commandMarker) {
		return nil, false, fmt.Errorf("command registry marker not found")
	}
	registration := fmt.Sprintf("\n\trootCmd.AddCommand(appcommand.%s())", data.FunctionName)
	text = strings.Replace(text, commandMarker, registration+commandMarker, 1)

	formatted, err := format.Source([]byte(text))
	if err != nil {
		return nil, false, fmt.Errorf("format registered command: %w", err)
	}
	return formatted, true, nil
}

const handlerTestTemplate = `package {{ .PackageName }}

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test{{ .TypeName }}HandlerList(t *testing.T) {
	handler := New{{ .TypeName }}Handler()

	req := httptest.NewRequest(http.MethodGet, "{{ .RoutePath }}", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected content type application/json, got %q", got)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["code"] != "OK" {
		t.Fatalf("expected code OK, got %v", body["code"])
	}
	if body["message"] != "success" {
		t.Fatalf("expected message success, got %v", body["message"])
	}
}

func Test{{ .TypeName }}HandlerCreate(t *testing.T) {
	handler := New{{ .TypeName }}Handler()

	req := httptest.NewRequest(http.MethodPost, "{{ .RoutePath }}", strings.NewReader("{}"))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected content type application/json, got %q", got)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["code"] != "OK" {
		t.Fatalf("expected code OK, got %v", body["code"])
	}
}

func Test{{ .TypeName }}HandlerCreateBadRequest(t *testing.T) {
	handler := New{{ .TypeName }}Handler()

	req := httptest.NewRequest(http.MethodPost, "{{ .RoutePath }}", strings.NewReader("{"))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
`

const commandGoTemplate = `package {{ .PackageName }}

import (
	"log/slog"

	"github.com/spf13/cobra"
)

func {{ .FunctionName }}() *cobra.Command {
	return &cobra.Command{
		Use:   "{{ .CommandName }}",
		Short: "Run {{ .CommandName }} command",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if err := ctx.Err(); err != nil {
				return err
			}

			slog.InfoContext(ctx, "command started", "command", "{{ .CommandName }}")
			return nil
		},
	}
}
`

const commandTestTemplate = `package {{ .PackageName }}

import (
	"context"
	"testing"
)

func Test{{ .FunctionName }}(t *testing.T) {
	command := {{ .FunctionName }}()
	command.SetArgs([]string{})

	if err := command.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
`

const domainEntityTemplate = "package {{ .PackageName }}\n{{ .EntityImports }}\n\ntype {{ .TypeName }} struct {\n\tID int64 `json:\"id\"`\n{{- range .Fields }}\n\t{{ .GoName }} {{ .GoType }}{{ .Tag }}\n{{- end }}\n}\n"

const domainRepositoryTemplate = `package {{ .PackageName }}

import "context"

type Repository interface {
	FindByID(ctx context.Context, id int64) (*{{ .TypeName }}, error)
	Save(ctx context.Context, {{ .ReceiverName }} *{{ .TypeName }}) error
	DeleteByID(ctx context.Context, id int64) error
}
`

const mysqlRepositoryTemplate = `package mysql

import (
	"context"
	"database/sql"
	"errors"

	domain "{{ .ModulePath }}/internal/domain/{{ .PackageName }}"
	"{{ .ModulePath }}/internal/infrastructure/database"
)

type {{ .TypeName }}Repository struct {
	db *sql.DB
}

func New{{ .TypeName }}Repository(db *sql.DB) *{{ .TypeName }}Repository {
	return &{{ .TypeName }}Repository{
		db: db,
	}
}

func (r *{{ .TypeName }}Repository) FindByID(ctx context.Context, id int64) (*domain.{{ .TypeName }}, error) {
	executor := database.ExecutorFromContext(ctx, r.db)
	row := executor.QueryRowContext(ctx, "SELECT {{ .SelectColumns }} FROM {{ .TableName }} WHERE id = ?", id)

	var {{ .ReceiverName }} domain.{{ .TypeName }}
	if err := row.Scan({{ .ScanTargets }}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &{{ .ReceiverName }}, nil
}

func (r *{{ .TypeName }}Repository) Save(ctx context.Context, {{ .ReceiverName }} *domain.{{ .TypeName }}) error {
	executor := database.ExecutorFromContext(ctx, r.db)
	if {{ .ReceiverName }}.ID == 0 {
{{- if .HasFields }}
		result, err := executor.ExecContext(ctx, "INSERT INTO {{ .TableName }} ({{ .InsertColumns }}) VALUES ({{ .InsertValues }})", {{ .InsertArgs }})
{{- else }}
		result, err := executor.ExecContext(ctx, "INSERT INTO {{ .TableName }} () VALUES ()")
{{- end }}
		if err != nil {
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		{{ .ReceiverName }}.ID = id
		return nil
	}

	_, err := executor.ExecContext(ctx, "UPDATE {{ .TableName }} SET {{ .UpdateSet }} WHERE id = ?", {{ .UpdateArgs }})
	return err
}

func (r *{{ .TypeName }}Repository) DeleteByID(ctx context.Context, id int64) error {
	executor := database.ExecutorFromContext(ctx, r.db)
	_, err := executor.ExecContext(ctx, "DELETE FROM {{ .TableName }} WHERE id = ?", id)
	return err
}
`

const mysqlRepositoryTestTemplate = `package mysql

import "testing"

func TestNew{{ .TypeName }}Repository(t *testing.T) {
	repo := New{{ .TypeName }}Repository(nil)
	if repo == nil {
		t.Fatalf("expected repository")
	}
}
`

const mysqlRepositoryIntegrationTestTemplate = `//go:build integration

package mysql

import (
	"context"
	"database/sql"
	"os"
	"testing"{{ .IntegrationImports }}

	domain "{{ .ModulePath }}/internal/domain/{{ .PackageName }}"
	_ "github.com/go-sql-driver/mysql"
)

func Test{{ .TypeName }}RepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run repository integration tests")
	}

	ctx := context.Background()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if _, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS {{ .TableName }}"); err != nil {
		t.Fatalf("drop table: %v", err)
	}
	if _, err := db.ExecContext(ctx, ` + "`" + `CREATE TABLE {{ .TableName }} (
    id BIGINT PRIMARY KEY AUTO_INCREMENT{{ .MigrationColumns }}
)` + "`" + `); err != nil {
		t.Fatalf("create table: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DROP TABLE IF EXISTS {{ .TableName }}")
	})

	repo := New{{ .TypeName }}Repository(db)
	{{ .ReceiverName }} := &domain.{{ .TypeName }}{ {{ .IntegrationEntityFields }}
	}

	if err := repo.Save(ctx, {{ .ReceiverName }}); err != nil {
		t.Fatalf("save: %v", err)
	}
	if {{ .ReceiverName }}.ID == 0 {
		t.Fatalf("expected generated id")
	}

	found, err := repo.FindByID(ctx, {{ .ReceiverName }}.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if found == nil {
		t.Fatalf("expected found entity")
	}{{ .IntegrationAssertions }}

	if err := repo.DeleteByID(ctx, {{ .ReceiverName }}.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	deleted, err := repo.FindByID(ctx, {{ .ReceiverName }}.ID)
	if err != nil {
		t.Fatalf("find deleted: %v", err)
	}
	if deleted != nil {
		t.Fatalf("expected deleted entity")
	}
}
`
