package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"searchservice/internal/models"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

const (
	indexName = "products"
)

type ElasticRepository struct {
	client *elasticsearch.Client
}

// NewElasticRepository создаёт новый экземпляр репозитория Elasticsearch
func NewElasticRepository(esURL string) (*ElasticRepository, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{esURL},
	}
	cli, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating Elasticsearch client: %w", err)
	}

	er := &ElasticRepository{client: cli}
	if err := er.ensureIndex(); err != nil {
		return nil, fmt.Errorf("failed to ensure index: %w", err)
	}
	return er, nil
}

func (er *ElasticRepository) ensureIndex() error {
	existsRes, err := er.client.Indices.Exists([]string{indexName})
	if err != nil {
		return fmt.Errorf("error checking index existence: %w", err)
	}
	defer existsRes.Body.Close()

	if existsRes.StatusCode == 200 {
		log.Println("Index 'products' already exists")
		return nil
	}

	mapping := `
    {
      "mappings": {
        "properties": {
          "name":        { "type": "text" },
          "description": { "type": "text" },
          "tags":        { "type": "text" },
          "seller":      { "type": "keyword" }
        }
      }
    }`

	createRes, err := er.client.Indices.Create(
		indexName,
		er.client.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("elasticsearch error on index create: %s", createRes.String())
	}

	log.Println("Index 'products' created")
	return nil
}

func (er *ElasticRepository) IndexProduct(product *models.Product) error {
	data, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("error marshaling product %#v: %w", product, err)
	}
	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: product.ProductID,
		Body:       strings.NewReader(string(data)),
		Refresh:    "true",
	}
	res, err := req.Do(context.Background(), er.client)
	if err != nil {
		return fmt.Errorf("error indexing document ID=%s: %w", product.ProductID, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("elasticsearch error for ID=%s: %s", product.ProductID, res.String())
	}

	log.Printf("Indexed product ID=%s\n", product.ProductID)
	return nil
}

func (er *ElasticRepository) SearchProductIDs(q string) ([]string, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []interface{}{
					map[string]interface{}{
						"match_phrase_prefix": map[string]interface{}{"name": q},
					},
					map[string]interface{}{
						"match_phrase_prefix": map[string]interface{}{"description": q},
					},
					map[string]interface{}{
						"match_phrase_prefix": map[string]interface{}{"tags": q},
					},
					map[string]interface{}{
						"wildcard": map[string]interface{}{
							"seller": fmt.Sprintf("*%s*", q),
						},
					},
				},
			},
		},
		"_source": false, // возвращаем только _id
	}

	var buf strings.Builder
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("error encoding search query: %w", err)
	}

	res, err := er.client.Search(
		er.client.Search.WithContext(context.Background()),
		er.client.Search.WithIndex(indexName),
		er.client.Search.WithBody(strings.NewReader(buf.String())),
		er.client.Search.WithTrackTotalHits(true),
		er.client.Search.WithSize(100),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting response from Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch returned error: %s", res.String())
	}

	// Парсим ответ
	type hit struct {
		ID string `json:"_id"`
	}
	type hitsWrapper struct {
		Hits []hit `json:"hits"`
	}
	type respWrapper struct {
		Hits hitsWrapper `json:"hits"`
	}
	var sr respWrapper
	if err := json.NewDecoder(res.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("error parsing search response: %w", err)
	}

	resultIDs := make([]string, 0, len(sr.Hits.Hits))
	for _, h := range sr.Hits.Hits {
		resultIDs = append(resultIDs, h.ID)
	}
	return resultIDs, nil
}
