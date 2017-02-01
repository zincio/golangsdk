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
	MaxAge    int       `json:"max_age"`
	NewerThan time.Time `json:"newer_than"`
}

func (z Zinc) GetProductDetails(productId string, retailer Retailer, options ProductOptions) (*ProductDetailsResponse, error) {
	values := url.Values{}
	values.Set("retailer", retailer)
	if options.MaxAge != 0 {
		values.Set("max_age", strconv.Itoa(options.MaxAge))
	}
	if !options.NewerThan.IsZero() {
		values.Set("newer_than", strconv.FormatInt(options.NewerThan.Unix(), 10))
	}
	requestPath := fmt.Sprintf("%v/products/%v?%v", z.ZincBaseURL, productId, values.Encode())

	respBody, err := z.sendGetRequest(requestPath)
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

func (z Zinc) sendGetRequest(requestPath string) ([]byte, error) {
	httpReq, err := http.NewRequest("GET", requestPath, nil)
	if err != nil {
		return nil, err
	}
	httpReq.SetBasicAuth(z.ClientToken, "")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
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
