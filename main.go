package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"time"
)

const (
	FILE_PATH    = "accounts.json"
	MIN_TX_COUNT = 10_000
	MIN_ICP_BAL  = 250_000
	ACTIVE_THR   = 60 * 60 * 24 * 30 // 30 days

	E8S      = 1_00_000_000
	DELAY_MS = 100
)

type Response struct {
	Total    int       `json:"total"`
	Accounts []Account `json:"accounts"`
}

type Account struct {
	Active            bool   `json:"active"`
	LargeBalance      bool   `json:"large_balance"`
	Name              string `json:"name,omitempty"`
	AccountIdentifier string `json:"account_identifier"`
	Balance           string `json:"balance"`
	TransactionCount  string `json:"transaction_count"`
	UpdatedAt         int64  `json:"updated_at"`
}

func (a *Account) UnmarshalJSON(data []byte) error {
	var account struct {
		AccountIdentifier string `json:"account_identifier"`
		Balance           string `json:"balance"`
		TransactionCount  string `json:"transaction_count"`
		UpdatedAt         int64  `json:"updated_at"`
	}
	if err := json.Unmarshal(data, &account); err != nil {
		return err
	}
	if name, ok := names[account.AccountIdentifier]; ok {
		a.Name = name
	}
	now := time.Now().Unix()
	a.UpdatedAt = account.UpdatedAt
	a.Active = now-a.UpdatedAt < ACTIVE_THR
	a.AccountIdentifier = account.AccountIdentifier
	a.Balance = account.Balance
	a.LargeBalance = 0 < a.BalanceICP().Cmp(big.NewInt(MIN_ICP_BAL))
	a.TransactionCount = account.TransactionCount
	return nil
}

func (a Account) Count() *big.Int {
	count, _ := new(big.Int).SetString(a.TransactionCount, 10)
	return count
}

func (a Account) BalanceICP() *big.Int {
	balance, _ := new(big.Int).SetString(a.Balance, 10)
	return new(big.Int).Div(balance, big.NewInt(E8S))
}

// Source: https://forum.dfinity.org/t/exchange-hot-wallets/43109
var names = map[string]string{
	"bad030b417484232fd2019cb89096feea3fdd3d9eb39e1d07bcb9a13c7673464": "Bitget",
	"609d3e1e45103a82adc97d4f88c51f78dedb25701e8e51e8c4fec53448aadc29": "Binance 1",
	"220c3a33f90601896e26f76fa619fe288742df1fa75426edfaf759d39f2455a5": "Binance 2",
	"d3e13d4777e22367532053190b6c6ccf57444a61337e996242b1abfb52cf92c8": "Binance 3",
	"acd76fff0536f863d9dd4b326a1435466f82305758b4b1b4f62ff9fa81c14073": "Bybit",
	"449ce7ad1298e2ed2781ed379aba25efc2748d14c60ede190ad7621724b9e8b2": "Coinbase 1",
	"4dfa940def17f1427ae47378c440f10185867677109a02bc8374fc25b9dee8af": "Coinbase 2",
	"dd15f3040edab88d2e277f9d2fa5cc11616ebf1442279092e37924ab7cce8a74": "Coinbase 3",
	"a6ed987d89796f921c8a49d275ec7c9aa04e75a8fc8cd2dbaa5da799f0215ab0": "Coinbase (Inactive 2021) 1",
	"660b1680dafeedaa68c1f1f4cf8af42ed1dfb8564646efe935a2b9a48528b605": "Coinbase (Inactive 2021) 2",
	"4878d23a09b554157b31323004e1cc053567671426ca4eec7b7e835db607b965": "Coinbase (Inactive 2021) 3",
	"8fe706db7b08f957a15199e07761039a7718937aabcc0fe48bc380a4daf9afb0": "Gate.io",
	"935b1a3adc28fd68cacc95afcdec62e985244ce0cfbbb12cdc7d0b8d198b416d": "HTX",
	"040834c30cdf5d7a13aae8b57d94ae2d07eefe2bc3edd8cf88298730857ac2eb": "Kraken",
	"efa01544f509c56dd85449edf2381244a48fad1ede5183836229c00ab00d52df": "KuCoin 1",
	"00c3df112e62ad353b7cc7bf8ad8ce2fec8f5e633f1733834bf71e40b250c685": "KuCoin 2",
	"9e62737aab36f0baffc1faac9edd92a99279723eb3feb2e916fa99bb7fe54b59": "MEXC",
	"e7a879ea563d273c46dd28c1584eaa132fad6f3e316615b3eb657d067f3519b5": "OKX 1",
	"d2c6135510eaf107bdc2128ef5962c7db2ae840efdf95b9395cdaf4983942978": "OKX 2",
	"178197f9833164374be1e0ff8e9cf8b78c964f3ea294ab0da9bddc800c7ac64f": "Unknown 1",
	"da29b27beb16a842882149b5380ff3b20f701c33ca8fddbecdb5201c600e0f0e": "Unknown 2",
}

func getAccounts(u string) ([]Account, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %s", resp.Status)
	}
	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response.Accounts, nil
}

func main() {
	enableLargeBalance := flag.Bool("enable-large-balance", false, "Enable adding accounts with large balances")
	flag.Parse()

	accounts := make(map[string]Account)

tx:
	for offset := 0; ; offset += 50 {
		accountList, err := getAccounts(fmt.Sprintf("https://ledger-api.internetcomputer.org/accounts?limit=50&sort_by=-transaction_count&offset=%d", offset))
		if err != nil {
			fmt.Println(err)
			break
		}
		for _, account := range accountList {
			count := account.Count()
			if 0 < count.Cmp(big.NewInt(MIN_TX_COUNT)) {
				accounts[account.AccountIdentifier] = account
			} else {
				break tx
			}
		}
		time.Sleep(DELAY_MS * time.Millisecond)
	}

	if *enableLargeBalance {
	bs:
		for offset := 0; ; offset += 50 {
			accountList, err := getAccounts(fmt.Sprintf("https://ledger-api.internetcomputer.org/accounts?limit=50&sort_by=-balance&offset=%d", offset))
			if err != nil {
				fmt.Println(err)
				break
			}
			for _, account := range accountList {
				balance := account.BalanceICP()
				if 0 < balance.Cmp(big.NewInt(MIN_ICP_BAL)) {
					accounts[account.AccountIdentifier] = account
				} else {
					break bs
				}
			}
			time.Sleep(DELAY_MS * time.Millisecond)
		}
	}

	fmt.Printf("Found %d accounts matching parameters.\n", len(accounts))
	raw, err := json.MarshalIndent(accounts, "", "\t")
	if err != nil {
		fmt.Println(err)
	}
	if err := os.WriteFile(FILE_PATH, raw, 0644); err != nil {
		fmt.Println(err)
	}
}
