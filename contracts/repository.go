package contracts

import (
	"context"

	"github.com/IhorXsh/Money-Transfer-Usecase/domain"
)

type AccountRepository interface {
	Retrieve(ctx context.Context, id domain.AccountID) (*domain.Account, error)
	UpdateMut(account *domain.Account) *Mutation
}

type Mutation struct {
	Table   string
	ID      string
	Updates map[string]interface{}
}

type Plan struct {
	mutations []*Mutation
}

func NewPlan() *Plan { return &Plan{} }

func (p *Plan) Add(m *Mutation) {
	if m != nil {
		p.mutations = append(p.mutations, m)
	}
}

func (p *Plan) Mutations() []*Mutation {
	return append([]*Mutation(nil), p.mutations...)
}
