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
