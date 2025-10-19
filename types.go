package main

type SearchRequest struct {
	Query  string `json:"query"`
	ShopID string `json:"shopid"`
	Limit  int    `json:"limit,omitempty"`
}

type SearchResult struct {
	Content    string  `json:"content"`
	File       string  `json:"file"`
	Filename   string  `json:"filename"`
	ShopID     string  `json:"shopid"`
	Chunk      int     `json:"chunk"`
	Similarity float64 `json:"similarity"`
}

type SearchResponse struct {
	Query   string         `json:"query"`
	ShopID  string         `json:"shopid"`
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
	Summary string         `json:"summary,omitempty"`
	Error   string         `json:"error,omitempty"`
}

type StatusResponse struct {
	Status       string `json:"status"`
	TotalRecords int    `json:"total_records"`
	Message      string `json:"message"`
}

type StatsResponse struct {
	TotalRecords int             `json:"total_records"`
	ByShop       []ShopStats     `json:"by_shop"`
	ByFile       []FileStats     `json:"by_file"`
	ByShopFile   []ShopFileStats `json:"by_shop_file"`
}

type ShopStats struct {
	ShopID string `json:"shopid"`
	Count  int    `json:"count"`
}

type FileStats struct {
	Filename string `json:"filename"`
	Count    int    `json:"count"`
}

type ShopFileStats struct {
	ShopID   string `json:"shopid"`
	Filename string `json:"filename"`
	Count    int    `json:"count"`
}

type BuildDocRequest struct {
	ShopID   string `json:"shopid"`
	Filename string `json:"filename"`
}

type BuildDocResponse struct {
	ShopID     string `json:"shopid"`
	Filename   string `json:"filename"`
	Chunks     int    `json:"chunks"`
	Embeddings int    `json:"embeddings"`
	Message    string `json:"message"`
	Error      string `json:"error,omitempty"`
}

type OllamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type OllamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type OllamaGenerateRequest struct {
	Model       string                 `json:"model"`
	Prompt      string                 `json:"prompt"`
	Stream      bool                   `json:"stream"`
	Options     map[string]interface{} `json:"options,omitempty"`
	Temperature float64                `json:"-"`
}

type OllamaGenerateResponse struct {
	Response string `json:"response"`
}
