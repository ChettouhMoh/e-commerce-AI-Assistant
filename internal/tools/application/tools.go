// Package application implements the concrete tools for the e-commerce AI assistant.
package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	catalogdomain "ecommerce-ai-assistant/internal/catalog/domain"
	toolsdomain "ecommerce-ai-assistant/internal/tools/domain"
)

// --- Tool definitions ---

// SearchProductsDefinition returns the definition for the search_products tool.
func SearchProductsDefinition() toolsdomain.ToolDefinition {
	return toolsdomain.ToolDefinition{
		Name:        "search_products",
		Description: "Search for products in the catalog by query, category, brand, or price range.",
		Parameters: map[string]toolsdomain.ParameterDef{
			"query": {
				Type:        "string",
				Description: "Free-text search query (e.g. 'red hoodie').",
			},
			"category": {
				Type:        "string",
				Description: "Category slug to filter by (e.g. 'hoodies', 'shoes').",
				Enum:        []string{"hoodies", "shoes", "pants", "accessories"},
			},
			"brand": {
				Type:        "string",
				Description: "Brand name to filter by.",
			},
			"min_price": {
				Type:        "number",
				Description: "Minimum price filter.",
			},
			"max_price": {
				Type:        "number",
				Description: "Maximum price filter.",
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum number of results to return (default 10, max 50).",
			},
		},
		Required: []string{},
	}
}

// GetProductDefinition returns the definition for the get_product tool.
func GetProductDefinition() toolsdomain.ToolDefinition {
	return toolsdomain.ToolDefinition{
		Name:        "get_product",
		Description: "Retrieve full details for a single product by its ID.",
		Parameters: map[string]toolsdomain.ParameterDef{
			"product_id": {
				Type:        "string",
				Description: "The unique identifier of the product.",
			},
		},
		Required: []string{"product_id"},
	}
}

// CheckInventoryDefinition returns the definition for the check_inventory tool.
func CheckInventoryDefinition() toolsdomain.ToolDefinition {
	return toolsdomain.ToolDefinition{
		Name:        "check_inventory",
		Description: "Check the stock level and availability for a product variant by SKU.",
		Parameters: map[string]toolsdomain.ParameterDef{
			"sku": {
				Type:        "string",
				Description: "The SKU of the product variant to check.",
			},
		},
		Required: []string{"sku"},
	}
}

// GetOrderStatusDefinition returns the definition for the get_order_status tool.
func GetOrderStatusDefinition() toolsdomain.ToolDefinition {
	return toolsdomain.ToolDefinition{
		Name:        "get_order_status",
		Description: "Retrieve the status and details of a specific order by reference number.",
		Parameters: map[string]toolsdomain.ParameterDef{
			"reference": {
				Type:        "string",
				Description: "The order reference or order number.",
			},
			"customer_id": {
				Type:        "string",
				Description: "The customer ID to verify ownership of the order.",
			},
		},
		Required: []string{"reference"},
	}
}

// GetCustomerOrdersDefinition returns the definition for the get_customer_orders tool.
func GetCustomerOrdersDefinition() toolsdomain.ToolDefinition {
	return toolsdomain.ToolDefinition{
		Name:        "get_customer_orders",
		Description: "List all orders for a given customer.",
		Parameters: map[string]toolsdomain.ParameterDef{
			"customer_id": {
				Type:        "string",
				Description: "The unique identifier of the customer.",
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum number of orders to return (default 20).",
			},
		},
		Required: []string{"customer_id"},
	}
}

// SearchPolicyKnowledgeDefinition returns the definition for the search_policy_knowledge tool.
func SearchPolicyKnowledgeDefinition() toolsdomain.ToolDefinition {
	return toolsdomain.ToolDefinition{
		Name:        "search_policy_knowledge",
		Description: "Search the knowledge base for policy information such as returns, shipping, and warranties.",
		Parameters: map[string]toolsdomain.ParameterDef{
			"query": {
				Type:        "string",
				Description: "The question or topic to search for (e.g. 'return policy').",
			},
			"top_k": {
				Type:        "integer",
				Description: "Number of policy chunks to return (default 3).",
			},
		},
		Required: []string{"query"},
	}
}

// --- Tool handlers ---

// SearchProducts searches the product catalog and returns a JSON array of products.
// The catalogRepo must implement Search(ctx, filter) → ([]Product, error).
func SearchProducts(ctx context.Context, args map[string]any) (string, error) {
	repo, ok := ctx.Value("catalog_repo").(interface {
		Search(ctx context.Context, filter catalogdomain.ProductFilter) ([]catalogdomain.Product, error)
	})
	if !ok {
		return "", fmt.Errorf("catalog repository not available")
	}

	filter := catalogdomain.ProductFilter{
		Query:    GetString(args, "query"),
		Category: GetString(args, "category"),
		Color:    GetString(args, "color"),
		Size:     GetString(args, "size"),
		MaxPrice: GetFloat(args, "max_price"),
		Limit:    GetInt(args, "limit"),
	}

	products, err := repo.Search(ctx, filter)
	if err != nil {
		return "", fmt.Errorf("product search failed: %w", err)
	}

	b, _ := json.Marshal(products)
	return string(b), nil
}

// GetProduct retrieves a single product by ID and returns its JSON representation.
func GetProduct(ctx context.Context, args map[string]any) (string, error) {
	productID, err := MustGetString(args, "product_id")
	if err != nil {
		return "", err
	}

	repo, ok := ctx.Value("catalog_repo").(interface {
		GetByID(ctx context.Context, id string) (*catalogdomain.Product, error)
	})
	if !ok {
		return "", fmt.Errorf("catalog repository not available")
	}

	product, err := repo.GetByID(ctx, productID)
	if err != nil {
		return "", fmt.Errorf("product lookup failed: %w", err)
	}
	if product == nil {
		return "", fmt.Errorf("product not found: %s", productID)
	}

	b, _ := json.Marshal(product)
	return string(b), nil
}

// CheckInventory checks stock for a variant by SKU and returns a JSON stock object.
func CheckInventory(ctx context.Context, args map[string]any) (string, error) {
	sku, err := MustGetString(args, "sku")
	if err != nil {
		return "", err
	}

	repo, ok := ctx.Value("catalog_repo").(interface {
		GetVariant(ctx context.Context, sku string) (*catalogdomain.Variant, error)
	})
	if !ok {
		return "", fmt.Errorf("catalog repository not available")
	}

	variant, err := repo.GetVariant(ctx, sku)
	if err != nil {
		return "", fmt.Errorf("inventory lookup failed: %w", err)
	}
	if variant == nil {
		return "", fmt.Errorf("SKU not found in inventory: %s", sku)
	}

	type stockResponse struct {
		SKU       string `json:"sku"`
		VariantID string `json:"variant_id"`
		Stock     int    `json:"stock"`
		Color     string `json:"color"`
		Size      string `json:"size"`
		InStock   bool   `json:"in_stock"`
	}

	resp := stockResponse{
		SKU:       variant.SKU,
		VariantID: variant.ProductID,
		Stock:     variant.Stock,
		Color:     variant.Color,
		Size:      variant.Size,
		InStock:   variant.Stock > 0,
	}

	b, _ := json.Marshal(resp)
	return string(b), nil
}

// GetOrderStatus retrieves order status by reference and returns a JSON order object.
func GetOrderStatus(ctx context.Context, args map[string]any) (string, error) {
	reference, err := MustGetString(args, "reference")
	if err != nil {
		return "", err
	}
	customerID := GetString(args, "customer_id")

	repo, ok := ctx.Value("orders_repo").(interface {
		GetByReference(ctx context.Context, reference, customerID string) (*catalogOrder, error)
	})
	if !ok {
		return "", fmt.Errorf("orders repository not available")
	}

	order, err := repo.GetByReference(ctx, reference, customerID)
	if err != nil {
		return "", fmt.Errorf("order lookup failed: %w", err)
	}
	if order == nil {
		return "", fmt.Errorf("order not found: %s", reference)
	}

	type orderSummary struct {
		ID          string      `json:"id"`
		OrderNumber string      `json:"order_number"`
		Status      string      `json:"status"`
		Total       float64     `json:"total"`
		Currency    string      `json:"currency"`
		Items       []orderItem `json:"items"`
		Tracking    *string     `json:"tracking_number,omitempty"`
		CreatedAt   string      `json:"created_at"`
	}

	summary := orderSummary{
		ID:          order.ID,
		OrderNumber: order.OrderNumber,
		Status:      string(order.Status),
		Total:       order.Total,
		Currency:    order.Currency,
		CreatedAt:   order.CreatedAt,
	}
	if order.TrackingNumber != "" {
		summary.Tracking = &order.TrackingNumber
	}
	for _, item := range order.Items {
		summary.Items = append(summary.Items, orderItem{
			SKU:         item.SKU,
			ProductName: item.ProductName,
			VariantName: item.VariantName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		})
	}

	b, _ := json.Marshal(summary)
	return string(b), nil
}

type orderItem struct {
	SKU         string  `json:"sku"`
	ProductName string  `json:"product_name"`
	VariantName string  `json:"variant_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// GetCustomerOrders lists orders for a customer and returns a JSON array.
func GetCustomerOrders(ctx context.Context, args map[string]any) (string, error) {
	customerID, err := MustGetString(args, "customer_id")
	if err != nil {
		return "", err
	}

	repo, ok := ctx.Value("orders_repo").(interface {
		ListByCustomer(ctx context.Context, customerID string) ([]catalogOrder, error)
	})
	if !ok {
		return "", fmt.Errorf("orders repository not available")
	}

	orders, err := repo.ListByCustomer(ctx, customerID)
	if err != nil {
		return "", fmt.Errorf("customer orders lookup failed: %w", err)
	}

	type orderBrief struct {
		ID          string  `json:"id"`
		OrderNumber string  `json:"order_number"`
		Status      string  `json:"status"`
		Total       float64 `json:"total"`
		Currency    string  `json:"currency"`
		CreatedAt   string  `json:"created_at"`
	}

	brevies := make([]orderBrief, 0, len(orders))
	for _, o := range orders {
		brevies = append(brevies, orderBrief{
			ID:          o.ID,
			OrderNumber: o.OrderNumber,
			Status:      string(o.Status),
			Total:       o.Total,
			Currency:    o.Currency,
			CreatedAt:   o.CreatedAt,
		})
	}

	b, _ := json.Marshal(brevies)
	return string(b), nil
}

// SearchPolicyKnowledge searches the knowledge base and returns matching policy chunks as JSON.
func SearchPolicyKnowledge(ctx context.Context, args map[string]any) (string, error) {
	query, err := MustGetString(args, "query")
	if err != nil {
		return "", err
	}
	topK := GetInt(args, "top_k")
	if topK <= 0 || topK > 10 {
		topK = 3
	}

	repo, ok := ctx.Value("knowledge_repo").(interface {
		Search(ctx context.Context, query string, topK int) ([]catalogRetrievalResult, error)
	})
	if !ok {
		return "", fmt.Errorf("knowledge repository not available")
	}

	results, err := repo.Search(ctx, query, topK)
	if err != nil {
		return "", fmt.Errorf("knowledge search failed: %w", err)
	}

	type policyChunk struct {
		Source string  `json:"source"`
		Text   string  `json:"text"`
		Score  float64 `json:"score"`
	}

	type policyResponse struct {
		Query   string        `json:"query"`
		Results []policyChunk `json:"results"`
	}

	resp := policyResponse{Query: query}
	for _, r := range results {
		resp.Results = append(resp.Results, policyChunk{
			Source: r.Chunk.Source,
			Text:   r.Chunk.Text,
			Score:  r.Score,
		})
	}

	b, _ := json.Marshal(resp)
	return string(b), nil
}

// --- Domain mirror types (thin wrappers matching the external domain types) ---
// These avoid importing catalog/orders/knowledge directly and instead define
// minimal structs that match the memory adapter return types.

type catalogProduct struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Category    string           `json:"category"`
	Price       float64          `json:"price"`
	Currency    string           `json:"currency"`
	Variants    []catalogVariant `json:"variants"`
	CreatedAt   time.Time        `json:"created_at"`
}

type catalogVariant struct {
	SKU       string `json:"sku"`
	ProductID string `json:"product_id"`
	Color     string `json:"color"`
	Size      string `json:"size"`
	Stock     int    `json:"stock"`
	Barcode   string `json:"barcode"`
}

type catalogOrder struct {
	ID             string      `json:"id"`
	CustomerID     string      `json:"customer_id"`
	OrderNumber    string      `json:"order_number"`
	Status         string      `json:"status"`
	Total          float64     `json:"total"`
	Currency       string      `json:"currency"`
	TrackingNumber string      `json:"tracking_number"`
	Items          []orderItem `json:"items"`
	CreatedAt      string      `json:"created_at"`
}

type catalogMoney struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type catalogAddress struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

type catalogRetrievalResult struct {
	Chunk catalogPolicyChunk `json:"chunk"`
	Score float64            `json:"score"`
}

type catalogPolicyChunk struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Text   string `json:"text"`
}

// --- Tool registration ---

// ToolCallRequest is the public type alias for use in external packages and tests.
type ToolCallRequest = toolsdomain.ToolCallRequest

// ToolResultData is the public type alias for use in external packages and tests.
type ToolResultData = toolsdomain.ToolResultData

// ToolDefinition is the public type alias for use in external packages and tests.
type ToolDefinition = toolsdomain.ToolDefinition

// BuildToolRegistry creates a Registry with all tools registered and an allowlist.
func BuildToolRegistry(catalogRepo interface{}, ordersRepo interface{}, knowledgeRepo interface{}) *Registry {
	tools := []ToolDefinitionWithHandler{
		{Definition: SearchProductsDefinition(), Handler: SearchProducts},
		{Definition: GetProductDefinition(), Handler: GetProduct},
		{Definition: CheckInventoryDefinition(), Handler: CheckInventory},
		{Definition: GetOrderStatusDefinition(), Handler: GetOrderStatus},
		{Definition: GetCustomerOrdersDefinition(), Handler: GetCustomerOrders},
		{Definition: SearchPolicyKnowledgeDefinition(), Handler: SearchPolicyKnowledge},
	}

	allowlist := []string{
		"search_products",
		"get_product",
		"check_inventory",
		"get_order_status",
		"get_customer_orders",
		"search_policy_knowledge",
	}

	return NewRegistry(tools, allowlist).
		WithCatalogRepo(catalogRepo).
		WithOrdersRepo(ordersRepo).
		WithKnowledgeRepo(knowledgeRepo)
}

// NewToolRegistry is a backward-compatible alias for BuildToolRegistry.
func NewToolRegistry(catalogRepo interface{}, ordersRepo interface{}, knowledgeRepo interface{}) *Registry {
	return BuildToolRegistry(catalogRepo, ordersRepo, knowledgeRepo)
}
