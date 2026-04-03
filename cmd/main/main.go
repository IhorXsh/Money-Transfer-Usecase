package main

import (
	"context"
	"fmt"
	"log"

	"github.com/IhorXsh/Money-Transfer-Usecase/domain"
	"github.com/IhorXsh/Money-Transfer-Usecase/repository"
	"github.com/IhorXsh/Money-Transfer-Usecase/usecases/transfer"
)

func main() {
	accounts := map[domain.AccountID]*domain.Account{
		"a1": domain.NewAccount("a1", 100, domain.AccountStatusActive),
		"a2": domain.NewAccount("a2", 50, domain.AccountStatusActive),
	}
	repo := repository.NewAccountRepo(accounts)
	uc := transfer.NewInteractor(repo)

	req := &transfer.TransferRequest{
		FromAccountID: "a1",
		ToAccountID:   "a2",
		Amount:        25,
	}

	plan, err := uc.Execute(context.Background(), req)
	if err != nil {
		log.Fatalf("transfer failed: %v", err)
	}

	fmt.Printf("transfer plan: %+v\n", plan.Mutations())
	fmt.Printf("balances: a1=%d, a2=%d\n", accounts["a1"].Balance, accounts["a2"].Balance)
}
