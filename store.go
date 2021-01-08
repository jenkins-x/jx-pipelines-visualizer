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
}

type Pipelines struct {
	Pipelines []Pipeline
	Counts    struct {
		Statuses     map[string]int
		Owners       map[string]int
		Repositories map[string]int
		Branches     map[string]int
		Builds       map[string]int
	}
}

func (s *Store) All() (*Pipelines, error) {
	request := bleve.NewSearchRequest(bleve.NewMatchAllQuery())
	request.SortBy([]string{"-Start"})
	request.AddFacet("Status", bleve.NewFacetRequest("Status", 5))
	request.AddFacet("Owner", bleve.NewFacetRequest("Owner", 5))
	request.AddFacet("Repository", bleve.NewFacetRequest("Repository", 20))
	request.AddFacet("Branch", bleve.NewFacetRequest("Branch", 20))
	request.AddFacet("Build", bleve.NewFacetRequest("Build", 20))
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
	request.AddFacet("Status", bleve.NewFacetRequest("Status", 5))
	request.AddFacet("Owner", bleve.NewFacetRequest("Owner", 5))
	request.AddFacet("Repository", bleve.NewFacetRequest("Repository", 20))
	request.AddFacet("Branch", bleve.NewFacetRequest("Branch", 20))
	request.AddFacet("Build", bleve.NewFacetRequest("Build", 20))
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
	if len(q.Owner) > 0 {
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
		switch facet.Field {
		case "Status":
			pipelines.Counts.Statuses = counts
		case "Owner":
			pipelines.Counts.Owners = counts
		case "Repository":
			pipelines.Counts.Repositories = counts
		case "Branch":
			pipelines.Counts.Branches = counts
		case "Build":
			pipelines.Counts.Builds = counts
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
		Start:           startDate,
		End:             endDate,
		Duration:        time.Duration(doc.Fields["Duration"].(float64)),
	}
}
