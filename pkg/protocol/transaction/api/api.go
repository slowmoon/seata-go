package api

import (
	"context"
	"fmt"
)

import (
	"github.com/seata/seata-go/pkg/common/log"
	"github.com/seata/seata-go/pkg/protocol/seatactx"
	"github.com/seata/seata-go/pkg/protocol/transaction"
	"github.com/seata/seata-go/pkg/protocol/transaction/manager"
)

type TransactionalExecutor interface {
	Execute(ctx context.Context, param interface{}) (interface{}, error)
	GetTransactionInfo() transaction.TransactionInfo
}

func Begin(ctx context.Context, name string) context.Context {
	if !seatactx.IsSeataContext(ctx) {
		ctx = seatactx.InitSeataContext(ctx)
	}

	seatactx.SetTxName(ctx, name)
	if seatactx.GetTransactionRole(ctx) == nil {
		seatactx.SetTransactionRole(ctx, transaction.LAUNCHER)
	}

	var tx *manager.GlobalTransaction
	if seatactx.HasXID(ctx) {
		tx = &manager.GlobalTransaction{
			Xid:    seatactx.GetXID(ctx),
			Status: transaction.GlobalStatusBegin,
			Role:   transaction.PARTICIPANT,
		}
		seatactx.SetTxStatus(ctx, transaction.GlobalStatusBegin)
	}

	// todo: Handle the transaction propagation.

	if tx == nil {
		tx = &manager.GlobalTransaction{
			Xid:    seatactx.GetXID(ctx),
			Status: transaction.GlobalStatusUnKnown,
			Role:   transaction.LAUNCHER,
		}
		seatactx.SetTxStatus(ctx, transaction.GlobalStatusUnKnown)
	}

	// todo timeout should read from config
	err := manager.GetGlobalTransactionManager().Begin(ctx, tx, 50, name)
	if err != nil {
		panic(fmt.Sprintf("transactionTemplate: begin transaction failed, error %v", err))
	}

	return ctx
}

// commit global transaction
func CommitOrRollback(ctx context.Context, err error) error {
	tx := &manager.GlobalTransaction{
		Xid:    seatactx.GetXID(ctx),
		Status: *seatactx.GetTxStatus(ctx),
		Role:   *seatactx.GetTransactionRole(ctx),
	}

	var resp error
	if err == nil {
		resp = manager.GetGlobalTransactionManager().Commit(ctx, tx)
		if resp != nil {
			log.Infof("transactionTemplate: commit transaction failed, error %v", err)
		}
	} else {
		resp = manager.GetGlobalTransactionManager().Rollback(ctx, tx)
		if resp != nil {
			log.Infof("transactionTemplate: Rollback transaction failed, error %v", err)
		}
	}
	return resp
}