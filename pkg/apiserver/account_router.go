package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/fkondej/go-showcase/v1/pkg/apiclient"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type accountRouter struct {
	accountService *AccountService
	logger         *log.Logger
}

func (ar *accountRouter) getMultipleAccounts(c *gin.Context) {
	var err error
	page := apiclient.AccountPage{}

	if page.PageNumber, err = strconv.Atoi(c.DefaultQuery("page[number]", "0")); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Wrong value in page[number] query parameter"})
		return
	}
	if page.PageSize, err = strconv.Atoi(c.DefaultQuery("page[size]", "100")); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Wrong value in page[size] query parameter"})
		return
	}
	accountNumber := c.Query("filter[account_number]")
	if len(accountNumber) > 0 {
		page.Filter.AccountNumber = strings.Split(accountNumber, ",")
	}
	bankID := c.Query("filter[bank_id]")
	if len(bankID) > 0 {
		page.Filter.BankID = strings.Split(bankID, ",")
	}
	bankIDCode := c.Query("filter[bank_id_code]")
	if len(bankIDCode) > 0 {
		page.Filter.BankIDCode = strings.Split(bankIDCode, ",")
	}
	country := c.Query("filter[country]")
	if len(country) > 0 {
		page.Filter.Country = strings.Split(country, ",")
	}
	customerID := c.Query("filter[customer_id]")
	if len(customerID) > 0 {
		page.Filter.CustomerID = strings.Split(customerID, ",")
	}
	iban := c.Query("filter[iban]")
	if len(iban) > 0 {
		page.Filter.IBAN = strings.Split(iban, ",")
	}
	accountList, err := ar.accountService.getAccountList(page)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": fmt.Sprintf("Problem getting accounts from storage %v", err)})
		return
	}
	c.JSON(200, gin.H{
		"data": accountList,
	})
}

func (ar *accountRouter) getOneAccount(c *gin.Context) {
	accountID := c.Param("accountId")
	data, err := ar.accountService.getAccount(accountID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": fmt.Sprintf("%v", err)})
		return
	}
	status := http.StatusOK
	if data == nil {
		status = http.StatusNotFound
	}
	c.JSON(status, gin.H{
		"data": data,
	})
}

func (ar *accountRouter) createAccount(c *gin.Context) {
	data := apiclient.CreateAccountResourceRequestData{}
	if err := c.ShouldBindBodyWith(&data, binding.JSON); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": fmt.Sprintf("Wrong request body %v", err)})
		return
	}
	err := ar.accountService.upsertAccount(data)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": fmt.Sprintf("Problem creating account in storage %v", err)})
		return
	}
	newData, err := ar.accountService.getAccount(data.Data.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": fmt.Sprintf("%v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   newData,
	})
}

func (ar *accountRouter) deleteAccount(c *gin.Context) {
	var (
		version    int
		err        error
		statusCode int
	)
	accountID := c.Param("accountId")
	if version, err = strconv.Atoi(c.Query("version")); err != nil {
		ar.logger.Printf("deleteAccount, wrong version %v: %v, FAILED", c.Query("version"), err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Wrong value in version query parameter"})
		return
	}
	deleted, err := ar.accountService.deleteAccount(accountID, version)
	if err != nil {
		ar.logger.Printf("deleteAccount, %v, FAILED", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Problem deleting account from storage"})
		return
	}
	if deleted {
		statusCode = http.StatusNoContent
	} else {
		statusCode = http.StatusNotFound
	}
	c.JSON(statusCode, gin.H{
		"status": "success",
	})
}

func SetupAccountRouting(router *gin.RouterGroup, accountService *AccountService, logger *log.Logger) {
	ar := accountRouter{
		accountService: accountService,
		logger:         logger,
	}
	router.GET("/", ar.getMultipleAccounts)
	router.GET("/:accountId", ar.getOneAccount)
	router.POST("/", ar.createAccount)
	router.DELETE("/:accountId", ar.deleteAccount)
}
