package payment

import (
	"done-hub/model"
	"done-hub/payment/gateway/alipay"
	"done-hub/payment/gateway/epay"
	"done-hub/payment/gateway/stripe"
	"done-hub/payment/gateway/wxpay"
	"done-hub/payment/types"

	"github.com/gin-gonic/gin"
)

type PaymentProcessor interface {
	Name() string
	Pay(config *types.PayConfig, gatewayConfig string) (*types.PayRequest, error)
	CreatedPay(notifyURL string, gatewayConfig *model.Payment) error
	HandleCallback(c *gin.Context, gatewayConfig string) (*types.PayNotify, error)
}

var Gateways = make(map[string]PaymentProcessor)

func init() {
	Gateways["epay"] = &epay.Epay{}
	Gateways["alipay"] = &alipay.Alipay{}
	Gateways["wxpay"] = &wxpay.WeChatPay{}
	Gateways["stripe"] = &stripe.Stripe{}
}
