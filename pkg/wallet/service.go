package wallet

import (
	"errors"
	"log"
	"os"
	"strconv"
	"github.com/KomGitHub/wallet/v1/pkg/types"
	"github.com/google/uuid"
	"io"
	"strings"
)

var ErrPhoneRegistered = errors.New("phone already registered")
var ErrAmountMustBePositive = errors.New("amount must be greater then zero")
var ErrAccountNotFound = errors.New("account not found")
var ErrNotEnoughBalance = errors.New("not enough amount on balance")
var ErrPaymentNotFound = errors.New("payment not found")
var ErrFavoriteNotFound = errors.New("favorite not found")

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites []*types.Favorite
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}
	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)

	return account, nil
}

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return ErrAccountNotFound
	}

	account.Balance += amount
	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}

	if account.Balance < amount {
		return nil, ErrNotEnoughBalance
	}

	account.Balance -= amount
	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment)
	return payment, nil
}

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			return acc, nil
		}
	}
	return nil, ErrAccountNotFound
}

func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}
	return nil, ErrPaymentNotFound
}

func (s *Service) Reject(paymentID string) error {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return err
	}
	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return err
	}
	account.Balance += payment.Amount
	payment.Status = types.PaymentStatusFail
	return nil
}

func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}
	
	newPayment, err := s.Pay(payment.AccountID, payment.Amount, payment.Category)
	if err != nil {
		return nil, err
	}
	return newPayment, nil
}

func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}
	
	favoriteID := uuid.New().String()
	favorite := &types.Favorite{
		ID:        favoriteID,
		AccountID: payment.AccountID,
		Name: name,
		Amount:    payment.Amount,
		Category:  payment.Category,
	}
	s.favorites = append(s.favorites, favorite)
	return favorite, nil
}

func (s *Service) FindFavoriteByID(favoriteID string) (*types.Favorite, error) {
	for _, favorite := range s.favorites {
		if favorite.ID == favoriteID {
			return favorite, nil
		}
	}
	return nil, ErrFavoriteNotFound
}

func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	favorite, err := s.FindFavoriteByID(favoriteID)
	if err != nil {
		return nil, err
	}
	
	payment, err := s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *Service) ExportToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()
	var export string
	for _, account := range s.accounts {
		if len(export) != 0 {
			export += "|"
		}
		export += strconv.FormatInt(account.ID, 10) + ";" + string(account.Phone) + ";" + strconv.FormatInt(int64(account.Balance), 10)
	}
	_, err = file.Write([]byte(export))
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func (s *Service) ImportFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()
	content := make([]byte, 0)
	buf := make([]byte, 4096)
	for {
		read, err := file.Read(buf)
		if err == io.EOF {
			content = append(content, buf[:read]...)
			break
		}
		if err != nil {
			log.Print(err)
			return err
		}
		content = append(content, buf[:read]...)
	}
	data := string(content)
	records := strings.Split(data, "|")
	for _, record := range records {
		if len(record) != 0 {
			items := strings.Split(record, ";")
			id, err := strconv.ParseInt(items[0], 10, 64)
			if err != nil {
				log.Print(err)
				break
			}
			balance, err := strconv.ParseInt(items[2], 10, 64)
			if err != nil {
				log.Print(err)
				break
			}

			account := &types.Account{
				ID:      id,
				Phone:   types.Phone(items[1]),
				Balance: types.Money(balance),
			}
			s.accounts = append(s.accounts, account)
		}
	}

	return nil
}

func (s *Service) Export(dir string) (err error) {
	if len(s.accounts) > 0 {
		file, err := os.Create(dir + "/accounts.dump")
		if err != nil {
			return err
		}
		defer func() {
			if cerr := file.Close(); cerr != nil {
				if err == nil {
					err = cerr
				}
			}
		}()
		var export string
		for _, account := range s.accounts {
			if len(export) != 0 {
				export += "\n"
			}
			export += strconv.FormatInt(account.ID, 10) + ";" + string(account.Phone) + ";" + strconv.FormatInt(int64(account.Balance), 10)
		}
		_, err = file.Write([]byte(export))
		if err != nil {
			return err
		}
	}
	if len(s.payments) > 0 {
		file, err := os.Create(dir + "/payments.dump")
		if err != nil {
			return err
		}
		defer func() {
			if cerr := file.Close(); cerr != nil {
				if err == nil {
					err = cerr
				}
			}
		}()
		var export string
		for _, payment := range s.payments {
			if len(export) != 0 {
				export += "\n"
			}
			export += payment.ID + ";" + strconv.FormatInt(payment.AccountID, 10) + ";" + strconv.FormatInt(int64(payment.Amount), 10) + ";" + string(payment.Category) + ";" + string(payment.Status)
		}
		_, err = file.Write([]byte(export))
		if err != nil {
			return err
		}
	}
	if len(s.favorites) > 0 {
		file, err := os.Create(dir + "/favorites.dump")
		if err != nil {
			return err
		}
		defer func() {
			if cerr := file.Close(); cerr != nil {
				if err == nil {
					err = cerr
				}
			}
		}()
		var export string
		for _, favorite := range s.favorites {
			if len(export) != 0 {
				export += "\n"
			}
			export += favorite.ID + ";" + strconv.FormatInt(favorite.AccountID, 10) + ";" + string(favorite.Name) + ";" + strconv.FormatInt(int64(favorite.Amount), 10) + ";" + string(favorite.Category)
		}
		_, err = file.Write([]byte(export))
		if err != nil {
			return err
		}
	}
	return nil
}