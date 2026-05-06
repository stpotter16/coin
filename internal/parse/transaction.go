package parse

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stpotter16/coin/internal/types"
)

func ParseTransactionWrite(r *http.Request) (types.TransactionWrite, error) {
	var body struct {
		AccountID    *int    `json:"account_id"`
		Amount       float64 `json:"amount"` // signed: positive = expense, negative = income
		Date         string  `json:"date"`
		Description  string  `json:"description"`
		MerchantName *string `json:"merchant_name"`
		Pending      bool    `json:"pending"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return types.TransactionWrite{}, err
	}
	if body.Description == "" {
		return types.TransactionWrite{}, fmt.Errorf("description is required")
	}
	if _, err := time.Parse("2006-01-02", body.Date); err != nil {
		return types.TransactionWrite{}, fmt.Errorf("invalid date %q", body.Date)
	}
	return types.TransactionWrite{
		AccountID:    body.AccountID,
		Amount:       body.Amount,
		Date:         body.Date,
		Description:  body.Description,
		MerchantName: body.MerchantName,
		Pending:      body.Pending,
	}, nil
}

// ParseTransactionImportCSV parses a CSV reader into import rows, resolving
// account names case-insensitively against the provided account list. Each row
// carries its own Errors slice; the caller decides whether to reject the batch.
// Returns a top-level error only for structural problems (bad CSV, missing
// required columns, no data rows).
func ParseTransactionImportCSV(r io.Reader, accounts []types.Account) ([]types.TransactionImportRow, error) {
	accountMap := make(map[string]int, len(accounts))
	for _, a := range accounts {
		accountMap[strings.ToLower(a.Name)] = a.ID
	}

	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true

	headers, err := cr.Read()
	if err == io.EOF {
		return nil, fmt.Errorf("empty file")
	}
	if err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}

	colIndex := make(map[string]int, len(headers))
	for i, h := range headers {
		colIndex[strings.ToLower(strings.TrimSpace(h))] = i
	}
	for _, req := range []string{"date", "description", "amount"} {
		if _, ok := colIndex[req]; !ok {
			return nil, fmt.Errorf("missing required column %q", req)
		}
	}

	get := func(record []string, col string) string {
		i, ok := colIndex[col]
		if !ok || i >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[i])
	}

	var rows []types.TransactionImportRow
	for rowNum := 1; ; rowNum++ {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading row %d: %w", rowNum, err)
		}

		var rowErrors []string

		description := get(record, "description")
		if description == "" {
			rowErrors = append(rowErrors, "description is required")
		}

		dateStr := get(record, "date")
		if dateStr == "" {
			rowErrors = append(rowErrors, "date is required")
		} else if _, err := time.Parse("2006-01-02", dateStr); err != nil {
			rowErrors = append(rowErrors, fmt.Sprintf("invalid date %q", dateStr))
		}

		var amount float64
		amountStr := get(record, "amount")
		if amountStr == "" {
			rowErrors = append(rowErrors, "amount is required")
		} else if amount, err = strconv.ParseFloat(amountStr, 64); err != nil {
			rowErrors = append(rowErrors, fmt.Sprintf("invalid amount %q", amountStr))
		}

		accountNameRaw := get(record, "account_name")
		var accountID *int
		if accountNameRaw != "" {
			if id, ok := accountMap[strings.ToLower(accountNameRaw)]; ok {
				accountID = &id
			}
		}

		merchantNameRaw := get(record, "merchant_name")
		var merchantName *string
		if merchantNameRaw != "" {
			merchantName = &merchantNameRaw
		}

		rows = append(rows, types.TransactionImportRow{
			RowNum:       rowNum,
			Date:         dateStr,
			Description:  description,
			Amount:       amount,
			AccountID:    accountID,
			AccountName:  accountNameRaw,
			MerchantName: merchantName,
			Errors:       rowErrors,
		})
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no data rows")
	}
	return rows, nil
}

func ParseTransactionImportWrite(r *http.Request) ([]types.TransactionWrite, error) {
	var body []struct {
		AccountID    *int    `json:"account_id"`
		Amount       float64 `json:"amount"`
		Date         string  `json:"date"`
		Description  string  `json:"description"`
		MerchantName *string `json:"merchant_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("no rows provided")
	}
	result := make([]types.TransactionWrite, 0, len(body))
	for i, row := range body {
		if row.Description == "" {
			return nil, fmt.Errorf("row %d: description is required", i+1)
		}
		if row.Date == "" {
			return nil, fmt.Errorf("row %d: date is required", i+1)
		}
		if _, err := time.Parse("2006-01-02", row.Date); err != nil {
			return nil, fmt.Errorf("row %d: invalid date %q", i+1, row.Date)
		}
		result = append(result, types.TransactionWrite{
			AccountID:    row.AccountID,
			Amount:       row.Amount,
			Date:         row.Date,
			Description:  row.Description,
			MerchantName: row.MerchantName,
		})
	}
	return result, nil
}

func ParseTransactionDTO(dto types.TransactionDTO) (types.Transaction, error) {
	date, err := time.Parse("2006-01-02", dto.TransactionDate)
	if err != nil {
		return types.Transaction{}, fmt.Errorf("invalid transaction date %q: %w", dto.TransactionDate, err)
	}

	t := types.Transaction{
		ID:               dto.ID,
		AccountName:      dto.AccountName,
		Amount:           dto.Amount,
		TransactionDate:  date,
		Description:      dto.Description,
		Pending:          dto.Pending,
		CreatedTime:      dto.CreatedTime,
		LastModifiedTime: dto.LastModifiedTime,
	}

	if dto.AccountID.Valid {
		id := int(dto.AccountID.Int64)
		t.AccountID = &id
	}

	if dto.MerchantName.Valid {
		t.MerchantName = types.MerchantName{Value: &dto.MerchantName.String}
	}

	if dto.PlanItemID.Valid {
		t.PlanItem = &types.AssignedPlanItem{
			ID:   int(dto.PlanItemID.Int64),
			Name: dto.PlanItemName.String,
		}
	}

	return t, nil
}
