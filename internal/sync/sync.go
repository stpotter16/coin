package sync

import (
	"context"
	"log"

	"github.com/stpotter16/coin/internal/crypto"
	"github.com/stpotter16/coin/internal/parse"
	"github.com/stpotter16/coin/internal/plaidclient"
	"github.com/stpotter16/coin/internal/store"
	"github.com/stpotter16/coin/internal/types"
)

type Syncer struct {
	store         store.Store
	plaidClient   plaidclient.Client
	encryptionKey []byte
}

func New(store store.Store, client plaidclient.Client, encryptionKey []byte) Syncer {
	return Syncer{
		store:         store,
		plaidClient:   client,
		encryptionKey: encryptionKey,
	}
}

// SyncAll syncs transactions and account balances for every linked plaid item.
func (s Syncer) SyncAll(ctx context.Context) {
	items, err := s.store.GetPlaidItems(ctx)
	if err != nil {
		log.Printf("sync: failed to load plaid items: %v", err)
		return
	}

	for _, item := range items {
		if err := s.syncItem(ctx, item); err != nil {
			log.Printf("sync: error syncing item %s (%s): %v", item.InstitutionName, item.PlaidItemID, err)
		}
	}
}

func (s Syncer) syncItem(ctx context.Context, item types.PlaidItem) error {
	accessToken, err := crypto.Decrypt(s.encryptionKey, item.PlaidAccessToken)
	if err != nil {
		return err
	}
	token := string(accessToken)

	if err := s.syncAccounts(ctx, item, token); err != nil {
		return err
	}

	if err := s.syncTransactions(ctx, item, token); err != nil {
		return err
	}

	return nil
}

func (s Syncer) syncAccounts(ctx context.Context, item types.PlaidItem, accessToken string) error {
	plaidAccounts, err := s.plaidClient.GetAccounts(ctx, accessToken)
	if err != nil {
		return err
	}

	for _, pa := range plaidAccounts {
		account, err := parse.ParsePlaidAccount(pa, item.ID)
		if err != nil {
			return err
		}
		if err := s.store.UpsertAccount(ctx, account); err != nil {
			return err
		}
	}

	return nil
}

func (s Syncer) syncTransactions(ctx context.Context, item types.PlaidItem, accessToken string) error {
	// Build plaid_account_id → internal id lookup map.
	accounts, err := s.store.GetAccountsByItemID(ctx, item.ID)
	if err != nil {
		return err
	}
	accountMap := make(map[string]int, len(accounts))
	for _, a := range accounts {
		accountMap[a.PlaidAccountID] = a.ID
	}

	cursor := item.TransactionCursor
	for {
		result, err := s.plaidClient.SyncTransactions(ctx, accessToken, cursor)
		if err != nil {
			return err
		}

		for _, pt := range result.Added {
			accountID, ok := accountMap[pt.GetAccountId()]
			if !ok {
				log.Printf("sync: unknown account id %s, skipping transaction %s", pt.GetAccountId(), pt.GetTransactionId())
				continue
			}
			tx, err := parse.ParsePlaidTransaction(pt, accountID)
			if err != nil {
				return err
			}
			if err := s.store.UpsertPlaidTransaction(ctx, tx); err != nil {
				return err
			}
		}

		for _, pt := range result.Modified {
			accountID, ok := accountMap[pt.GetAccountId()]
			if !ok {
				log.Printf("sync: unknown account id %s, skipping transaction %s", pt.GetAccountId(), pt.GetTransactionId())
				continue
			}
			tx, err := parse.ParsePlaidTransaction(pt, accountID)
			if err != nil {
				return err
			}
			if err := s.store.UpsertPlaidTransaction(ctx, tx); err != nil {
				return err
			}
		}

		for _, rt := range result.Removed {
			if err := s.store.DeletePlaidTransaction(ctx, rt.GetTransactionId()); err != nil {
				return err
			}
		}

		nextCursor := result.NextCursor
		if err := s.store.UpdatePlaidItemCursor(ctx, item.ID, nextCursor); err != nil {
			return err
		}
		cursor = &nextCursor

		if !result.HasMore {
			break
		}
	}

	log.Printf("sync: completed item %s (%s)", item.InstitutionName, item.PlaidItemID)
	return nil
}
