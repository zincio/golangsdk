package golangsdk

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
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

func NewZinc(clientToken string) (*Zinc, error) {
	z := Zinc{
		ClientToken: clientToken,
		ZincBaseURL: zincBaseURL,
	}
	return &z, nil
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
	Retailer           string              `json:"retailer"`
	Epids              []ExternalProductId `json:"epids"`
	ProductDetails     []string            `json:"product_details"`
	Title              string              `json:"title"`
	VariantSpecifics   []VariantSpecific   `json:"variant_specifics"`
	ProductId          string              `json:"product_id"`
	MainImage          string              `json:"main_image"`
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
	Message string `json:"message"`
}

type ProductOptions struct {
	MaxAge    int           `json:"max_age"`
	NewerThan time.Time     `json:"newer_than"`
	Timeout   time.Duration `json:"timeout"`
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

	respBody, err := z.sendGetRequest(requestPath, options.Timeout)
	if err != nil {
		return nil, err
	}
	var resp ProductOffersResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}
	if resp.Status == "failed" {
		return &resp, fmt.Errorf("Zinc API returned status 'failed' response=%v", resp)
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
	requestPath := fmt.Sprintf("%v/products/%v?%v", z.ZincBaseURL, productId, values.Encode())

	respBody, err := z.sendGetRequest(requestPath, options.Timeout)
	if err != nil {
		return nil, err
	}
	var resp ProductDetailsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}
	if resp.Status == "failed" {
		return &resp, fmt.Errorf("Zinc API returned status 'failed' response=%v", resp)
	}
	return &resp, nil
}

func (z Zinc) sendGetRequest(requestPath string, timeout time.Duration) ([]byte, error) {
	httpReq, err := http.NewRequest("GET", requestPath, nil)
	if err != nil {
		return nil, err
	}
	httpReq.SetBasicAuth(z.ClientToken, "")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}
