package snclient

import (
	"context"
	"fmt"
	"pkg/perflib"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func init() {
	AvailableChecks["check_pdh"] = CheckEntry{"check_pdh", NewCheckPDH}
}

type CheckPDH struct {
	query string
}

func NewCheckPDH() CheckHandler {
	return &CheckPDH{}
}

func (l *CheckPDH) Build() *CheckData {
	return &CheckData{
		name:        "check_pdh",
		usage:       "check_pdh query",
		description: "This check queries windows performance counters (pdh).",
		args: map[string]CheckArgument{
			"query": {value: &l.query, description: "The WMI query to execute"},
		},
		implemented: Windows,
		result: &CheckResult{
			State: CheckExitOK,
		},
		emptyState:    ExitCodeUnknown,
		emptySyntax:   "query did not return any result.",
		hasArgsFilter: true, // otherwise empty-syntax won't be applied
		topSyntax:     "${list}",
		detailSyntax:  "%(line)",
		exampleDefault: `
    check_pdh '\Processor(*)\% Processor Time'
    some example output
	`,
		exampleArgs: `'\Processor(*)\% Processor Time'`,
	}
}

type PDHQuery struct {
	tableName     string
	instanceIndex string
	counterName   string
}

func (l *CheckPDH) ParseQuery(query string) (PDHQuery, error) {
	parts := strings.Split(query, "\\")
	result := PDHQuery{}
	tableNameFound := false
	counterNameFound := false
	re := regexp.MustCompile(`\((.*)\)`)
	for _, part := range parts {
		if len(part) != 0 {
			if !tableNameFound {
				indexMatch := re.FindStringSubmatch(part)
				if len(indexMatch) != 0 {
					result.instanceIndex = indexMatch[1]
					result.tableName = strings.Split(part, "(")[0]
				} else {
					result.tableName = part
				}
				tableNameFound = true
			} else if !counterNameFound {
				result.counterName = part
				counterNameFound = true
			}
		}
	}
	if !tableNameFound || !counterNameFound {
		return result, fmt.Errorf("query not in the correct format \\table\\counter")
	}
	return result, nil
}

func (l *CheckPDH) Check(_ context.Context, snc *Agent, check *CheckData, _ []Argument) (*CheckResult, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("check_pdh is a windows only check")
	}
	enabled, _, _ := snc.config.Section("/modules").GetBool("CheckPDH")
	if !enabled {
		return nil, fmt.Errorf("module CheckPDH is not enabled in /modules section")
	}

	if l.query == "" {
		return nil, fmt.Errorf("perfcounter query required")
	}
	qParts, err := l.ParseQuery(l.query)
	if err != nil {
		return nil, err
	}
	tableName := qParts.tableName
	index := perflib.CounterNameTable.LookupIndex(tableName)
	if index == 0 {
		return nil, fmt.Errorf("perfcounter table with name %s not found", tableName)
	}

	perfObject, err := perflib.QueryPerformanceData(strconv.Itoa(int(index)))
	if err != nil {
		return nil, fmt.Errorf("perfcounter query failed: %s", err.Error())
	}
	for _, instance := range perfObject[0].Instances {
		values := []string{}
		entry := map[string]string{}
		for _, counter := range instance.Counters {
			if counter.Def.Name == qParts.counterName {
				l.AddPerfData(check, counter, perfObject[0].Name, instance.Name)

				values = append(values, fmt.Sprintf("%v", counter.Value))
				entryName := fmt.Sprintf("\\%v(%v)\\%v", perfObject[0].Name, instance.Name, counter.Def.Name)
				entry[entryName] = fmt.Sprintf("%v", counter.Value)
			}
		}
		entry["line"] = strings.Join(values, ", ")
		check.listData = append(check.listData, entry)
	}

	return check.Finalize()
}

func (l *CheckPDH) AddPerfData(check *CheckData, counter *perflib.PerfCounter, perfName string, instanceName string) {
	check.result.Metrics = append(check.result.Metrics, &CheckMetric{
		ThresholdName: fmt.Sprintf("\\%s(%s)\\%s", perfName, instanceName, counter.Def.Name),
		Name:          fmt.Sprintf("\\%s(%s)\\%s", perfName, instanceName, counter.Def.Name),
		Value:         counter.Value,
		Warning:       check.warnThreshold,
		Critical:      check.critThreshold,
	})
}
