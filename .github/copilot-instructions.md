# Go Vector Database Project

This is a Go application that connects to PostgreSQL with pgvector extension for vector similarity search, integrated with Ollama for embedding generation.

## Project Structure
- Go application with PostgreSQL and pgvector integration
- Docker Compose setup for local development
- Ollama integration for embedding generation using Qwen3-Embedding-8B model
- Environment-based configuration

## Development Guidelines
- Use Go modules for dependency management
- Follow Go best practices for database connections
- Use environment variables for configuration
- Implement proper error handling and logging
- Use Docker for consistent development environment

## Key Components
- PostgreSQL database with pgvector extension
- Go application with database connectivity
- Ollama service for embedding generation
- Docker Compose for orchestration