package plaidclient

import (
	"context"
	"errors"
	"fmt"

	"github.com/plaid/plaid-go/v21/plaid"
)

type Client struct {
	api *plaid.APIClient
}

type ExchangeResult struct {
	AccessToken string
	ItemID      string
}

type SyncResult struct {
	Added     []plaid.Transaction
	Modified  []plaid.Transaction
	Removed   []plaid.RemovedTransaction
	NextCursor string
	HasMore   bool
}

func New(getenv func(string) string) (Client, error) {
	clientID := getenv("COIN_PLAID_CLIENT_ID")
	secret := getenv("COIN_PLAID_SECRET")
	env := getenv("COIN_PLAID_ENV")

	if clientID == "" || secret == "" || env == "" {
		return Client{}, errors.New("COIN_PLAID_CLIENT_ID, COIN_PLAID_SECRET and COIN_PLAID_ENV must be set")
	}

	plaidEnv, err := resolveEnvironment(env)
	if err != nil {
		return Client{}, err
	}

	cfg := plaid.NewConfiguration()
	cfg.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	cfg.AddDefaultHeader("PLAID-SECRET", secret)
	cfg.UseEnvironment(plaidEnv)

	return Client{api: plaid.NewAPIClient(cfg)}, nil
}

func (c Client) CreateLinkToken(ctx context.Context, userID string) (string, error) {
	user := plaid.NewLinkTokenCreateRequestUser(userID)
	req := plaid.NewLinkTokenCreateRequest(
		"Coin",
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_US},
		*user,
	)
	req.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})

	resp, _, err := c.api.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*req).Execute()
	if err != nil {
		return "", fmt.Errorf("create link token: %w", err)
	}

	return resp.GetLinkToken(), nil
}

func (c Client) ExchangePublicToken(ctx context.Context, publicToken string) (ExchangeResult, error) {
	req := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	resp, _, err := c.api.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(*req).Execute()
	if err != nil {
		return ExchangeResult{}, fmt.Errorf("exchange public token: %w", err)
	}

	return ExchangeResult{
		AccessToken: resp.GetAccessToken(),
		ItemID:      resp.GetItemId(),
	}, nil
}

func (c Client) SyncTransactions(ctx context.Context, accessToken string, cursor *string) (SyncResult, error) {
	req := plaid.NewTransactionsSyncRequest(accessToken)
	if cursor != nil {
		req.SetCursor(*cursor)
	}

	resp, _, err := c.api.PlaidApi.TransactionsSync(ctx).TransactionsSyncRequest(*req).Execute()
	if err != nil {
		return SyncResult{}, fmt.Errorf("sync transactions: %w", err)
	}

	return SyncResult{
		Added:      resp.GetAdded(),
		Modified:   resp.GetModified(),
		Removed:    resp.GetRemoved(),
		NextCursor: resp.GetNextCursor(),
		HasMore:    resp.GetHasMore(),
	}, nil
}

func (c Client) GetAccounts(ctx context.Context, accessToken string) ([]plaid.AccountBase, error) {
	req := plaid.NewAccountsGetRequest(accessToken)
	resp, _, err := c.api.PlaidApi.AccountsGet(ctx).AccountsGetRequest(*req).Execute()
	if err != nil {
		return nil, fmt.Errorf("get accounts: %w", err)
	}

	return resp.GetAccounts(), nil
}

func resolveEnvironment(env string) (plaid.Environment, error) {
	switch env {
	case "sandbox":
		return plaid.Sandbox, nil
	case "production":
		return plaid.Production, nil
	default:
		return "", fmt.Errorf("unknown COIN_PLAID_ENV %q: must be sandbox or production", env)
	}
}
