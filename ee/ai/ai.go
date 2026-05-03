// Package ai provides the enterprise AI suite: Chat (RAG), writing assistant,
// and AI-powered search reranking.
//
// Gated behind [license.FeatureAIChat] and [license.FeatureAIAssistant].
//
// The core open-source search (internal/search, internal/vectorstore) provides
// full-text and vector search. This package adds:
//   - AI Chat: conversational RAG over wiki content with source citations
//   - AI writing assistant: summarize, translate, expand, tone-shift
//   - AI answer generation in search results
//   - Streaming responses via SSE
package ai
