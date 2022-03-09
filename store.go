package visualizer

import (
	"fmt"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
)

type Store struct {
	Index bleve.Index
}

func NewStore() (*Store, error) {
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name

	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultAnalyzer = keyword.Name

	startFieldMapping := bleve.NewDateTimeFieldMapping()
	startFieldMapping.Name = "Start"
	indexMapping.DefaultMapping.AddFieldMapping(startFieldMapping)
	indexMapping.DefaultMapping.AddFieldMappingsAt("End", bleve.NewDateTimeFieldMapping())

	index, err := bleve.NewMemOnly(indexMapping)
	if err != nil {
		return nil, fmt.Errorf("failed to created a Bleve in-memory Index: %w", err)
	}

	store := &Store{
		Index: index,
	}
	return store, nil
}

func (s *Store) Add(p Pipeline) error {
	return s.Index.Index(p.Name, p)
}

func (s *Store) Delete(name string) error {
	return s.Index.Delete(name)
}

type Query struct {
	Owner      string
	Repository string
	Branch     string
	Query      string
}

type Pipelines struct {
	Pipelines []Pipeline
	Counts    struct {
		Statuses     map[string]int
		Repositories map[string]int
		Authors      map[string]int
		Durations    map[string]int
	}
}

func (s *Store) All() (*Pipelines, error) {
	request := bleve.NewSearchRequest(bleve.NewMatchAllQuery())
	request.SortBy([]string{"-Start"})
	addFacetRequests(request)
	request.Size = 10000
	request.Fields = []string{"*"}
	result, err := s.Index.Search(request)
	if err != nil {
		return nil, fmt.Errorf("failed to search all: %w", err)
	}

	pipelines := bleveResultToPipelines(result)
	return &pipelines, nil
}

func (s *Store) Query(q Query) (*Pipelines, error) {
	request := bleve.NewSearchRequest(q.ToBleveQuery())
	request.SortBy([]string{"-Start"})
	addFacetRequests(request)
	request.Size = 10000
	request.Fields = []string{"*"}
	result, err := s.Index.Search(request)
	if err != nil {
		return nil, fmt.Errorf("failed to search for %v: %w", q, err)
	}

	pipelines := bleveResultToPipelines(result)
	return &pipelines, nil
}

func (q Query) ToBleveQuery() query.Query {
	var queryString strings.Builder
	if len(q.Query) > 0 {
		queryString.WriteString("+")
		queryString.WriteString(q.Query)
	}
	if len(q.Owner) > 0 {
		if queryString.Len() > 0 {
			queryString.WriteString(" ")
		}
		queryString.WriteString("+Owner:")
		queryString.WriteString(q.Owner)
	}
	if len(q.Repository) > 0 {
		if queryString.Len() > 0 {
			queryString.WriteString(" ")
		}
		queryString.WriteString("+Repository:")
		queryString.WriteString(q.Repository)
	}
	if len(q.Branch) > 0 {
		if queryString.Len() > 0 {
			queryString.WriteString(" ")
		}
		queryString.WriteString("+Branch:")
		queryString.WriteString(q.Branch)
	}
	return bleve.NewQueryStringQuery(queryString.String())
}

func bleveResultToPipelines(result *bleve.SearchResult) Pipelines {
	var pipelines Pipelines

	for _, doc := range result.Hits {
		pipeline := bleveDocToPipeline(doc)
		pipelines.Pipelines = append(pipelines.Pipelines, pipeline)
	}

	for _, facet := range result.Facets {
		counts := map[string]int{}
		for _, term := range facet.Terms {
			counts[term.Term] = term.Count
		}
		for _, numericRange := range facet.NumericRanges {
			counts[numericRange.Name] = numericRange.Count
		}
		counts["Other"] = facet.Other
		switch facet.Field {
		case "Status":
			pipelines.Counts.Statuses = counts
		case "Repository":
			pipelines.Counts.Repositories = counts
		case "Author":
			pipelines.Counts.Authors = counts
		case "Duration":
			pipelines.Counts.Durations = counts
		}
	}

	return pipelines
}

func bleveDocToPipeline(doc *search.DocumentMatch) Pipeline {
	var (
		startDate, endDate time.Time
	)
	if start, ok := doc.Fields["Start"].(string); ok {
		startDate, _ = time.Parse(time.RFC3339, start)
	}
	if end, ok := doc.Fields["End"].(string); ok {
		endDate, _ = time.Parse(time.RFC3339, end)
	}
	return Pipeline{
		Name:            doc.Fields["Name"].(string),
		Namespace:       doc.Fields["Namespace"].(string),
		Provider:        doc.Fields["Provider"].(string),
		Owner:           doc.Fields["Owner"].(string),
		Repository:      doc.Fields["Repository"].(string),
		Branch:          doc.Fields["Branch"].(string),
		Build:           doc.Fields["Build"].(string),
		Context:         doc.Fields["Context"].(string),
		Author:          doc.Fields["Author"].(string),
		AuthorAvatarURL: doc.Fields["AuthorAvatarURL"].(string),
		Commit:          doc.Fields["Commit"].(string),
		Status:          doc.Fields["Status"].(string),
		Description:     doc.Fields["Description"].(string),
		GitUrl:          doc.Fields["GitUrl"].(string),
		Start:           startDate,
		End:             endDate,
		Duration:        time.Duration(doc.Fields["Duration"].(float64)),
	}
}

func addFacetRequests(request *bleve.SearchRequest) {
	request.AddFacet("Status", bleve.NewFacetRequest("Status", 4))
	request.AddFacet("Repository", bleve.NewFacetRequest("Repository", 3))
	request.AddFacet("Author", bleve.NewFacetRequest("Author", 3))
	durationFacet := bleve.NewFacetRequest("Duration", 4)
	durationFacet.AddNumericRange("< 5 min", nil, durationAsFloat64Ptr(5*time.Minute))
	durationFacet.AddNumericRange("5-15 min", durationAsFloat64Ptr(5*time.Minute), durationAsFloat64Ptr(15*time.Minute))
	durationFacet.AddNumericRange("15-30 min", durationAsFloat64Ptr(15*time.Minute), durationAsFloat64Ptr(30*time.Minute))
	durationFacet.AddNumericRange("> 30 min", durationAsFloat64Ptr(30*time.Minute), nil)
	request.AddFacet("Duration", durationFacet)
}

func durationAsFloat64Ptr(d time.Duration) *float64 {
	nanos := float64(d.Nanoseconds())
	return &nanos
}
