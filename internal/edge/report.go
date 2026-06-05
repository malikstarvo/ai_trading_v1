package edge

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
)

type ReportData struct {
	GeneratedAt string
	Symbols     string
	Timeframes  string
	Features    []ReportFeature
	TopFeatures []string
	Summary     ReportSummary
}

type ReportFeature struct {
	Name      string
	Metrics   map[string]float64
}

type ReportSummary struct {
	TotalFeatures int
	TotalMetrics  int
	Top3          string
}

func (s *Study) ExportHTML(filePath string) error {
	data := s.buildReportData()

	tmpl := template.Must(template.New("report").Parse(htmlTemplate))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write html: %w", err)
	}
	return nil
}

func (s *Study) ExportCSV(filePath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create csv: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"feature", "label_horizon", "metric", "value", "samples"})

	for _, r := range s.allResults {
		w.Write([]string{
			r.FeatureName,
			r.LabelHorizon,
			r.MetricName,
			fmt.Sprintf("%.6f", r.MetricValue),
			fmt.Sprintf("%d", r.Samples),
		})
	}

	return nil
}

func (s *Study) ExportJSON(filePath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	data := s.buildReportData()
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	return os.WriteFile(filePath, b, 0644)
}

func (s *Study) buildReportData() ReportData {
	featureMap := make(map[string]map[string]float64)
	for _, r := range s.allResults {
		if featureMap[r.FeatureName] == nil {
			featureMap[r.FeatureName] = make(map[string]float64)
		}
		key := r.MetricName
		if r.LabelHorizon != "" {
			key = r.MetricName + "_" + r.LabelHorizon
		}
		featureMap[r.FeatureName][key] = r.MetricValue
	}

	var features []ReportFeature
	for name, metrics := range featureMap {
		features = append(features, ReportFeature{Name: name, Metrics: metrics})
	}

	sort.Slice(features, func(i, j int) bool {
		return features[i].Name < features[j].Name
	})

	top := s.TopFeatures(5)

	symbols := ""
	timeframes := ""
	for i, sym := range s.cfg.Symbols {
		if i > 0 {
			symbols += ", "
		}
		symbols += sym
	}
	for i, tf := range s.cfg.Timeframes {
		if i > 0 {
			timeframes += ", "
		}
		timeframes += tf
	}

	top3 := ""
	for i, name := range top {
		if i >= 3 {
			break
		}
		if i > 0 {
			top3 += ", "
		}
		top3 += fmt.Sprintf("%d. %s", i+1, name)
	}

	return ReportData{
		GeneratedAt: "now",
		Symbols:     symbols,
		Timeframes:  timeframes,
		Features:    features,
		TopFeatures: top,
		Summary: ReportSummary{
			TotalFeatures: len(features),
			TotalMetrics:  len(s.allResults),
			Top3:          top3,
		},
	}
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Edge Validation Report</title>
<style>
body { font-family: -apple-system, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; background: #0d1117; color: #c9d1d9; }
h1 { color: #58a6ff; }
h2 { color: #8b949e; border-bottom: 1px solid #30363d; padding-bottom: 8px; }
table { border-collapse: collapse; width: 100%; margin: 16px 0; }
th, td { text-align: left; padding: 8px 12px; border-bottom: 1px solid #30363d; }
th { background: #161b22; color: #58a6ff; }
tr:hover { background: #1c2128; }
.top { color: #3fb950; }
.mid { color: #d29922; }
.low { color: #f85149; }
.summary { background: #161b22; border-radius: 8px; padding: 16px; margin: 16px 0; }
.feature-name { font-weight: bold; color: #f0f6fc; }
</style>
</head>
<body>
<h1>Edge Validation Report</h1>
<p>Symbols: {{.Symbols}} &middot; Timeframes: {{.Timeframes}}</p>
<p>Generated: {{.GeneratedAt}}</p>

<div class="summary">
<h2>Summary</h2>
<p><strong>Features analyzed:</strong> {{.Summary.TotalFeatures}}</p>
<p><strong>Total metrics:</strong> {{.Summary.TotalMetrics}}</p>
<p><strong>Top 3:</strong> {{.Summary.Top3}}</p>
</div>

<h2>Top Features</h2>
<ol>
{{range .TopFeatures}}
<li><strong>{{.}}</strong></li>
{{end}}
</ol>

<h2>All Features</h2>
{{range .Features}}
<h3 class="feature-name">{{.Name}}</h3>
<table>
<thead><tr><th>Metric</th><th>Value</th></tr></thead>
<tbody>
{{range $key, $val := .Metrics}}
<tr><td>{{$key}}</td><td>{{printf "%.4f" $val}}</td></tr>
{{end}}
</tbody>
</table>
{{end}}
</body>
</html>`
