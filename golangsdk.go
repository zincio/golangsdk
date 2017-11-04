package golangsdk

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	zincBaseURL = "https://api.zinc.io/v1"
)

type Retailer string

const (
	Amazon     Retailer = "amazon"
	AmazonUK   Retailer = "amazon_uk"
	AmazonCA   Retailer = "amazon_ca"
	AmazonMX   Retailer = "amazon_mx"
	Walmart    Retailer = "walmart"
	Aliexpress Retailer = "aliexpress"
)

var DefaultProductOptions = ProductOptions{
	Timeout: time.Duration(time.Second * 90),
}

type Zinc struct {
	ClientToken string
	ZincBaseURL string
}

func GetRetailer(retailer string) (Retailer, error) {
	switch retailer {
	case "amazon":
		return Amazon, nil
	case "amazon_uk":
		return AmazonUK, nil
	case "amazon_ca":
		return AmazonCA, nil
	case "amazon_mx":
		return AmazonMX, nil
	case "walmart":
		return Walmart, nil
	case "aliexpress":
		return Aliexpress, nil
	default:
		return Amazon, fmt.Errorf("Invalid retailer string")
	}
}

func NewZinc(clientToken string) (*Zinc, error) {
	z := Zinc{
		ClientToken: clientToken,
		ZincBaseURL: zincBaseURL,
	}
	return &z, nil
}

type OrderRequest struct {
	Retailer            Retailer             `json:"retailer"`
	Products            []Product            `json:"products"`
	ShippingMethod      string               `json:"shipping_method,omitempty"`
	Shipping            *Shipping            `json:"shipping,omitempty"`
	ShippingAddress     *Address             `json:"shipping_address"`
	BillingAddress      *Address             `json:"billing_address,omitempty"`
	PaymentMethod       *PaymentMethod       `json:"payment_method,omitempty"`
	RetailerCredentials *RetailerCredentials `json:"retailer_credentials,omitempty"`
	GiftMessage         string               `json:"gift_message,omitempty"`
	IsGift              bool                 `json:"is_gift"`
	MaxPrice            int                  `json:"max_price"`
	Webhooks            *Webhooks            `json:"webhooks,omitempty"`
	Bundled             bool                 `json:"bundled"`
	Addax               bool                 `json:"addax"`
}

type Product struct {
	ProductId               string                   `json:"product_id"`
	Quantity                int                      `json:"quantity"`
	SellerSelectionCriteria *SellerSelectionCriteria `json:"seller_selection_criteria,omitempty"`
}

type Shipping struct {
	OrderBy  string `json:"order_by,omitempty"`
	MaxDays  int    `json:"max_days"`
	MaxPrice int    `json:"max_price"`
}

type Address struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	ZipCode      string `json:"zip_code"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	PhoneNumber  string `json:"phone_number"`
}

type PaymentMethod struct {
	NameOnCard      string `json:"name_on_card,omitempty"`
	Number          string `json:"number,omitempty"`
	SecurityCode    string `json:"security_code,omitempty"`
	ExpirationMonth int    `json:"expiration_month,omitempty"`
	ExpirationYear  int    `json:"expiration_year,omitempty"`
	UseGift         bool   `json:"use_gift"`
}

type RetailerCredentials struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	VerificationCode string `json:"verification_code,omitempty"`
}

type Webhooks struct {
	RequestSucceeded string `json:"request_succeeded"`
	RequestFailed    string `json:"request_failed"`
	TrackingObtained string `json:"tracking_obtained"`
	StatusUpdated    string `json:"status_updated"`
}

type SellerSelectionCriteria struct {
	Prime bool `json:"prime"`
}

type OrderResponse struct {
	RequestId        string            `json:"request_id"`
	Type             string            `json:"_type"`
	Code             string            `json:"code"`
	Data             ErrorDataResponse `json:"data"`
	ErrorMessage     string            `json:"message"`
	PriceComponents  PriceComponents   `json:"price_components"`
	MerchantOrderIds []MerchantOrderId `json:"merchant_order_ids"`
	Tracking         []Tracking        `json:"tracking"`
	Request          OrderRequest      `json:"request"`
}

type PriceComponents struct {
	Shipping int `json:"shipping"`
	Subtotal int `json:"subtotal"`
	Tax      int `json:"tax"`
	Total    int `json:"total"`
}

type MerchantOrderId struct {
	MerchantOrderId string    `json:"merchant_order_id"`
	Merchant        string    `json:"merchant"`
	Account         string    `json:"account"`
	PlacedAt        time.Time `json:"placed_at"`
}

type Tracking struct {
	MerchantOrderId string    `json:"merchant_order_id"`
	ObtainedAt      time.Time `json:"obtained_at"`
	Carrier         string    `json:"carrier"`
	TrackingNumber  string    `json:"tracking_number"`
	ProductIds      []string  `json:"product_ids"`
	TrackingURL     string    `json:"tracking_url"`
}

type ProductOffersResponse struct {
	Code     string            `json:"code"`
	Data     ErrorDataResponse `json:"data"`
	Status   string            `json:"status"`
	Retailer string            `json:"retailer"`
	Offers   []ProductOffer    `json:"offers"`
}

type ProductOffer struct {
	Available            bool             `json:"available"`
	Addon                bool             `json:"addon"`
	Condition            string           `json:"condition"`
	ShippingOptions      []ShippingOption `json:"shipping_options"`
	HandlingDays         HandlingDays     `json:"handling_days"`
	PrimeOnly            bool             `json:"prime_only"`
	MarketplaceFulfilled bool             `json:"marketplace_fulfilled"`
	Currency             string           `json:"currency"`
	Seller               Seller           `json:"seller"`
	BuyBoxWinner         bool             `json:"buy_box_winner"`
	International        bool             `json:"international"`
	OfferId              string           `json:"offer_id"`
	Price                int              `json:"price"`
}

type ShippingOption struct {
	Price int `json:"price"`
}

type HandlingDays struct {
	Max int `json:"max"`
	Min int `json:"min"`
}

type Seller struct {
	NumRatings      int    `json:"num_ratings"`
	PercentPositive int    `json:"percent_positive"`
	FirstParty      bool   `json:"first_party"`
	Name            string `json:"name"`
	Id              string `json:"id"`
}

type ProductDetailsResponse struct {
	Code               string              `json:"code"`
	Data               ErrorDataResponse   `json:"data"`
	Status             string              `json:"status"`
	ProductDescription string              `json:"product_description"`
	PostDescription    string              `json:"post_description"`
	Retailer           string              `json:"retailer"`
	Epids              []ExternalProductId `json:"epids"`
	ProductDetails     []string            `json:"product_details"`
	Title              string              `json:"title"`
	VariantSpecifics   []VariantSpecific   `json:"variant_specifics"`
	ProductId          string              `json:"product_id"`
	MainImage          string              `json:"main_image"`
	Brand              string              `json:"brand"`
	MPN                string              `json:"mpn"`
	Images             []string            `json:"images"`
	FeatureBullets     []string            `json:"feature_bullets"`
}

type ExternalProductId struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type VariantSpecific struct {
	Dimension string `json:"dimension"`
	Value     string `json:"value"`
}

type ErrorDataResponse struct {
	Message         string           `json:"message"`
	ValidatorErrors []ValidatorError `json:"validator_errors"`
	AllVariants     []Variant        `json:"all_variants"`
}

type ValidatorError struct {
	Message string `json:"message"`
	Path    string `json:"path"`
	Value   string `json:"value"`
}

type Variant struct {
	VariantSpecifics []VariantSpecific `json:"variant_specifics"`
	ProductId        string            `json:"product_id"`
}

type ProductOptions struct {
	MaxAge    int           `json:"max_age"`
	Priority  int           `json:"priority"`
	NewerThan time.Time     `json:"newer_than"`
	Timeout   time.Duration `json:"timeout"`
}

type ZincError struct {
	ErrorMessage string            `json:"error"`
	Code         string            `json:"code"`
	Data         ErrorDataResponse `json:"data"`
}

func (z ZincError) Error() string {
	return z.ErrorMessage
}

func SimpleError(errorStr string) ZincError {
	return ZincError{ErrorMessage: errorStr}
}

func (z Zinc) GetProductInfo(productId string, retailer Retailer, options ProductOptions) (*ProductOffersResponse, *ProductDetailsResponse, error) {
	offersChan := make(chan *ProductOffersResponse, 1)
	detailsChan := make(chan *ProductDetailsResponse, 1)
	errorsChan := make(chan error, 2)

	go func() {
		offers, err := z.GetProductOffers(productId, retailer, options)
		errorsChan <- err
		offersChan <- offers
	}()

	go func() {
		details, err := z.GetProductDetails(productId, retailer, options)
		errorsChan <- err
		detailsChan <- details
	}()

	offers := <-offersChan
	details := <-detailsChan
	for i := 0; i < 2; i++ {
		err := <-errorsChan
		if err != nil {
			return nil, nil, err
		}
	}
	return offers, details, nil
}

func (z Zinc) SendOrder(order OrderRequest) (*OrderResponse, error) {
	requestPath := fmt.Sprintf("%v/orders", z.ZincBaseURL)
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(order); err != nil {
		return nil, SimpleError(err.Error())
	}
	var resp OrderResponse
	if err := z.SendRequest("POST", requestPath, body, time.Duration(time.Second*30), &resp); err != nil {
		return nil, SimpleError(err.Error())
	}
	return &resp, nil
}

func (z Zinc) GetProductOffers(productId string, retailer Retailer, options ProductOptions) (*ProductOffersResponse, error) {
	values := url.Values{}
	values.Set("retailer", string(retailer))
	values.Set("version", "2")
	if options.MaxAge != 0 {
		values.Set("max_age", strconv.Itoa(options.MaxAge))
	}
	if !options.NewerThan.IsZero() {
		values.Set("newer_than", strconv.FormatInt(options.NewerThan.Unix(), 10))
	}
	requestPath := fmt.Sprintf("%v/products/%v/offers?%v", z.ZincBaseURL, productId, values.Encode())

	var resp ProductOffersResponse
	if err := z.SendRequest("GET", requestPath, nil, options.Timeout, &resp); err != nil {
		return nil, SimpleError(err.Error())
	}
	if resp.Status == "failed" {
		msg := fmt.Sprintf("Zinc API returned status 'failed' data=%+v", resp.Data)
		return &resp, ZincError{Code: resp.Code, ErrorMessage: msg, Data: resp.Data}
	}
	return &resp, nil
}

func (z Zinc) GetProductDetails(productId string, retailer Retailer, options ProductOptions) (*ProductDetailsResponse, error) {
	values := url.Values{}
	values.Set("retailer", string(retailer))
	if options.MaxAge != 0 {
		values.Set("max_age", strconv.Itoa(options.MaxAge))
	}
	if !options.NewerThan.IsZero() {
		values.Set("newer_than", strconv.FormatInt(options.NewerThan.Unix(), 10))
	}
	if options.Priority != 0 {
		values.Set("priority", strconv.Itoa(options.Priority))
	}
	requestPath := fmt.Sprintf("%v/products/%v?%v", z.ZincBaseURL, productId, values.Encode())

	var resp ProductDetailsResponse
	if err := z.SendRequest("GET", requestPath, nil, options.Timeout, &resp); err != nil {
		return nil, SimpleError(err.Error())
	}
	if resp.Status == "failed" {
		msg := fmt.Sprintf("Zinc API returned status 'failed' data=%+v", resp.Data)
		return &resp, ZincError{Code: resp.Code, ErrorMessage: msg, Data: resp.Data}
	}
	return &resp, nil
}

func cleanRespBody(respBody []byte) []byte {
	str := string(respBody)
	i := strings.Index(str, "HTTP/1.1 200 OK")
	if i == -1 {
		return respBody
	}
	return []byte(str[:i])
}

func (z Zinc) SendRequest(method, requestPath string, body io.Reader, timeout time.Duration, resp interface{}) error {
	httpReq, err := http.NewRequest(method, requestPath, body)
	if err != nil {
		return err
	}
	httpReq.SetBasicAuth(z.ClientToken, "")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: timeout}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}
	cleanedBody := cleanRespBody(respBody)
	if err := json.Unmarshal(cleanedBody, resp); err != nil {
		log.Printf("[Golangsdk] Unable to unmarshal response request_path=%v body=%v", requestPath, string(cleanedBody))
		return SimpleError(err.Error())
	}
	return nil
}
