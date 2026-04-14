package parse_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stpotter16/coin/internal/parse"
	"github.com/stpotter16/coin/internal/types"
)

func makeRequest(t *testing.T, body string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

var testAccounts = []types.Account{
	{ID: 1, Name: "Chase Checking"},
	{ID: 2, Name: "Amex Card"},
}

// ---------------------------------------------------------------------------
// ParseTransactionImportCSV
// ---------------------------------------------------------------------------

func TestParseTransactionImportCSV_Valid(t *testing.T) {
	csv := "date,description,amount\n2025-01-10,Coffee,4.50\n2025-01-15,Paycheck,-2000.00\n"
	rows, err := parse.ParseTransactionImportCSV(strings.NewReader(csv), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	if rows[0].Description != "Coffee" {
		t.Errorf("row 0 Description = %q, want %q", rows[0].Description, "Coffee")
	}
	if rows[0].Amount != 4.50 {
		t.Errorf("row 0 Amount = %f, want 4.50", rows[0].Amount)
	}
	if rows[1].Amount != -2000.00 {
		t.Errorf("row 1 Amount = %f, want -2000.00", rows[1].Amount)
	}
	if len(rows[0].Errors) != 0 {
		t.Errorf("row 0 unexpected errors: %v", rows[0].Errors)
	}
}

func TestParseTransactionImportCSV_AccountResolution(t *testing.T) {
	csv := "date,description,amount,account_name\n2025-02-01,Rent,1500.00,Chase Checking\n2025-02-02,Gas,40.00,Unknown Bank\n"
	rows, err := parse.ParseTransactionImportCSV(strings.NewReader(csv), testAccounts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows[0].AccountID == nil || *rows[0].AccountID != 1 {
		t.Errorf("row 0 AccountID = %v, want 1", rows[0].AccountID)
	}
	if rows[1].AccountID != nil {
		t.Errorf("row 1 AccountID = %v, want nil (unmatched)", rows[1].AccountID)
	}
	if rows[1].AccountName != "Unknown Bank" {
		t.Errorf("row 1 AccountName = %q, want %q", rows[1].AccountName, "Unknown Bank")
	}
}

func TestParseTransactionImportCSV_AccountResolutionCaseInsensitive(t *testing.T) {
	csv := "date,description,amount,account_name\n2025-03-01,Coffee,4.50,chase checking\n"
	rows, err := parse.ParseTransactionImportCSV(strings.NewReader(csv), testAccounts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows[0].AccountID == nil || *rows[0].AccountID != 1 {
		t.Errorf("expected case-insensitive match, got AccountID = %v", rows[0].AccountID)
	}
}

func TestParseTransactionImportCSV_OptionalColumns(t *testing.T) {
	csv := "date,description,amount,merchant_name\n2025-04-01,Amazon,29.99,Amazon.com\n"
	rows, err := parse.ParseTransactionImportCSV(strings.NewReader(csv), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows[0].MerchantName == nil || *rows[0].MerchantName != "Amazon.com" {
		t.Errorf("MerchantName = %v, want %q", rows[0].MerchantName, "Amazon.com")
	}
}

func TestParseTransactionImportCSV_QuotedFields(t *testing.T) {
	csv := "date,description,amount\n2025-05-01,\"Coffee, Shop\",4.50\n"
	rows, err := parse.ParseTransactionImportCSV(strings.NewReader(csv), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows[0].Description != "Coffee, Shop" {
		t.Errorf("Description = %q, want %q", rows[0].Description, "Coffee, Shop")
	}
}

func TestParseTransactionImportCSV_RowErrors_MissingDescription(t *testing.T) {
	csv := "date,description,amount\n2025-06-01,,10.00\n"
	rows, err := parse.ParseTransactionImportCSV(strings.NewReader(csv), nil)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if len(rows[0].Errors) == 0 {
		t.Error("expected row error for missing description")
	}
}

func TestParseTransactionImportCSV_RowErrors_InvalidDate(t *testing.T) {
	csv := "date,description,amount\n01/10/2025,Coffee,4.50\n"
	rows, err := parse.ParseTransactionImportCSV(strings.NewReader(csv), nil)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if len(rows[0].Errors) == 0 {
		t.Error("expected row error for invalid date format")
	}
}

func TestParseTransactionImportCSV_RowErrors_InvalidAmount(t *testing.T) {
	csv := "date,description,amount\n2025-06-01,Coffee,not-a-number\n"
	rows, err := parse.ParseTransactionImportCSV(strings.NewReader(csv), nil)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if len(rows[0].Errors) == 0 {
		t.Error("expected row error for invalid amount")
	}
}

func TestParseTransactionImportCSV_MissingRequiredColumn(t *testing.T) {
	csv := "date,amount\n2025-07-01,10.00\n"
	_, err := parse.ParseTransactionImportCSV(strings.NewReader(csv), nil)
	if err == nil {
		t.Error("expected error for missing required column 'description'")
	}
}

func TestParseTransactionImportCSV_EmptyFile(t *testing.T) {
	_, err := parse.ParseTransactionImportCSV(strings.NewReader(""), nil)
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestParseTransactionImportCSV_HeaderOnlyNoDataRows(t *testing.T) {
	_, err := parse.ParseTransactionImportCSV(strings.NewReader("date,description,amount\n"), nil)
	if err == nil {
		t.Error("expected error for file with no data rows")
	}
}

func TestParseTransactionImportCSV_MixedValidAndInvalidRows(t *testing.T) {
	csv := "date,description,amount\n2025-08-01,Valid,10.00\n,Missing date and desc,\n2025-08-03,Also Valid,5.00\n"
	rows, err := parse.ParseTransactionImportCSV(strings.NewReader(csv), nil)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3", len(rows))
	}
	if len(rows[0].Errors) != 0 {
		t.Errorf("row 1 should be valid, got errors: %v", rows[0].Errors)
	}
	if len(rows[1].Errors) == 0 {
		t.Error("row 2 should have errors")
	}
	if len(rows[2].Errors) != 0 {
		t.Errorf("row 3 should be valid, got errors: %v", rows[2].Errors)
	}
}

// ---------------------------------------------------------------------------
// ParseTransactionImportWrite
// ---------------------------------------------------------------------------

func TestParseTransactionImportWrite_Valid(t *testing.T) {
	body := `[
		{"description":"Coffee","amount":4.50,"date":"2025-01-10"},
		{"description":"Paycheck","amount":-2000.00,"date":"2025-01-15","merchant_name":"Acme Corp"}
	]`
	rows, err := parse.ParseTransactionImportWrite(makeRequest(t, body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	if rows[0].Description != "Coffee" {
		t.Errorf("row 0 Description = %q, want %q", rows[0].Description, "Coffee")
	}
	if rows[1].MerchantName == nil || *rows[1].MerchantName != "Acme Corp" {
		t.Errorf("row 1 MerchantName = %v, want %q", rows[1].MerchantName, "Acme Corp")
	}
}

func TestParseTransactionImportWrite_EmptyArray(t *testing.T) {
	_, err := parse.ParseTransactionImportWrite(makeRequest(t, `[]`))
	if err == nil {
		t.Error("expected error for empty array, got nil")
	}
}

func TestParseTransactionImportWrite_MissingDescription(t *testing.T) {
	body := `[{"amount":10.00,"date":"2025-01-01"}]`
	_, err := parse.ParseTransactionImportWrite(makeRequest(t, body))
	if err == nil {
		t.Error("expected error for missing description, got nil")
	}
}

func TestParseTransactionImportWrite_InvalidDate(t *testing.T) {
	body := `[{"description":"Coffee","amount":4.50,"date":"01/10/2025"}]`
	_, err := parse.ParseTransactionImportWrite(makeRequest(t, body))
	if err == nil {
		t.Error("expected error for invalid date format, got nil")
	}
}

func TestParseTransactionImportWrite_SecondRowInvalid(t *testing.T) {
	body := `[
		{"description":"Valid","amount":10.00,"date":"2025-03-01"},
		{"description":"Bad date","amount":5.00,"date":"not-a-date"}
	]`
	_, err := parse.ParseTransactionImportWrite(makeRequest(t, body))
	if err == nil {
		t.Error("expected error for invalid row in batch, got nil")
	}
	if !strings.Contains(err.Error(), "row 2") {
		t.Errorf("error should mention row number, got: %v", err)
	}
}

func TestParseTransactionImportWrite_InvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`not json`)))
	_, err := parse.ParseTransactionImportWrite(r)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
