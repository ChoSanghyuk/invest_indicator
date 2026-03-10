# API Documentation

## Overview
This document describes the REST API endpoints for the Investment Indicator application.

## Data Types

### Basic Model Types

#### User
```go
type User struct {
    ID       int
    Username string
    Email    string
    Password string
    IsAdmin  bool
}
```

#### Fund
```go
type Fund struct {
    ID       uint
    Name     string
    IsExcept bool  // default: false
}
```

#### Asset
```go
type Asset struct {
    ID        uint
    Name      string
    Category  Category  // Enum: 1=현금, 2=달러, 3=금, 4=단기채권, 5=국내ETF, 6=국내주식, 7=국내코인, 8=해외주식, 9=해외ETF, 10=레버리지, 11=해외코인, 12=국내안전자산ETF
    Code      string
    Currency  string    // "WON" or "USD"
    Top       float64
    Bottom    float64
    SellPrice float64
    BuyPrice  float64
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt time.Time
}
```

#### Invest
```go
type Invest struct {
    ID        uint
    FundID    uint
    Fund      Fund
    AssetID   uint
    Asset     Asset
    Price     float64
    Count     float64
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt time.Time
}
```

#### InvestSummary
```go
type InvestSummary struct {
    ID      uint
    FundID  uint
    Fund    Fund
    AssetID uint
    Asset   Asset
    Count   float64
    Sum     float64
}
```

#### Market
```go
type Market struct {
    ID        uint
    Status    uint      // MarketLevel: 1=MAJOR_BEAR, 2=BEAR, 3=VOLATILIY, 4=BULL, 5=MAJOR_BULL
    CreatedAt time.Time
}
```

#### DailyIndex
```go
type DailyIndex struct {
    CreatedAt      time.Time
    FearGreedIndex uint
    NasDaq         float64
    Sp500          float64
}
```

#### CliIndex
```go
type CliIndex struct {
    CreatedAt time.Time
    Index     float64
}
```

#### HighYieldSpread
```go
type HighYieldSpread struct {
    CreatedAt time.Time
    Spread    float64
}
```

#### EmaHist
```go
type EmaHist struct {
    ID      uint
    AssetID uint
    Asset   Asset
    Date    time.Time
    Ema     float64
    NDays   uint
}
```

### Category Enum
```
1  = 현금 (Won)
2  = 달러 (Dollar)
3  = 금 (Gold)
4  = 단기채권 (ShortTermBond)
5  = 국내ETF (DomesticETF)
6  = 국내주식 (DomesticStock)
7  = 국내코인 (DomesticCoin)
8  = 해외주식 (ForeignStock)
9  = 해외ETF (ForeignETF)
10 = 레버리지 (Leverage)
11 = 해외코인 (ForeignCoin)
12 = 국내안전자산ETF (DomesticStableETF)
```

Stable categories: 현금, 달러, 금, 단기채권, 국내안전자산ETF

### Currency Enum
```
WON (KRW)
USD
```

### Market Level Enum
```
1 = MAJOR_BEAR (Max volatile: 15%, Min volatile: 10%)
2 = BEAR (Max volatile: 20%, Min volatile: 15%)
3 = VOLATILIY (Max volatile: 25%, Min volatile: 20%)
4 = BULL (Max volatile: 30%, Min volatile: 25%)
5 = MAJOR_BULL (Max volatile: 40%, Min volatile: 30%)
```

## Authentication

All endpoints (except `/login`) require authentication via JWT token.

### Login
**Endpoint:** `POST /login/`

**Description:** Authenticate user and receive JWT token

**Request Type:** `LoginRequest`
```go
type LoginRequest struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    Password string `json:"password"`
}
```

**Request Body Example:**
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "secure_password"
}
```

**Response Type:** `JWTResponse`
```go
type JWTResponse struct {
    Token  string `json:"token"`
    Expiry int64  `json:"expiry"`  // Unix timestamp
}
```

**Response Example:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiry": 1709856000
}
```

**Status Codes:**
- `200 OK` - Successfully authenticated
- `401 Unauthorized` - Invalid credentials

---

## Market Endpoints

### Get Market Status
**Endpoint:** `GET /market/:date?`

**Description:** Retrieve market status for a specific date (optional)

**Path Parameters:**
- `date` (optional): Date in YYYYMMDD format

**Request:** None

**Response Type:** `Market`
```go
type Market struct {
    ID        uint
    Status    uint      // MarketLevel: 1=MAJOR_BEAR, 2=BEAR, 3=VOLATILIY, 4=BULL, 5=MAJOR_BULL
    CreatedAt time.Time
}
```

**Response Example:**
```json
{
  "ID": 1,
  "Status": 3,
  "CreatedAt": "2024-03-08T12:00:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid date format

---

### Get Market Indicators
**Endpoint:** `GET /market/indicators/:date?`

**Description:** Retrieve daily and CLI market indicators for a specific date

**Path Parameters:**
- `date` (optional): Date in YYYYMMDD format

**Request:** None

**Response Type:** Array containing `[DailyIndex, CliIndex]`
```go
type DailyIndex struct {
    CreatedAt      time.Time
    FearGreedIndex uint
    NasDaq         float64
    Sp500          float64
}

type CliIndex struct {
    CreatedAt time.Time
    Index     float64
}
```

**Response Example:**
```json
[
  {
    "CreatedAt": "2024-03-08T00:00:00Z",
    "FearGreedIndex": 65,
    "NasDaq": 15234.50,
    "Sp500": 4567.89
  },
  {
    "CreatedAt": "2024-03-08T00:00:00Z",
    "Index": 102.5
  }
]
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid date format

---

### Get Weekly Market Indicators
**Endpoint:** `GET /market/weekly_indicators`

**Description:** Retrieve weekly market indicators including Fear & Greed Index, NASDAQ, S&P 500, and High Yield Spread

**Request:** None

**Response Type:** Map of `MarketIndexInner`
```go
type MarketIndexInner struct {
    Value     string    `json:"value"`
    Status    string    `json:"status"`
    GraphData []float64 `json:"graph"`
}
```

**Response Example:**
```json
{
  "Fear & Greed Index": {
    "value": "65",
    "status": "GREED",
    "graph": [45.0, 50.0, 55.0, 60.0, 62.0, 63.0, 65.0]
  },
  "NASDAQ": {
    "value": "15234.50",
    "status": "2.50%",
    "graph": [15000.0, 15050.0, 15100.0, 15150.0, 15180.0, 15200.0, 15234.50]
  },
  "S&P 500": {
    "value": "4567.89",
    "status": "1.20%",
    "graph": [4500.0, 4520.0, 4530.0, 4540.0, 4550.0, 4560.0, 4567.89]
  },
  "High Yield Spread": {
    "value": "3.45",
    "status": "0.50%",
    "graph": [3.40, 3.41, 3.42, 3.43, 3.44, 3.44, 3.45]
  }
}
```

**Notes:**
- Fear & Greed Index status: "GREED" if value > 50, otherwise "FEAR"
- NASDAQ and S&P 500 status: Percentage change from previous day
- High Yield Spread status: Percentage change from previous day
- GraphData contains 7 values (one week of data)

**Status Codes:**
- `200 OK` - Success

---

### Change Market Status
**Endpoint:** `POST /market/`

**Description:** Update market status

**Request Type:** `SaveMarketStatusParam`
```go
type SaveMarketStatusParam struct {
    Status uint `json:"status" validate:"required,market_status"`
}
```

**Request Body Example:**
```json
{
  "status": 3
}
```

**Response Type:** Plain text string
```
시장 상태 저장 성공
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid parameters (status must be 1-5)

---

## Asset Endpoints

### Get All Assets
**Endpoint:** `GET /assets`

**Description:** Retrieve all assets with current prices

**Request:** None

**Response Type:** Array of `assetResponse`
```go
type assetResponse struct {
    ID        uint    `json:"id"`
    Name      string  `json:"name"`
    Category  string  `json:"category"`      // Category name in Korean
    Code      string  `json:"code"`
    Currency  string  `json:"currency"`      // "WON" or "USD"
    Top       float64 `json:"top"`
    Bottom    float64 `json:"bottom"`
    SellPrice float64 `json:"sell"`
    BuyPrice  float64 `json:"buy"`
    Ema       float64 `json:"ema"`
    NDays     float64 `json:"ndays"`
    Price     float64 `json:"price"`         // Current/present price
}
```

**Response Example:**
```json
[
  {
    "id": 1,
    "name": "Apple Inc.",
    "category": "해외주식",
    "code": "AAPL",
    "currency": "USD",
    "top": 150.0,
    "bottom": 100.0,
    "sell": 145.0,
    "buy": 105.0,
    "ema": 0,
    "ndays": 0,
    "price": 125.0
  }
]
```

**Status Codes:**
- `200 OK` - Success

---

### Get Asset List
**Endpoint:** `GET /assets/list`

**Description:** Retrieve simple list of asset IDs and names

**Request:** None

**Response Type:** Array of `assetListResponse`
```go
type assetListResponse struct {
    AssetId   uint   `json:"asset_id"`
    AssetName string `json:"name"`
}
```

**Response Example:**
```json
[
  {
    "asset_id": 1,
    "name": "Apple Inc."
  },
  {
    "asset_id": 2,
    "name": "Samsung Electronics"
  }
]
```

**Status Codes:**
- `200 OK` - Success

---

### Get Asset by ID
**Endpoint:** `GET /assets/:id`

**Description:** Retrieve detailed asset information by ID including latest EMA

**Path Parameters:**
- `id`: Asset ID (integer)

**Request:** None

**Response Type:** `assetResponse`
```go
type assetResponse struct {
    ID        uint    `json:"id"`
    Name      string  `json:"name"`
    Category  string  `json:"category"`
    Code      string  `json:"code"`
    Currency  string  `json:"currency"`
    Top       float64 `json:"top"`
    Bottom    float64 `json:"bottom"`
    SellPrice float64 `json:"sell"`
    BuyPrice  float64 `json:"buy"`
    Ema       float64 `json:"ema"`           // Latest EMA value
    NDays     float64 `json:"ndays"`         // EMA period
}
```

**Response Example:**
```json
{
  "id": 1,
  "name": "Apple Inc.",
  "category": "해외주식",
  "code": "AAPL",
  "currency": "USD",
  "top": 150.0,
  "bottom": 100.0,
  "sell": 145.0,
  "buy": 105.0,
  "ema": 120.5,
  "ndays": 20.0
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Asset not found

---

### Get Asset History
**Endpoint:** `GET /assets/:id/hist`

**Description:** Retrieve investment history for an asset

**Path Parameters:**
- `id`: Asset ID (integer)

**Request:** None

**Response Type:** Array of `HistResponse`
```go
type HistResponse struct {
    FundId    uint    `json:"fund_id"`
    AssetId   uint    `json:"asset_id"`
    AssetName string  `json:"asset_name"`
    Count     float64 `json:"count"`
    Price     float64 `json:"price"`
    CreatedAt string  `json:"created_at"`   // Format: "20060102"
}
```

**Response Example:**
```json
[
  {
    "fund_id": 1,
    "asset_id": 1,
    "asset_name": "Apple Inc.",
    "count": 10.0,
    "price": 120.0,
    "created_at": "20240308"
  }
]
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Asset not found

---

### Add Asset
**Endpoint:** `POST /assets/`

**Description:** Create a new asset

**Request Type:** `AddAssetReq`
```go
type AddAssetReq struct {
    Name      string  `json:"name" validate:"required"`
    Category  string  `json:"category" validate:"required,category"`  // Korean category name
    Code      string  `json:"code"`
    Currency  string  `json:"currency" validate:"required"`           // "WON" or "USD"
    Top       float64 `json:"top"`
    Bottom    float64 `json:"bottom"`
    SellPrice float64 `json:"sell"`
    BuyPrice  float64 `json:"buy"`
    Ema       float64 `json:"ema"`
    Ndays     uint    `json:"ndays"`
}
```

**Request Body Example:**
```json
{
  "name": "Apple Inc.",
  "category": "해외주식",
  "code": "AAPL",
  "currency": "USD",
  "top": 150.0,
  "bottom": 100.0,
  "sell": 145.0,
  "buy": 105.0,
  "ema": 120.5,
  "ndays": 20
}
```

**Notes:**
- Fields `top`, `bottom`, `ema`, and `ndays` are optional
- If `top`/`bottom` are 0, they will be auto-calculated
- If `buy` is 0, it will be set to `bottom` value
- If `ema` is 0, average price will be calculated automatically

**Response Type:** Plain text string
```
자산 정보 저장 성공
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid parameters or validation failure

---

### Update Asset
**Endpoint:** `PUT /assets/`

**Description:** Update existing asset information

**Request Type:** `UpdateAssetReq`
```go
type UpdateAssetReq struct {
    ID        uint    `json:"id" validate:"required"`
    Name      string  `json:"name"`
    Category  string  `json:"category"`
    Code      string  `json:"code"`
    Currency  string  `json:"currency"`
    Top       float64 `json:"top"`
    Bottom    float64 `json:"bottom"`
    SellPrice float64 `json:"sell"`
    BuyPrice  float64 `json:"buy"`
}
```

**Request Body Example:**
```json
{
  "id": 1,
  "name": "Apple Inc.",
  "category": "해외주식",
  "code": "AAPL",
  "currency": "USD",
  "top": 160.0,
  "bottom": 110.0,
  "sell": 155.0,
  "buy": 115.0
}
```

**Response Type:** Plain text string
```
자산 정보 갱신 성공
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid parameters or validation failure
- `404 Not Found` - Asset not found

---

### Delete Asset
**Endpoint:** `DELETE /assets/`

**Description:** Delete an asset

**Request Type:** `DeleteAssetReq`
```go
type DeleteAssetReq struct {
    ID uint `json:"id" validate:"required"`
}
```

**Request Body Example:**
```json
{
  "id": 1
}
```

**Response Type:** Plain text string
```
자산 정보 삭제 성공
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid parameters or validation failure
- `404 Not Found` - Asset not found

---

## Fund Endpoints

### Get Total Fund Status
**Endpoint:** `GET /funds/`

**Description:** Retrieve total status of all funds with amounts (converted to KRW)

**Request:** None

**Response Type:** Map of `TotalStatusResp` (keyed by fund ID)
```go
type TotalStatusResp struct {
    ID     uint    `json:"id"`
    Name   string  `json:"name"`
    Amount float64 `json:"amount"`  // Total amount in KRW
}
```

**Response Example:**
```json
{
  "1": {
    "id": 1,
    "name": "Main Fund",
    "amount": 50000000.0
  },
  "2": {
    "id": 2,
    "name": "Reserve Fund",
    "amount": 25000000.0
  }
}
```

**Notes:**
- USD amounts are converted to KRW using current exchange rate
- Only includes assets with non-zero counts

**Status Codes:**
- `200 OK` - Success

---

### Add Fund
**Endpoint:** `POST /funds/`

**Description:** Create a new fund

**Request Type:** `AddFundReq`
```go
type AddFundReq struct {
    Name string `json:"name" validate:"required"`
}
```

**Request Body Example:**
```json
{
  "name": "New Investment Fund"
}
```

**Response Type:** Plain text string
```
자금 정보 저장 성공
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid parameters or validation failure

---

### Get Fund Assets
**Endpoint:** `GET /funds/:id/assets`

**Description:** Retrieve all assets held by a specific fund with current values

**Path Parameters:**
- `id`: Fund ID (integer)

**Request:** None

**Response Type:** Array of `fundAssetsResponse`
```go
type fundAssetsResponse struct {
    Name         string `json:"name"`
    Amount       string `json:"amount"`         // Total value in KRW
    AmountDollar string `json:"amount_dollar"`  // Total value in USD (if applicable)
    ProfitRate   string `json:"profit_rate"`    // Percentage profit/loss
    Division     string `json:"division"`       // Category name in Korean
    Quantity     string `json:"quantity"`       // Asset count
    Price        string `json:"price"`          // Unit price (not used)
    PriceDollar  string `json:"price_dollar"`   // Unit price in USD (not used)
    IsStable     bool   `json:"isStable"`       // Whether asset is stable category
}
```

**Response Example:**
```json
[
  {
    "name": "Apple Inc.",
    "amount": "15000000.00",
    "amount_dollar": "12000.00",
    "profit_rate": "15.50",
    "division": "해외주식",
    "quantity": "100.00",
    "price": "",
    "price_dollar": "",
    "isStable": false
  },
  {
    "name": "Samsung Electronics",
    "amount": "8500000.00",
    "amount_dollar": "",
    "profit_rate": "8.20",
    "division": "국내주식",
    "quantity": "150.00",
    "price": "",
    "price_dollar": "",
    "isStable": false
  }
]
```

**Notes:**
- Assets with zero count are excluded
- Profit rate calculation: `100 * (current_value + sold_value - buy_value) / buy_value`
- Empty string for profit_rate indicates cash assets or calculation error

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Fund not found

---

### Get Fund History
**Endpoint:** `GET /funds/:id/hist`

**Description:** Retrieve investment history for a specific fund

**Path Parameters:**
- `id`: Fund ID (integer)

**Query Parameters:**
- `start` (optional): Start date in YYYY-MM-DD format
- `end` (optional): End date in YYYY-MM-DD format

**Request:** None (query parameters optional)

**Response Type:** Array of `HistResponse`
```go
type HistResponse struct {
    FundId    uint    `json:"fund_id"`
    AssetId   uint    `json:"asset_id"`
    AssetName string  `json:"asset_name"`
    Count     float64 `json:"count"`          // Positive for buy, negative for sell
    Price     float64 `json:"price"`
    CreatedAt string  `json:"created_at"`     // Format: "2006-01-02 15:04:05"
}
```

**Response Example:**
```json
[
  {
    "fund_id": 1,
    "asset_id": 1,
    "asset_name": "Apple Inc.",
    "count": 10.0,
    "price": 120.0,
    "created_at": "2024-03-08 09:30:00"
  },
  {
    "fund_id": 1,
    "asset_id": 2,
    "asset_name": "Samsung Electronics",
    "count": -5.0,
    "price": 65000.0,
    "created_at": "2024-03-07 14:20:00"
  }
]
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Fund not found

---

### Get Fund Portion
**Endpoint:** `GET /funds/:id/portion`

**Description:** Retrieve stable vs volatile asset portion for a fund (as percentage)

**Path Parameters:**
- `id`: Fund ID (integer)

**Request:** None

**Response Type:** `fundPortionResponse`
```go
type fundPortionResponse struct {
    Stable   int `json:"stable"`    // Percentage (0-100)
    Volatile int `json:"volatile"`  // Percentage (0-100)
}
```

**Response Example:**
```json
{
  "stable": 60,
  "volatile": 40
}
```

**Notes:**
- Values are percentages that sum to 100
- Stable categories: 현금, 달러, 금, 단기채권, 국내안전자산ETF
- USD amounts converted to KRW for calculation
- Returns `{"stable": 0, "volatile": 0}` if fund has no assets

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Fund not found

---

### Get Available Amounts
**Endpoint:** `GET /funds/:id/available_amounts`

**Description:** Retrieve available investment amounts for a fund based on current market status

**Path Parameters:**
- `id`: Fund ID (integer)

**Request:** None

**Response Type:** Returned from `InvestStatusIndicator.InvestAvailableAmount()`
```
Structure depends on implementation of InvestAvailableAmount
```

**Response Example:**
```json
{

}
```

**Notes:**
- Implementation-specific response structure
- Calculated based on market level and fund portion rules

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Fund not found

---

## Investment Endpoints

### Record Investment
**Endpoint:** `POST /invest/`

**Description:** Record a new investment transaction (buy or sell)

**Request Type:** `SaveInvestParam`
```go
type SaveInvestParam struct {
    FundId    uint    `json:"fund_id" validate:"required"`
    AssetId   uint    `json:"asset_id"`
    AssetName string  `json:"name"`
    AssetCode string  `json:"code"`
    Price     float64 `json:"price" validate:"required"`
    Count     float64 `json:"count" validate:"required"`  // Positive for buy, negative for sell
}
```

**Request Body Example:**
```json
{
  "fund_id": 1,
  "asset_id": 5,
  "name": "",
  "code": "",
  "price": 120.5,
  "count": 10.0
}
```

**Notes:**
- You can specify asset by `asset_id`, `name` (AssetName), or `code` (AssetCode)
- At least one asset identifier must be provided
- If multiple identifiers provided, priority: `asset_id` > `name` > `code`
- Positive `count` for purchases, negative `count` for sales

**Response Type:** Plain text string
```
Invest 이력 저장 성공
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid parameters, validation failure, or missing asset information

---

## Event Endpoints

### Get All Events
**Endpoint:** `GET /events/`

**Description:** Retrieve all available events with their status

**Request:** None

**Response Type:** Array of `EventResponse`
```go
type EventResponse struct {
    Id          uint   `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Active      bool   `json:"active"`
}
```

**Response Example:**
```json
[
  {
    "id": 1,
    "title": "Daily Market Update",
    "description": "Updates market indicators daily",
    "active": true
  },
  {
    "id": 2,
    "title": "Portfolio Rebalancing",
    "description": "Rebalances portfolio based on market conditions",
    "active": false
  }
]
```

**Status Codes:**
- `200 OK` - Success

---

### Switch Event Status
**Endpoint:** `POST /events/switch`

**Description:** Enable or disable an event

**Request Type:** `EventStatusChangeRequest`
```go
type EventStatusChangeRequest struct {
    Id     uint `json:"id"`
    Active bool `json:"active"`
}
```

**Request Body Example:**
```json
{
  "id": 1,
  "active": true
}
```

**Response Type:** Plain text string
```
Event 실행 성공
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid parameters or validation failure

---

### Launch Event
**Endpoint:** `POST /events/launch`

**Description:** Manually trigger/launch an event execution

**Request Type:** `EventLaunchRequest`
```go
type EventLaunchRequest struct {
    Id uint `json:"id"`
}
```

**Request Body Example:**
```json
{
  "id": 1
}
```

**Response Type:** Plain text string
```
event 실행 성공
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid parameters or validation failure

---

## Model Endpoints

### Get Categories
**Endpoint:** `GET /categories`

**Description:** Retrieve all available asset categories

**Request:** None

**Response Type:** Array of strings (Korean category names)
```
[]string
```

**Response Example:**
```json
[
  "현금",
  "달러",
  "금",
  "단기채권",
  "국내ETF",
  "국내주식",
  "국내코인",
  "해외주식",
  "해외ETF",
  "레버리지",
  "해외코인",
  "국내안전자산ETF"
]
```

**Notes:**
- These are the valid values for the `category` field when creating/updating assets
- Stable categories: 현금, 달러, 금, 단기채권, 국내안전자산ETF

**Status Codes:**
- `200 OK` - Success

---

### Get Currencies
**Endpoint:** `GET /currencies`

**Description:** Retrieve all available currencies

**Request:** None

**Response Type:** Array of strings
```
[]string
```

**Response Example:**
```json
[
  "WON",
  "USD"
]
```

**Notes:**
- These are the valid values for the `currency` field when creating/updating assets

**Status Codes:**
- `200 OK` - Success

---

## Error Handling

All endpoints return appropriate HTTP status codes:
- `200 OK` - Successful request
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Authenticated but not authorized (non-admin trying non-GET methods)
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

Error responses are typically in plain text format describing the error.

## Authentication & Authorization

- All endpoints require a Bearer token in the Authorization header: `Authorization: Bearer <token>`
- A pass key can also be used for bot/HTTP testing: `Authorization: <passkey>`
- Non-admin users can only perform GET requests
- Admin users can perform all HTTP methods
