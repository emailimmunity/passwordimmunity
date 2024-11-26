package payment

import (
    "context"
    "testing"
    "time"

    "github.com/emailimmunity/passwordimmunity/db/models"
    "github.com/google/uuid"
    "github.com/mollie/mollie-api-go/v2/mollie"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type mockMollieClient struct {
    mock.Mock
}

func (m *mockMollieClient) Get(paymentID string) (*mollie.Payment, error) {
    args := m.Called(paymentID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*mollie.Payment), args.Error(1)
}

func (m *mockMollieClient) Create(payment *mollie.Payment) (*mollie.Payment, error) {
    args := m.Called(payment)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*mollie.Payment), args.Error(1)
}

type mockRepository struct {
    mock.Mock
}

func (m *mockRepository) UpdatePaymentStatus(ctx context.Context, paymentID, status string) error {
    args := m.Called(ctx, paymentID, status)
    return args.Error(0)
}

func (m *mockRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {
    args := m.Called(ctx, payment)
    return args.Error(0)
}

type mockLicensingService struct {
    mock.Mock
}

func (m *mockLicensingService) ActivateLicense(ctx context.Context, orgID uuid.UUID, licenseType string, validUntil time.Time) error {
    args := m.Called(ctx, orgID, licenseType, validUntil)
    return args.Error(0)
}

func TestHandleWebhook(t *testing.T) {
    tests := []struct {
        name           string
        paymentID     string
        setupMocks    func(*mockMollieClient, *mockRepository, *mockLicensingService)
        expectedError bool
    }{
        {
            name:       "successful payment and license activation",
            paymentID: "tr_test_123",
            setupMocks: func(mc *mockMollieClient, mr *mockRepository, ml *mockLicensingService) {
                orgID := uuid.New()
                payment := &mollie.Payment{
                    ID:     "tr_test_123",
                    Status: mollie.PaymentStatusPaid,
                    Metadata: map[string]string{
                        "organization_id": orgID.String(),
                        "license_type":    "enterprise",
                        "period":          "yearly",
                    },
                }
                mc.On("Get", "tr_test_123").Return(payment, nil)
                mr.On("UpdatePaymentStatus", mock.Anything, "tr_test_123", "paid").Return(nil)
                ml.On("ActivateLicense", mock.Anything, orgID, "enterprise", mock.MatchedBy(func(t time.Time) bool {
                    // Verify duration is roughly 1 year
                    return t.After(time.Now().Add(364 * 24 * time.Hour))
                })).Return(nil)
            },
            expectedError: false,
        },
        {
            name:       "failed payment",
            paymentID: "tr_test_456",
            setupMocks: func(mc *mockMollieClient, mr *mockRepository, ml *mockLicensingService) {
                orgID := uuid.New()
                payment := &mollie.Payment{
                    ID:     "tr_test_456",
                    Status: mollie.PaymentStatusFailed,
                    Metadata: map[string]string{
                        "organization_id": orgID.String(),
                        "license_type":    "premium",
                        "period":          "monthly",
                    },
                }
                mc.On("Get", "tr_test_456").Return(payment, nil)
                mr.On("UpdatePaymentStatus", mock.Anything, "tr_test_456", "failed").Return(nil)
            },
            expectedError: false,
        },
        {
            name:       "invalid organization ID",
            paymentID: "tr_test_789",
            setupMocks: func(mc *mockMollieClient, mr *mockRepository, ml *mockLicensingService) {
                payment := &mollie.Payment{
                    ID:     "tr_test_789",
                    Status: mollie.PaymentStatusPaid,
                    Metadata: map[string]string{
                        "organization_id": "invalid-uuid",
                        "license_type":    "premium",
                        "period":          "monthly",
                    },
                }
                mc.On("Get", "tr_test_789").Return(payment, nil)
            },
            expectedError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockClient := new(mockMollieClient)
            mockRepo := new(mockRepository)
            mockLicensing := new(mockLicensingService)

            tt.setupMocks(mockClient, mockRepo, mockLicensing)

            service := &MollieService{
                client:     mockClient,
                repository: mockRepo,
                licensing:  mockLicensing,
            }

            err := service.HandleWebhook(context.Background(), tt.paymentID)

            if tt.expectedError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
}

func TestHandleWebhook_Validation(t *testing.T) {
    tests := []struct {
        name          string
        setupMock     func(*mockMollieClient) *mollie.Payment
        expectedError string
    }{
        {
            name: "invalid license type in metadata",
            setupMock: func(mc *mockMollieClient) *mollie.Payment {
                payment := &mollie.Payment{
                    ID: "tr_test_123",
                    Metadata: map[string]string{
                        "organization_id": uuid.New().String(),
                        "license_type":    "invalid",
                        "period":          "monthly",
                    },
                }
                mc.On("Get", "tr_test_123").Return(payment, nil)
                return payment
            },
            expectedError: "invalid license type in payment metadata: invalid",
        },
        {
            name: "invalid period in metadata",
            setupMock: func(mc *mockMollieClient) *mollie.Payment {
                payment := &mollie.Payment{
                    ID: "tr_test_123",
                    Metadata: map[string]string{
                        "organization_id": uuid.New().String(),
                        "license_type":    "premium",
                        "period":          "invalid",
                    },
                }
                mc.On("Get", "tr_test_123").Return(payment, nil)
                return payment
            },
            expectedError: "invalid period in payment metadata: invalid",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockClient := new(mockMollieClient)
            mockRepo := new(mockRepository)
            mockLicensing := new(mockLicensingService)

            payment := tt.setupMock(mockClient)

            service := &MollieService{
                client:     mockClient,
                repository: mockRepo,
                licensing:  mockLicensing,
            }

            err := service.HandleWebhook(context.Background(), payment.ID)
            assert.EqualError(t, err, tt.expectedError)

            mockClient.AssertExpectations(t)
            mockRepo.AssertNotCalled(t, "UpdatePaymentStatus")
            mockLicensing.AssertNotCalled(t, "ActivateLicense")
        })
    }
}

            mockClient.AssertExpectations(t)
            mockRepo.AssertExpectations(t)
        })
    }
}

func TestCreatePayment_Validation(t *testing.T) {
    tests := []struct {
        name          string
        request       PaymentRequest
        expectedError string
    }{
        {
            name: "invalid license type",
            request: PaymentRequest{
                OrganizationID: uuid.New(),
                LicenseType:    "invalid",
                Period:         "monthly",
            },
            expectedError: "invalid license type: invalid",
        },
        {
            name: "invalid period",
            request: PaymentRequest{
                OrganizationID: uuid.New(),
                LicenseType:    "premium",
                Period:         "invalid",
            },
            expectedError: "invalid period: invalid",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockClient := new(mockMollieClient)
            mockRepo := new(mockRepository)

            service := &MollieService{
                client:     mockClient,
                repository: mockRepo,
            }

            payment, err := service.CreatePayment(context.Background(), tt.request)
            assert.Nil(t, payment)
            assert.EqualError(t, err, tt.expectedError)

            mockClient.AssertNotCalled(t, "Create")
            mockRepo.AssertNotCalled(t, "CreatePayment")
        })
    }
}

func TestIsValidLicenseType(t *testing.T) {
    tests := []struct {
        licenseType string
        expected    bool
    }{
        {"enterprise", true},
        {"premium", true},
        {"free", true},
        {"invalid", false},
        {"", false},
    }

    for _, tt := range tests {
        t.Run(tt.licenseType, func(t *testing.T) {
            assert.Equal(t, tt.expected, isValidLicenseType(tt.licenseType))
        })
    }
}

func TestIsValidPeriod(t *testing.T) {
    tests := []struct {
        period   string
        expected bool
    }{
        {"monthly", true},
        {"yearly", true},
        {"invalid", false},
        {"", false},
    }

    for _, tt := range tests {
        t.Run(tt.period, func(t *testing.T) {
            assert.Equal(t, tt.expected, isValidPeriod(tt.period))
        })
    }
}

            mockClient.AssertExpectations(t)
            mockRepo.AssertExpectations(t)
            mockLicensing.AssertExpectations(t)
        })
    }
}

func TestCreatePayment(t *testing.T) {
    tests := []struct {
        name           string
        request        PaymentRequest
        setupMocks     func(*mockMollieClient, *mockRepository)
        expectedError  bool
        checkResponse  func(*testing.T, *models.Payment)
    }{
        {
            name: "successful enterprise yearly payment",
            request: PaymentRequest{
                OrganizationID: uuid.New(),
                LicenseType:    "enterprise",
                Period:         "yearly",
                Description:    "Enterprise Yearly License",
                RedirectURL:    "https://example.com/return",
                WebhookURL:     "https://example.com/webhook",
            },
            setupMocks: func(mc *mockMollieClient, mr *mockRepository) {
                mc.On("Create", mock.MatchedBy(func(p *mollie.Payment) bool {
                    return p.Amount.Value == "999.00" &&
                           p.Amount.Currency == "EUR" &&
                           p.Description == "Enterprise Yearly License"
                })).Return(&mollie.Payment{
                    ID: "tr_test_123",
                }, nil)

                mr.On("CreatePayment", mock.Anything, mock.MatchedBy(func(p *models.Payment) bool {
                    return p.Amount == "999.00" &&
                           p.Currency == "EUR" &&
                           p.Status == "pending" &&
                           p.LicenseType == "enterprise" &&
                           p.Period == "yearly"
                })).Return(nil)
            },
            expectedError: false,
            checkResponse: func(t *testing.T, p *models.Payment) {
                assert.Equal(t, "999.00", p.Amount)
                assert.Equal(t, "EUR", p.Currency)
                assert.Equal(t, "pending", p.Status)
                assert.Equal(t, "enterprise", p.LicenseType)
                assert.Equal(t, "yearly", p.Period)
            },
        },
        {
            name: "mollie creation fails",
            request: PaymentRequest{
                OrganizationID: uuid.New(),
                LicenseType:    "premium",
                Period:         "monthly",
                Description:    "Premium Monthly License",
            },
            setupMocks: func(mc *mockMollieClient, mr *mockRepository) {
                mc.On("Create", mock.Anything).Return(nil, assert.AnError)
            },
            expectedError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockClient := new(mockMollieClient)
            mockRepo := new(mockRepository)

            tt.setupMocks(mockClient, mockRepo)

            service := &MollieService{
                client:     mockClient,
                repository: mockRepo,
            }

            payment, err := service.CreatePayment(context.Background(), tt.request)

            if tt.expectedError {
                assert.Error(t, err)
                assert.Nil(t, payment)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, payment)
                if tt.checkResponse != nil {
                    tt.checkResponse(t, payment)
                }
            }

            mockClient.AssertExpectations(t)
            mockRepo.AssertExpectations(t)
        })
    }
}
