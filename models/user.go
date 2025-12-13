package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OTPVerification struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	OTP       string    `json:"otp"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Verified  bool      `json:"verified"`
}

type PersonalDetails struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BusinessDetails struct {
	ID                 uuid.UUID `json:"id"`
	UserID             uuid.UUID `json:"user_id"`
	BusinessName       string    `json:"business_name"`
	TradeLicenseNumber string    `json:"trade_license_number"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type TradeLicense struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Filename  string    `json:"filename"`
	FileURL   string    `json:"file_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AccountStatus struct {
	UserID             uuid.UUID `json:"user_id"`
	Email              string    `json:"email"`
	Status             string    `json:"status"` // "new" or "old"
	HasPersonalDetails bool      `json:"has_personal_details"`
	HasBusinessDetails bool      `json:"has_business_details"`
	HasTradeLicense    bool      `json:"has_trade_license"`
	IsComplete         bool      `json:"is_complete"`
}

type RegistrationSummary struct {
	PersonalInfo PersonalDetails `json:"personal_info"`
	BusinessInfo BusinessDetails `json:"business_info"`
	TradeLicense TradeLicense    `json:"trade_license"`
}

type FinancingRequest struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	Amount          float64   `json:"amount"`
	Purpose         string    `json:"purpose"`
	RepaymentPeriod int       `json:"repayment_period"` // in months
	Status          string    `json:"status"`           // "pending", "approved", "rejected", "disbursed"
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Database methods
func (u *User) Create(db *sql.DB) error {
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	query := `INSERT INTO users (id, email, created_at, updated_at) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, u.ID, u.Email, u.CreatedAt, u.UpdatedAt)
	return err
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	user := &User{}
	query := `SELECT id, email, created_at, updated_at FROM users WHERE email = $1`
	err := db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func GetUserByID(db *sql.DB, id uuid.UUID) (*User, error) {
	user := &User{}
	query := `SELECT id, email, created_at, updated_at FROM users WHERE id = $1`
	err := db.QueryRow(query, id).Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (otp *OTPVerification) Create(db *sql.DB) error {
	otp.ID = uuid.New()
	otp.CreatedAt = time.Now()
	otp.ExpiresAt = time.Now().Add(10 * time.Minute) // OTP expires in 10 minutes
	otp.Verified = false

	query := `INSERT INTO otp_verifications (id, email, otp, expires_at, created_at, verified) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(query, otp.ID, otp.Email, otp.OTP, otp.ExpiresAt, otp.CreatedAt, otp.Verified)
	return err
}

func VerifyOTP(db *sql.DB, email, otp string) (*OTPVerification, error) {
	otpVerification := &OTPVerification{}
	query := `SELECT id, email, otp, expires_at, created_at, verified 
	          FROM otp_verifications 
	          WHERE email = $1 AND otp = $2 AND verified = false 
	          ORDER BY created_at DESC LIMIT 1`

	err := db.QueryRow(query, email, otp).Scan(
		&otpVerification.ID, &otpVerification.Email, &otpVerification.OTP,
		&otpVerification.ExpiresAt, &otpVerification.CreatedAt, &otpVerification.Verified,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Check if OTP is expired
	if time.Now().After(otpVerification.ExpiresAt) {
		return nil, nil
	}

	// Mark as verified
	updateQuery := `UPDATE otp_verifications SET verified = true WHERE id = $1`
	_, err = db.Exec(updateQuery, otpVerification.ID)
	if err != nil {
		return nil, err
	}

	return otpVerification, nil
}

func (pd *PersonalDetails) CreateOrUpdate(db *sql.DB) error {
	var existingID uuid.UUID
	checkQuery := `SELECT id FROM personal_details WHERE user_id = $1`
	err := db.QueryRow(checkQuery, pd.UserID).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Create new
		pd.ID = uuid.New()
		pd.CreatedAt = time.Now()
		pd.UpdatedAt = time.Now()
		query := `INSERT INTO personal_details (id, user_id, full_name, email, phone_number, created_at, updated_at) 
		          VALUES ($1, $2, $3, $4, $5, $6, $7)`
		_, err = db.Exec(query, pd.ID, pd.UserID, pd.FullName, pd.Email, pd.PhoneNumber, pd.CreatedAt, pd.UpdatedAt)
	} else if err == nil {
		// Update existing
		pd.ID = existingID
		pd.UpdatedAt = time.Now()
		query := `UPDATE personal_details SET full_name = $1, email = $2, phone_number = $3, updated_at = $4 
		          WHERE user_id = $5`
		_, err = db.Exec(query, pd.FullName, pd.Email, pd.PhoneNumber, pd.UpdatedAt, pd.UserID)
	}

	return err
}

func GetPersonalDetails(db *sql.DB, userID uuid.UUID) (*PersonalDetails, error) {
	pd := &PersonalDetails{}
	query := `SELECT id, user_id, full_name, email, phone_number, created_at, updated_at 
	          FROM personal_details WHERE user_id = $1`
	err := db.QueryRow(query, userID).Scan(
		&pd.ID, &pd.UserID, &pd.FullName, &pd.Email, &pd.PhoneNumber, &pd.CreatedAt, &pd.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return pd, err
}

func (bd *BusinessDetails) CreateOrUpdate(db *sql.DB) error {
	var existingID uuid.UUID
	checkQuery := `SELECT id FROM business_details WHERE user_id = $1`
	err := db.QueryRow(checkQuery, bd.UserID).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Create new
		bd.ID = uuid.New()
		bd.CreatedAt = time.Now()
		bd.UpdatedAt = time.Now()
		query := `INSERT INTO business_details (id, user_id, business_name, trade_license_number, created_at, updated_at) 
		          VALUES ($1, $2, $3, $4, $5, $6)`
		_, err = db.Exec(query, bd.ID, bd.UserID, bd.BusinessName, bd.TradeLicenseNumber, bd.CreatedAt, bd.UpdatedAt)
	} else if err == nil {
		// Update existing
		bd.ID = existingID
		bd.UpdatedAt = time.Now()
		query := `UPDATE business_details SET business_name = $1, trade_license_number = $2, updated_at = $3 
		          WHERE user_id = $4`
		_, err = db.Exec(query, bd.BusinessName, bd.TradeLicenseNumber, bd.UpdatedAt, bd.UserID)
	}

	return err
}

func GetBusinessDetails(db *sql.DB, userID uuid.UUID) (*BusinessDetails, error) {
	bd := &BusinessDetails{}
	query := `SELECT id, user_id, business_name, trade_license_number, created_at, updated_at 
	          FROM business_details WHERE user_id = $1`
	err := db.QueryRow(query, userID).Scan(
		&bd.ID, &bd.UserID, &bd.BusinessName, &bd.TradeLicenseNumber, &bd.CreatedAt, &bd.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return bd, err
}

func (tl *TradeLicense) CreateOrUpdate(db *sql.DB) error {
	var existingID uuid.UUID
	checkQuery := `SELECT id FROM trade_licenses WHERE user_id = $1`
	err := db.QueryRow(checkQuery, tl.UserID).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Create new
		tl.ID = uuid.New()
		tl.CreatedAt = time.Now()
		tl.UpdatedAt = time.Now()
		query := `INSERT INTO trade_licenses (id, user_id, filename, file_url, created_at, updated_at) 
		          VALUES ($1, $2, $3, $4, $5, $6)`
		_, err = db.Exec(query, tl.ID, tl.UserID, tl.Filename, tl.FileURL, tl.CreatedAt, tl.UpdatedAt)
	} else if err == nil {
		// Update existing
		tl.ID = existingID
		tl.UpdatedAt = time.Now()
		query := `UPDATE trade_licenses SET filename = $1, file_url = $2, updated_at = $3 
		          WHERE user_id = $4`
		_, err = db.Exec(query, tl.Filename, tl.FileURL, tl.UpdatedAt, tl.UserID)
	}

	return err
}

func GetTradeLicense(db *sql.DB, userID uuid.UUID) (*TradeLicense, error) {
	tl := &TradeLicense{}
	query := `SELECT id, user_id, filename, file_url, created_at, updated_at 
	          FROM trade_licenses WHERE user_id = $1`
	err := db.QueryRow(query, userID).Scan(
		&tl.ID, &tl.UserID, &tl.Filename, &tl.FileURL, &tl.CreatedAt, &tl.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return tl, err
}

func GetAccountStatus(db *sql.DB, userID uuid.UUID) (*AccountStatus, error) {
	user, err := GetUserByID(db, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	status := &AccountStatus{
		UserID: userID,
		Email:  user.Email,
		Status: "new",
	}

	// Check personal details
	pd, err := GetPersonalDetails(db, userID)
	if err != nil {
		return nil, err
	}
	status.HasPersonalDetails = pd != nil

	// Check business details
	bd, err := GetBusinessDetails(db, userID)
	if err != nil {
		return nil, err
	}
	status.HasBusinessDetails = bd != nil

	// Check trade license
	tl, err := GetTradeLicense(db, userID)
	if err != nil {
		return nil, err
	}
	status.HasTradeLicense = tl != nil

	// Determine if account is "old" (complete)
	status.IsComplete = status.HasPersonalDetails && status.HasBusinessDetails && status.HasTradeLicense
	if status.IsComplete {
		status.Status = "old"
	}

	return status, nil
}

func GetRegistrationSummary(db *sql.DB, userID uuid.UUID) (*RegistrationSummary, error) {
	pd, err := GetPersonalDetails(db, userID)
	if err != nil {
		return nil, err
	}
	if pd == nil {
		return nil, nil
	}

	bd, err := GetBusinessDetails(db, userID)
	if err != nil {
		return nil, err
	}
	if bd == nil {
		return nil, nil
	}

	tl, err := GetTradeLicense(db, userID)
	if err != nil {
		return nil, err
	}
	if tl == nil {
		return nil, nil
	}

	return &RegistrationSummary{
		PersonalInfo: *pd,
		BusinessInfo: *bd,
		TradeLicense: *tl,
	}, nil
}

func (fr *FinancingRequest) Create(db *sql.DB) error {
	fr.ID = uuid.New()
	fr.CreatedAt = time.Now()
	fr.UpdatedAt = time.Now()
	if fr.Status == "" {
		fr.Status = "pending"
	}

	query := `INSERT INTO financing_requests (id, user_id, amount, purpose, repayment_period, status, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := db.Exec(query, fr.ID, fr.UserID, fr.Amount, fr.Purpose, fr.RepaymentPeriod, fr.Status, fr.CreatedAt, fr.UpdatedAt)
	return err
}

func GetFinancingRequestsByUserID(db *sql.DB, userID uuid.UUID) ([]FinancingRequest, error) {
	query := `SELECT id, user_id, amount, purpose, repayment_period, status, created_at, updated_at 
	          FROM financing_requests WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []FinancingRequest
	for rows.Next() {
		var fr FinancingRequest
		err := rows.Scan(&fr.ID, &fr.UserID, &fr.Amount, &fr.Purpose, &fr.RepaymentPeriod, &fr.Status, &fr.CreatedAt, &fr.UpdatedAt)
		if err != nil {
			return nil, err
		}
		requests = append(requests, fr)
	}

	return requests, rows.Err()
}

func GetFinancingRequestByID(db *sql.DB, id uuid.UUID) (*FinancingRequest, error) {
	fr := &FinancingRequest{}
	query := `SELECT id, user_id, amount, purpose, repayment_period, status, created_at, updated_at 
	          FROM financing_requests WHERE id = $1`
	err := db.QueryRow(query, id).Scan(
		&fr.ID, &fr.UserID, &fr.Amount, &fr.Purpose, &fr.RepaymentPeriod, &fr.Status, &fr.CreatedAt, &fr.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return fr, err
}

func GetLatestFinancingRequestByUserID(db *sql.DB, userID uuid.UUID) (*FinancingRequest, error) {
	fr := &FinancingRequest{}
	query := `SELECT id, user_id, amount, purpose, repayment_period, status, created_at, updated_at 
	          FROM financing_requests WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`
	err := db.QueryRow(query, userID).Scan(
		&fr.ID, &fr.UserID, &fr.Amount, &fr.Purpose, &fr.RepaymentPeriod, &fr.Status, &fr.CreatedAt, &fr.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return fr, err
}
