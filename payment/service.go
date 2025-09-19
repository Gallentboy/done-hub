package payment

import (
	"done-hub/common/config"
	"done-hub/common/logger"
	"done-hub/model"
	"done-hub/payment/types"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

type PaymentService struct {
	Payment *model.Payment
	gateway PaymentProcessor
}

type PayMoney struct {
	Amount   float64
	Currency model.CurrencyType
}

func NewPaymentService(uuid string) (*PaymentService, error) {
	payment, err := model.GetPaymentByUUID(uuid)
	if err != nil {
		return nil, errors.New("payment not found")
	}

	gateway, ok := Gateways[payment.Type]
	if !ok {
		return nil, errors.New("payment gateway not found")
	}

	return &PaymentService{
		Payment: payment,
		gateway: gateway,
	}, nil
}

func (s *PaymentService) CreatedPay() error {
	notifyURL := s.getNotifyURL()
	return s.gateway.CreatedPay(notifyURL, s.Payment)
}

func (s *PaymentService) Pay(tradeNo string, amount float64, user *model.User) (*types.PayRequest, error) {
	config := &types.PayConfig{
		Money:     amount,
		TradeNo:   tradeNo,
		NotifyURL: s.getNotifyURL(),
		ReturnURL: s.getReturnURL(),
		Currency:  s.Payment.Currency,
		User:      user,
	}
	payRequest, err := s.gateway.Pay(config, s.Payment.Config)
	if err != nil {
		return nil, err
	}

	return payRequest, nil
}

func (s *PaymentService) HandleCallback(c *gin.Context, gatewayConfig string) (*types.PayNotify, error) {
	payNotify, err := s.gateway.HandleCallback(c, gatewayConfig)
	if err != nil {
		logger.SysError(fmt.Sprintf("%s payment callback error: %v", s.gateway.Name(), err))

	}

	return payNotify, err
}

func (s *PaymentService) getNotifyURL() string {
	var notifyDomain string

	// 优先使用全局支付回调地址配置
	if config.PaymentCallbackAddress != "" {
		notifyDomain = config.PaymentCallbackAddress
	} else if s.Payment.NotifyDomain != "" {
		// 其次使用支付网关的回调域名配置
		notifyDomain = s.Payment.NotifyDomain
	} else {
		// 最后使用服务器地址
		notifyDomain = config.ServerAddress
	}

	notifyDomain = strings.TrimSuffix(notifyDomain, "/")
	return fmt.Sprintf("%s/api/payment/notify/%s", notifyDomain, s.Payment.UUID)
}

func (s *PaymentService) getReturnURL() string {
	serverAdd := strings.TrimSuffix(config.ServerAddress, "/")
	return fmt.Sprintf("%s/panel/log", serverAdd)
}
