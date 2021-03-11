// Copyright 2021 The Rode Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sonar

// Event is...
type Event struct {
	TaskID      string            `json:"taskid"`
	Status      string            `json:"status"`
	AnalyzedAt  string            `json:"analyzedat"`
	GitCommit   string            `json:"revision"`
	Project     *Project          `json:"project"`
	QualityGate *QualityGate      `json:"qualityGate"`
	Branch      *Branch           `json:"branch"`
	Properties  map[string]string `json:"properties"`
}

// Branch is...
type Branch struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	IsMain bool   `json:"isMain"`
	URL    string `json:"url"`
}

// Project is
type Project struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// QualityGate is...
type QualityGate struct {
	Conditions []*Condition `json:"conditions"`
	Name       string       `json:"name"`
	Status     string       `json:"status"`
}

// Condition is...
type Condition struct {
	ErrorThreshold string `json:"errorThreshold"`
	Metric         string `json:"metric"`
	OnLeakPeriod   bool   `json:"onLeakPeriod"`
	Operator       string `json:"operator"`
	Status         string `json:"status"`
}

// Overall Measures
type OverallMeasures struct {
	Component struct {
		ID        string `json:"id"`
		Key       string `json:"key"`
		Name      string `json:"name"`
		Qualifier string `json:"qualifier"`
		Measures  []struct {
			Metric    string `json:"metric"`
			Value     string `json:"value,omitempty"`
			BestValue bool   `json:"bestValue,omitempty"`
			Periods   []struct {
				Index     int    `json:"index"`
				Value     string `json:"value"`
				BestValue bool   `json:"bestValue"`
			} `json:"periods,omitempty"`
			Period struct {
				Index     int    `json:"index"`
				Value     string `json:"value"`
				BestValue bool   `json:"bestValue"`
			} `json:"period,omitempty"`
		} `json:"measures"`
	} `json:"component"`
	Metrics []struct {
		Key                   string `json:"key"`
		Name                  string `json:"name"`
		Description           string `json:"description"`
		Domain                string `json:"domain"`
		Type                  string `json:"type"`
		HigherValuesAreBetter bool   `json:"higherValuesAreBetter,omitempty"`
		Qualitative           bool   `json:"qualitative"`
		Hidden                bool   `json:"hidden"`
		Custom                bool   `json:"custom"`
		BestValue             string `json:"bestValue,omitempty"`
		DecimalScale          int    `json:"decimalScale,omitempty"`
		WorstValue            string `json:"worstValue,omitempty"`
	} `json:"metrics"`
	Periods []struct {
		Index int    `json:"index"`
		Mode  string `json:"mode"`
		Date  string `json:"date"`
	} `json:"periods"`
}

// Individual Issues
type Issues struct {
	Total  int `json:"total"`
	P      int `json:"p"`
	Ps     int `json:"ps"`
	Paging struct {
		PageIndex int `json:"pageIndex"`
		PageSize  int `json:"pageSize"`
		Total     int `json:"total"`
	} `json:"paging"`
	EffortTotal int `json:"effortTotal"`
	DebtTotal   int `json:"debtTotal"`
	Issues      []struct {
		Key       string `json:"key"`
		Rule      string `json:"rule"`
		Severity  string `json:"severity"`
		Component string `json:"component"`
		Project   string `json:"project"`
		Line      int    `json:"line"`
		Hash      string `json:"hash"`
		TextRange struct {
			StartLine   int `json:"startLine"`
			EndLine     int `json:"endLine"`
			StartOffset int `json:"startOffset"`
			EndOffset   int `json:"endOffset"`
		} `json:"textRange"`
		Flows        []interface{} `json:"flows"`
		Status       string        `json:"status"`
		Message      string        `json:"message"`
		Effort       string        `json:"effort"`
		Debt         string        `json:"debt"`
		Tags         []string      `json:"tags"`
		Transitions  []interface{} `json:"transitions"`
		Actions      []interface{} `json:"actions"`
		Comments     []interface{} `json:"comments"`
		CreationDate string        `json:"creationDate"`
		UpdateDate   string        `json:"updateDate"`
		Type         string        `json:"type"`
		Organization string        `json:"organization"`
		Scope        string        `json:"scope"`
	} `json:"issues"`
	Components []struct {
		Organization string `json:"organization"`
		Key          string `json:"key"`
		UUID         string `json:"uuid"`
		Enabled      bool   `json:"enabled"`
		Qualifier    string `json:"qualifier"`
		Name         string `json:"name"`
		LongName     string `json:"longName"`
		Path         string `json:"path,omitempty"`
	} `json:"components"`
	Rules []struct {
		Key      string `json:"key"`
		Name     string `json:"name"`
		Lang     string `json:"lang"`
		Status   string `json:"status"`
		LangName string `json:"langName"`
	} `json:"rules"`
	Users     []interface{} `json:"users"`
	Languages []struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"languages"`
	Facets []struct {
		Property string `json:"property"`
		Values   []struct {
			Val   string `json:"val"`
			Count int    `json:"count"`
		} `json:"values"`
	} `json:"facets"`
}
