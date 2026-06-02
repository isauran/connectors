package bpmn

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// DMNDefinitions represents the root definitions element of a DMN XML.
type DMNDefinitions struct {
	XMLName   xml.Name   `xml:"definitions"`
	Decisions []Decision `xml:"decision"`
}

// Decision represents a single decision component.
type Decision struct {
	ID            string        `xml:"id,attr"`
	Name          string        `xml:"name,attr"`
	DecisionTable DecisionTable `xml:"decisionTable"`
}

// DecisionTable represents a DMN decision table.
type DecisionTable struct {
	ID        string   `xml:"id,attr"`
	HitPolicy string   `xml:"hitPolicy,attr"` // UNIQUE, FIRST, ANY, COLLECT
	Inputs    []Input  `xml:"input"`
	Outputs   []Output `xml:"output"`
	Rules     []Rule   `xml:"rule"`
}

// Input represents an input column in the decision table.
type Input struct {
	ID              string          `xml:"id,attr"`
	Label           string          `xml:"label,attr"`
	InputExpression InputExpression `xml:"inputExpression"`
}

// InputExpression defines the variable name evaluated.
type InputExpression struct {
	TypeRef string `xml:"typeRef,attr"`
	Text    string `xml:"text"`
}

// Output represents an output column in the decision table.
type Output struct {
	ID      string `xml:"id,attr"`
	Label   string `xml:"label,attr"`
	Name    string `xml:"name,attr"`
	TypeRef string `xml:"typeRef,attr"`
}

// Rule represents a single rule row in the decision table.
type Rule struct {
	ID            string        `xml:"id,attr"`
	InputEntries  []InputEntry  `xml:"inputEntry"`
	OutputEntries []OutputEntry `xml:"outputEntry"`
}

// InputEntry represents a condition inside a rule row.
type InputEntry struct {
	ID   string `xml:"id,attr"`
	Text string `xml:"text"`
}

// OutputEntry represents a resulting output value inside a rule row.
type OutputEntry struct {
	ID   string `xml:"id,attr"`
	Text string `xml:"text"`
}

// ParseDMN parses raw DMN XML data into Go structures.
func ParseDMN(xmlData []byte) (*DMNDefinitions, error) {
	var defs DMNDefinitions
	err := xml.Unmarshal(xmlData, &defs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal DMN XML: %w", err)
	}
	return &defs, nil
}

// Evaluate evaluates a decision table using input variables.
func Evaluate(dmn *DMNDefinitions, decisionID string, variables map[string]interface{}) (map[string]interface{}, error) {
	var dec *Decision
	for _, d := range dmn.Decisions {
		if d.ID == decisionID {
			dec = &d
			break
		}
	}

	if dec == nil {
		return nil, fmt.Errorf("decision with ID %s not found", decisionID)
	}

	table := dec.DecisionTable
	var matchedRules []Rule

	for _, rule := range table.Rules {
		if len(rule.InputEntries) != len(table.Inputs) {
			return nil, fmt.Errorf("rule %s has mismatched inputs count: expected %d, got %d", rule.ID, len(table.Inputs), len(rule.InputEntries))
		}

		match := true
		for i, entry := range rule.InputEntries {
			inputVarName := strings.TrimSpace(table.Inputs[i].InputExpression.Text)
			val := variables[inputVarName]

			if !matchInputEntry(val, entry.Text) {
				match = false
				break
			}
		}

		if match {
			matchedRules = append(matchedRules, rule)
		}
	}

	if len(matchedRules) == 0 {
		return nil, nil // No rules matched
	}

	// Apply Hit Policy
	policy := strings.ToUpper(table.HitPolicy)
	if policy == "" {
		policy = "UNIQUE"
	}

	switch policy {
	case "UNIQUE":
		if len(matchedRules) > 1 {
			return nil, fmt.Errorf("hit policy UNIQUE violated: multiple rules matched (%d)", len(matchedRules))
		}
		return buildOutputMap(table.Outputs, matchedRules[0].OutputEntries)
	case "FIRST":
		return buildOutputMap(table.Outputs, matchedRules[0].OutputEntries)
	case "ANY":
		// Ensure output matches. For simplicity we just return the first one
		return buildOutputMap(table.Outputs, matchedRules[0].OutputEntries)
	}

	return nil, fmt.Errorf("hit policy %s is not supported", policy)
}

func buildOutputMap(outputs []Output, entries []OutputEntry) (map[string]interface{}, error) {
	if len(outputs) != len(entries) {
		return nil, fmt.Errorf("outputs count mismatch: expected %d, got %d", len(outputs), len(entries))
	}

	result := make(map[string]interface{})
	for i, out := range outputs {
		valStr := strings.TrimSpace(entries[i].Text)
		// Strip quotes for string types
		valStr = strings.Trim(valStr, `"'`)

		// Attempt conversions
		if b, err := strconv.ParseBool(valStr); err == nil {
			result[out.Name] = b
		} else if f, err := strconv.ParseFloat(valStr, 64); err == nil {
			result[out.Name] = f
		} else {
			result[out.Name] = valStr
		}
	}

	return result, nil
}

func matchInputEntry(val interface{}, entryText string) bool {
	entryText = strings.TrimSpace(entryText)
	if entryText == "" || entryText == "-" {
		return true
	}

	strVal := strings.TrimSpace(fmt.Sprintf("%v", val))

	if strings.HasPrefix(entryText, "<") {
		limitStr := strings.TrimSpace(strings.TrimPrefix(entryText, "<"))
		v, err1 := strconv.ParseFloat(strVal, 64)
		l, err2 := strconv.ParseFloat(limitStr, 64)
		return err1 == nil && err2 == nil && v < l
	}
	if strings.HasPrefix(entryText, ">") {
		limitStr := strings.TrimSpace(strings.TrimPrefix(entryText, ">"))
		v, err1 := strconv.ParseFloat(strVal, 64)
		l, err2 := strconv.ParseFloat(limitStr, 64)
		return err1 == nil && err2 == nil && v > l
	}

	entryText = strings.Trim(entryText, `"'`)
	return strVal == entryText
}
