package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	db "sstonn/db/sqlc"
)

type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD EUR VND"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	args := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: db.Currency(req.Currency),
		Balance:  0,
	}

	if account, err := server.store.CreateAccount(ctx, args); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	} else {
		ctx.JSON(http.StatusCreated, account)
	}
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if account, err := server.store.GetAccount(ctx, req.ID); err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	} else {
		ctx.JSON(http.StatusOK, account)
	}

}

type listAccountsRequest struct {
	Page     int64 `form:"page,default=1" binding:"min=1"`
	PageSize int64 `form:"pageSize,default=10" binding:"min=5,max=20"`
}

func (server *Server) listAccounts(ctx *gin.Context) {
	var req listAccountsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	accountsChan := make(chan []db.Account, 1)
	accountsErrChan := make(chan error, 1)
	totalChan := make(chan int64, 1)
	totalErrChan := make(chan error, 1)

	go func() {
		accounts, err := server.store.ListAccounts(ctx, db.ListAccountsParams{
			Limit:  req.PageSize,
			Offset: (req.Page - 1) * req.PageSize,
		})
		if err != nil {
			accountsChan <- nil
			accountsErrChan <- err
			return
		}
		accountsChan <- accounts
		accountsErrChan <- nil
	}()

	go func() {
		total, err := server.store.CountAccounts(ctx)
		if err != nil {
			totalChan <- 0
			totalErrChan <- err
			return
		}
		totalChan <- total
		totalErrChan <- nil
	}()

	accounts := <-accountsChan
	accountsErr := <-accountsErrChan
	total := <-totalChan
	totalErr := <-totalErrChan

	if accountsErr != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(accountsErr))
		return
	}
	if totalErr != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(totalErr))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"accounts": accounts, "page": req.Page, "pageSize": req.PageSize, "total": total})
}
