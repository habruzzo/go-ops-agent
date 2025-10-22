package agents

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/habruzzo/agent/core"
)

// RAGAgent extends the AI agent with retrieval-augmented generation capabilities
type RAGAgent struct {
	*AIAgent
	knowledgeBase map[string][]Document
	embeddings    map[string][]float64
	mu            sync.RWMutex
}

// Document represents a piece of knowledge in the RAG system
type Document struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// RAGConfig holds configuration for the RAG agent
type RAGConfig struct {
	MaxRetrievedDocs     int     `json:"max_retrieved_docs"`
	SimilarityThreshold  float64 `json:"similarity_threshold"`
	EnableSemanticSearch bool    `json:"enable_semantic_search"`
}

// NewRAGAgent creates a new RAG-enabled AI agent
func NewRAGAgent(name string) *RAGAgent {
	baseAgent := NewAIAgent(name)
	return &RAGAgent{
		AIAgent:       baseAgent,
		knowledgeBase: make(map[string][]Document),
		embeddings:    make(map[string][]float64),
	}
}

// AddDocument adds a document to the knowledge base
func (r *RAGAgent) AddDocument(doc Document) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate embedding for the document
	embedding := r.generateEmbedding(doc.Content)

	// Store document and embedding
	category := r.categorizeDocument(doc)
	r.knowledgeBase[category] = append(r.knowledgeBase[category], doc)
	r.embeddings[doc.ID] = embedding

	slog.Info("Document added to knowledge base",
		"plugin", r.name,
		"type", "agent",
		"doc_id", doc.ID,
		"category", category)
}

// AddMetricsData adds metrics data as documents to the knowledge base
func (r *RAGAgent) AddMetricsData(data []core.DataPoint) {
	for _, point := range data {
		doc := Document{
			ID:      fmt.Sprintf("metric_%s_%d", point.Metric, point.Timestamp.Unix()),
			Content: r.formatMetricAsDocument(point),
			Metadata: map[string]interface{}{
				"type":      "metric",
				"metric":    point.Metric,
				"value":     point.Value,
				"timestamp": point.Timestamp,
				"source":    point.Source,
			},
			Timestamp: point.Timestamp,
		}
		r.AddDocument(doc)
	}
}

// ProcessQueryWithRAG processes a query using RAG
func (r *RAGAgent) ProcessQueryWithRAG(ctx context.Context, query string) (*core.AgentResponse, error) {
	if r.status != core.PluginStatusRunning {
		return nil, core.NewPluginError("rag-agent", "process-query", "agent is not running")
	}

	// Retrieve relevant documents
	relevantDocs := r.retrieveRelevantDocuments(query, 5)

	// Build context from retrieved documents
	contextInfo := r.buildContextFromDocuments(relevantDocs)

	// Create enhanced prompt with retrieved context
	enhancedPrompt := r.buildRAGPrompt(query, contextInfo)

	// Call AI API with enhanced context
	response, err := r.callAIAPI(enhancedPrompt)
	if err != nil {
		return nil, core.WrapError(err, core.ErrorTypeNetwork, "rag-agent", "ai-api-call", "AI API call failed")
	}

	// Convert response to AgentResponse
	agentResponse := r.convertResponseToAgentResponse(response, query)

	// Add RAG metadata
	agentResponse.Metadata["rag_documents_used"] = len(relevantDocs)
	agentResponse.Metadata["rag_sources"] = r.extractSources(relevantDocs)

	return agentResponse, nil
}

// retrieveRelevantDocuments finds documents relevant to the query
func (r *RAGAgent) retrieveRelevantDocuments(query string, maxDocs int) []Document {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Generate embedding for the query
	queryEmbedding := r.generateEmbedding(query)

	var scoredDocs []ScoredDocument

	// Search through all documents
	for category, docs := range r.knowledgeBase {
		for _, doc := range docs {
			if embedding, exists := r.embeddings[doc.ID]; exists {
				similarity := r.cosineSimilarity(queryEmbedding, embedding)
				if similarity > 0.3 { // Threshold for relevance
					scoredDocs = append(scoredDocs, ScoredDocument{
						Document: doc,
						Score:    similarity,
						Category: category,
					})
				}
			}
		}
	}

	// Sort by similarity score
	sort.Slice(scoredDocs, func(i, j int) bool {
		return scoredDocs[i].Score > scoredDocs[j].Score
	})

	// Return top documents
	if len(scoredDocs) > maxDocs {
		scoredDocs = scoredDocs[:maxDocs]
	}

	var result []Document
	for _, scored := range scoredDocs {
		result = append(result, scored.Document)
	}

	return result
}

// ScoredDocument represents a document with its relevance score
type ScoredDocument struct {
	Document Document
	Score    float64
	Category string
}

// generateEmbedding creates a simple embedding for text (in production, use OpenAI embeddings)
func (r *RAGAgent) generateEmbedding(text string) []float64 {
	// Simple word-based embedding (in production, use OpenAI's embedding API)
	words := strings.Fields(strings.ToLower(text))
	embedding := make([]float64, 100) // 100-dimensional embedding

	for i, word := range words {
		if i >= 100 {
			break
		}
		// Simple hash-based embedding
		hash := 0
		for _, char := range word {
			hash = hash*31 + int(char)
		}
		embedding[i] = float64(hash%1000) / 1000.0
	}

	return embedding
}

// cosineSimilarity calculates cosine similarity between two embeddings
func (r *RAGAgent) cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// categorizeDocument determines the category of a document
func (r *RAGAgent) categorizeDocument(doc Document) string {
	content := strings.ToLower(doc.Content)

	if strings.Contains(content, "cpu") || strings.Contains(content, "memory") || strings.Contains(content, "disk") {
		return "system_metrics"
	}
	if strings.Contains(content, "error") || strings.Contains(content, "exception") || strings.Contains(content, "fail") {
		return "errors"
	}
	if strings.Contains(content, "user") || strings.Contains(content, "request") || strings.Contains(content, "api") {
		return "application_metrics"
	}
	if strings.Contains(content, "network") || strings.Contains(content, "latency") || strings.Contains(content, "response") {
		return "network_metrics"
	}

	return "general"
}

// formatMetricAsDocument formats a metric as a document
func (r *RAGAgent) formatMetricAsDocument(point core.DataPoint) string {
	return fmt.Sprintf("Metric: %s, Value: %.2f, Source: %s, Timestamp: %s",
		point.Metric, point.Value, point.Source, point.Timestamp.Format(time.RFC3339))
}

// buildContextFromDocuments builds context string from retrieved documents
func (r *RAGAgent) buildContextFromDocuments(docs []Document) string {
	if len(docs) == 0 {
		return "No relevant context found."
	}

	var contextParts []string
	for i, doc := range docs {
		contextParts = append(contextParts, fmt.Sprintf("Context %d: %s", i+1, doc.Content))
	}

	return strings.Join(contextParts, "\n\n")
}

// buildRAGPrompt creates a prompt with retrieved context
func (r *RAGAgent) buildRAGPrompt(query, context string) map[string]interface{} {
	systemPrompt := `You are an observability expert AI agent with access to real-time system data. 
You have been provided with relevant context from the system's knowledge base.

Relevant Context:
` + context + `

Instructions:
- Use the provided context to answer questions accurately
- If the context doesn't contain enough information, say so
- Provide specific details from the context when available
- Maintain a helpful, technical tone

User Query: ` + query

	return map[string]interface{}{
		"model": r.model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": query,
			},
		},
		"max_tokens":  1000,
		"temperature": 0.7,
	}
}

// extractSources extracts source information from documents
func (r *RAGAgent) extractSources(docs []Document) []string {
	sources := make([]string, 0, len(docs))
	for _, doc := range docs {
		if source, ok := doc.Metadata["source"].(string); ok {
			sources = append(sources, source)
		}
	}
	return sources
}

// GetKnowledgeBaseStats returns statistics about the knowledge base
func (r *RAGAgent) GetKnowledgeBaseStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := map[string]interface{}{
		"total_documents": 0,
		"categories":      make(map[string]int),
	}

	totalDocs := 0
	categories := make(map[string]int)

	for category, docs := range r.knowledgeBase {
		categories[category] = len(docs)
		totalDocs += len(docs)
	}

	stats["total_documents"] = totalDocs
	stats["categories"] = categories

	return stats
}
