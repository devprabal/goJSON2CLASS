package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type JavaType struct {
	Title      string
	Properties map[string]interface{}
	Items      interface{}
}

var typedefClassesList []string

func writeJavaCodeToFile(outFile string, javaCode string) {
	var generatedJavaCode string = javaCode

	err := os.WriteFile(outFile, []byte(generatedJavaCode), 0644)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Println("Done!")
}

func generateJavaCode(schema *Schema) string {
	var builder strings.Builder

	processSchemaForJava(&builder, schema, "")
	return builder.String()
}

func getJavaType(property interface{}) string {
	switch p := property.(type) {
	case map[string]interface{}:
		if pType, ok := p["type"].(string); ok {
			switch pType {
			case "string":
				return "String"
			case "number", "decimal":
				return "double"
			case "integer":
				return "int"
			case "boolean":
				return "boolean"
			case "object":
				title, ok := p["title"].(string)
				if ok {
					return getFirstWordFromTitle(title)
				}
			}
		}
	}
	return "unknown"
}

func getItemJavaType(property interface{}) string {
	switch p := property.(type) {
	case map[string]interface{}:
		if pType, ok := p["type"].(string); ok {
			switch pType {
			case "string":
				return "String"
			case "number", "decimal":
				return "double"
			case "integer":
				return "int"
			case "object":
				title, ok := p["title"].(string)
				if ok {
					return getFirstWordFromTitle(title)
				}
			}
		}
	}
	return "unknown"
}

func processSchemaForJava(builder *strings.Builder, schema *Schema, indent string) {
	if schema.Properties != nil {
		var propertyNames []string
		for name := range schema.Properties {
			propertyNames = append(propertyNames, name)
		}
		sort.Strings(propertyNames)

		if schema.Title != "" {
			firstClassName := getFirstWordFromTitle(schema.Title)
			addToTypedefClassesListJava(firstClassName)
		}

		for _, name := range propertyNames {
			property := schema.Properties[name]
			if propertyMap, ok := property.(map[string]interface{}); ok {
				if nestedSchema, ok := propertyMap["properties"].(map[string]interface{}); ok {
					nestedTitle, ok := propertyMap["title"].(string)
					if !ok {
						nestedTitle = name
					}
					nestedPropertyMap := nestedSchema
					nestedSchema := &Schema{
						Title:      nestedTitle,
						Properties: nestedPropertyMap,
					}
					processSchemaForJava(builder, nestedSchema, indent)
				} else if isJavaArrayType(property) {
					propertyMap := property.(map[string]interface{})
					nestedSchema := propertyMap["items"].(map[string]interface{})
					if isJavaObjectType(nestedSchema) {
						nestedTitle := nestedSchema["title"].(string)
						nestedSchema = nestedSchema["properties"].(map[string]interface{})
						nestedPropertyMap := nestedSchema
						itemsSchema := &Schema{
							Title:      nestedTitle,
							Properties: nestedPropertyMap,
						}
						processSchemaForJava(builder, itemsSchema, indent)
					}
				}
			}
		}

		var className string = getFirstWordFromTitle(schema.Title)
		builder.WriteString(indent + "class " + className + " {\n")

		for _, name := range propertyNames {
			property := schema.Properties[name]
			if isJavaArrayType(property) {
				itemType := getJavaArrayType(property)
				builder.WriteString(indent + "    " + "List<" + itemType + "> " + name + ";\n")
			} else {
				propertyType := getJavaType(property)
				builder.WriteString(indent + "    " + propertyType + " " + name + ";\n")
			}
		}

		builder.WriteString(indent + "}\n\n")
	}
}

func isJavaArrayType(property interface{}) bool {
	switch p := property.(type) {
	case map[string]interface{}:
		if _, ok := p["type"]; ok {
			return p["type"] == "array"
		}
	}
	return false
}

func isJavaObjectType(property interface{}) bool {
	switch p := property.(type) {
	case map[string]interface{}:
		if _, ok := p["type"]; ok {
			return p["type"] == "object"
		}
	}
	return false
}

func getJavaArrayType(property interface{}) string {
	switch p := property.(type) {
	case map[string]interface{}:
		if items, ok := p["items"]; ok {
			return getItemJavaType(items)
		}
	}
	return "unknown"
}

func addToTypedefClassesListJava(className string) {
	typedefClassesList = append(typedefClassesList, className)
}
