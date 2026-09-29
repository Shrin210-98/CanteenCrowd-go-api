package settings

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ccms.com/api/internal/constants"
	"ccms.com/api/internal/database"
	"ccms.com/api/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ============================================================
// BULK IMPORT — EMPLOYEES + USERS
// ============================================================

// BulkImportRow represents one parsed CSV row (employee + optional user)
type BulkImportRow struct {
	Row int `json:"row"`

	// Employee fields
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Email          string `json:"email"`
	Phone          string `json:"phone,omitempty"`
	HireDate       string `json:"hireDate"` // YYYY-MM-DD
	PositionTitle  string `json:"positionTitle"`
	DepartmentName string `json:"departmentName"`
	Salary         string `json:"salary,omitempty"`

	// Optional user account
	UserType string `json:"userType,omitempty"` // staff | external | ""
	Username string `json:"username,omitempty"`
	RoleName string `json:"roleName,omitempty"`
}

// BulkImportRowResponse — one row + its validation errors
type BulkImportRowResponse struct {
	Row    int               `json:"row"`
	Data   BulkImportRow     `json:"data"`
	Errors map[string]string `json:"errors"` // field → message
	HasErr bool              `json:"hasError"`
}

// BulkImportResponse is the unified response for both dry_run and commit
type BulkImportResponse struct {
	DryRun    bool                    `json:"dryRun"`
	Total     int                     `json:"total"`
	Valid     int                     `json:"valid"`
	Invalid   int                     `json:"invalid"`
	CanCommit bool                    `json:"canCommit"`
	Rows      []BulkImportRowResponse `json:"rows"`

	// Only populated when dryRun=false and commit succeeded
	Created      int                   `json:"created,omitempty"`
	Failed       int                   `json:"failed,omitempty"`
	Summary      *BulkImportSummary    `json:"summary,omitempty"`
	CreatedUsers []BulkImportUserBrief `json:"createdUsers,omitempty"`
	Passwords    []BulkImportPassword  `json:"temporaryPasswords,omitempty"`
}

type BulkImportSummary struct {
	EmployeesCreated int `json:"employeesCreated"`
	UsersCreated     int `json:"usersCreated"`
	EmployeesOnly    int `json:"employeesOnly"`
}

type BulkImportUserBrief struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

// BulkImportPassword — returned only in commit response so admin can share credentials
type BulkImportPassword struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ============================================================
// BULK IMPORT — EMPLOYEES ONLY
// ============================================================

type BulkEmployeeRow struct {
	Row int `json:"row"`

	FirstName             string `json:"firstName"`
	LastName              string `json:"lastName"`
	Email                 string `json:"email"`
	Phone                 string `json:"phone,omitempty"`
	HireDate              string `json:"hireDate"`
	PositionTitle         string `json:"positionTitle"`
	DepartmentName        string `json:"departmentName"`
	Salary                string `json:"salary,omitempty"`
	Address               string `json:"address,omitempty"`
	EmergencyContactName  string `json:"emergencyContactName,omitempty"`
	EmergencyContactPhone string `json:"emergencyContactPhone,omitempty"`
	ProfileDescription    string `json:"profileDescription,omitempty"`
}

type BulkEmployeeRowResponse struct {
	Row    int               `json:"row"`
	Data   BulkEmployeeRow   `json:"data"`
	Errors map[string]string `json:"errors"`
	HasErr bool              `json:"hasError"`
}

type BulkEmployeeResponse struct {
	DryRun    bool                      `json:"dryRun"`
	Total     int                       `json:"total"`
	Valid     int                       `json:"valid"`
	Invalid   int                       `json:"invalid"`
	CanCommit bool                      `json:"canCommit"`
	Rows      []BulkEmployeeRowResponse `json:"rows"`

	// Only populated when dryRun=false and commit succeeded
	Created int `json:"created,omitempty"`
	Failed  int `json:"failed,omitempty"`
}

// ============================================================
// HANDLER 1: EMPLOYEE + USER IMPORT
// ============================================================

// ============================================================
// EMPLOYEES + USERS IMPORT
// ============================================================

// BulkImportUsers handles POST /settings/bulk-import/users
//
// Modes:
//
//	?dry_run=true   → validate only, return rows + errors
//	?dry_run=false  → validate + commit (all-or-nothing)
func (h *Handler) BulkImportUsers(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userType, _ := r.Context().Value(constants.ContextKeyUserType).(string)
	if userType != constants.UserTypeTenantOwner {
		utils.ErrorResponse(w, http.StatusForbidden, "Only tenant owner can bulk import")
		return
	}

	dryRun := r.URL.Query().Get("dry_run") == "true"

	// Parse upload
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "File too large")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing 'file' field")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid CSV: "+err.Error())
		return
	}

	if len(records) < 2 {
		utils.ErrorResponse(w, http.StatusBadRequest, "CSV must have a header + at least one data row")
		return
	}

	if len(records) > 501 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Maximum 500 rows per import")
		return
	}

	// Map header
	headerMap := make(map[string]int)
	for i, col := range records[0] {
		headerMap[strings.TrimSpace(strings.ToLower(col))] = i
	}

	required := []string{"firstname", "lastname", "email", "hiredate", "positiontitle", "departmentname"}
	for _, col := range required {
		if _, ok := headerMap[col]; !ok {
			utils.ErrorResponse(w, http.StatusBadRequest, "Missing required column: "+col)
			return
		}
	}

	// Parse rows
	rows := make([]BulkImportRow, 0, len(records)-1)
	for i, record := range records[1:] {
		rows = append(rows, BulkImportRow{
			Row:            i + 2,
			FirstName:      getCol(record, headerMap, "firstname"),
			LastName:       getCol(record, headerMap, "lastname"),
			Email:          getCol(record, headerMap, "email"),
			Phone:          getCol(record, headerMap, "phone"),
			HireDate:       getCol(record, headerMap, "hiredate"),
			PositionTitle:  getCol(record, headerMap, "positiontitle"),
			DepartmentName: getCol(record, headerMap, "departmentname"),
			Salary:         getCol(record, headerMap, "salary"),
			UserType:       getCol(record, headerMap, "usertype"),
			Username:       getCol(record, headerMap, "username"),
			RoleName:       getCol(record, headerMap, "rolename"),
		})
	}

	// ─────────────── VALIDATION ───────────────
	errorsMap := make([]map[string]string, len(rows)) // one map per row
	for i := range errorsMap {
		errorsMap[i] = make(map[string]string)
	}

	// 1. Basic field validation
	validateBasicFields(rows, errorsMap)

	// 2. Intra-file duplicate detection
	detectIntraFileDuplicates(rows, errorsMap)

	// 3. Fetch reference data
	deptNames := uniqueField(rows, func(r BulkImportRow) string { return r.DepartmentName })
	posTitles := uniqueField(rows, func(r BulkImportRow) string { return r.PositionTitle })
	roleNames := uniqueField(rows, func(r BulkImportRow) string { return r.RoleName })
	emails := uniqueField(rows, func(r BulkImportRow) string { return r.Email })
	usernames := uniqueField(rows, func(r BulkImportRow) string { return r.Username })

	deptMap, err := h.fetchDepartments(r, tenantID, deptNames)
	if err != nil {
		utils.HandleDBError(w, err, "GetDepartmentsByNames")
		return
	}

	posMap, err := h.fetchPositions(r, tenantID, posTitles)
	if err != nil {
		utils.HandleDBError(w, err, "GetPositionsByTitles")
		return
	}

	roleMap, err := h.fetchRoles(r, tenantID, roleNames)
	if err != nil {
		utils.HandleDBError(w, err, "GetRolesByNames")
		return
	}

	existingEmails, _ := h.db.GetExistingEmails(r.Context(), database.GetExistingEmailsParams{
		TenantID: tenantID,
		Emails:   emails,
	})
	emailSet := make(map[string]bool, len(existingEmails))
	for _, e := range existingEmails {
		emailSet[e] = true
	}

	existingUsernames, _ := h.db.GetExistingUsernames(r.Context(), database.GetExistingUsernamesParams{
		TenantID:  tenantID,
		Usernames: usernames,
	})
	usernameSet := make(map[string]bool, len(existingUsernames))
	for _, u := range existingUsernames {
		usernameSet[u] = true
	}

	// 4. Reference validation
	for i := range rows {
		row := &rows[i]
		errs := errorsMap[i]

		if len(errs) > 0 {
			continue // skip if basic validation already failed
		}

		if _, ok := deptMap[row.DepartmentName]; !ok {
			errs["departmentName"] = "Department not found: " + row.DepartmentName
		}
		if _, ok := posMap[row.PositionTitle]; !ok {
			errs["positionTitle"] = "Position not found: " + row.PositionTitle
		}
		if emailSet[row.Email] {
			errs["email"] = "Email already exists"
		}
		if row.UserType != "" && row.Username != "" && usernameSet[row.Username] {
			errs["username"] = "Username already exists"
		}
		if row.UserType != "" && row.RoleName != "" {
			if _, ok := roleMap[row.RoleName]; !ok {
				errs["roleName"] = "Role not found: " + row.RoleName
			}
		}
	}

	// ─────────────── BUILD RESPONSE ───────────────
	respRows := make([]BulkImportRowResponse, len(rows))
	validCount := 0
	for i, row := range rows {
		errs := errorsMap[i]
		hasErr := len(errs) > 0
		if !hasErr {
			validCount++
		}
		respRows[i] = BulkImportRowResponse{
			Row:    row.Row,
			Data:   row,
			Errors: errs,
			HasErr: hasErr,
		}
	}

	resp := BulkImportResponse{
		DryRun:    dryRun,
		Total:     len(rows),
		Valid:     validCount,
		Invalid:   len(rows) - validCount,
		CanCommit: validCount == len(rows),
		Rows:      respRows,
	}

	// ─────────────── DRY RUN — return now ───────────────
	if dryRun {
		msg := fmt.Sprintf("Validation complete: %d valid, %d invalid", resp.Valid, resp.Invalid)
		utils.SuccessResponse(w, http.StatusOK, resp, msg)
		return
	}

	// ─────────────── COMMIT — reject if any errors ───────────────
	if resp.Invalid > 0 {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{
			"data":    resp,
			"message": fmt.Sprintf("Cannot import: %d rows have errors", resp.Invalid),
		})
		return
	}

	// ─────────────── INSERT IN TRANSACTION ───────────────
	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		utils.HandleDBError(w, err, "BeginTransaction")
		return
	}
	defer tx.Rollback(r.Context())

	qtx := database.New(h.Pool).WithTx(tx)

	created := make([]BulkImportUserBrief, 0, len(rows))
	passwords := make([]BulkImportPassword, 0, len(rows))
	employeesCreated := 0
	usersCreated := 0
	employeesOnly := 0

	for _, row := range rows {
		hireDate, _ := time.Parse("2006-01-02", row.HireDate)

		var salary pgtype.Numeric
		if row.Salary != "" {
			_ = salary.Scan(row.Salary)
		}

		emp, err := qtx.CreateEmployee(r.Context(), database.CreateEmployeeParams{
			TenantID:     tenantID,
			UserID:       nil,
			FirstName:    row.FirstName,
			LastName:     row.LastName,
			Email:        row.Email,
			Phone:        strPtrOrNil(row.Phone),
			HireDate:     hireDate,
			PositionID:   posMap[row.PositionTitle],
			DepartmentID: deptMap[row.DepartmentName],
			Salary:       salary,
			IsActive:     true,
		})
		if err != nil {
			utils.HandleDBError(w, err, "CreateEmployee")
			return
		}
		employeesCreated++

		if row.UserType != "" {
			rawPassword := generateInitialPassword(row.FirstName, row.Phone)
			passwordHash, err := utils.HashPassword(rawPassword)
			if err != nil {
				utils.HandleDBError(w, err, "HashPassword")
				return
			}

			var roleID *uuid.UUID
			if row.RoleName != "" {
				if rid, ok := roleMap[row.RoleName]; ok {
					roleID = &rid
				}
			}

			user, err := qtx.CreateUser(r.Context(), database.CreateUserParams{
				TenantID:            tenantID,
				Username:            row.Username,
				Email:               row.Email,
				PasswordHash:        passwordHash,
				UserType:            row.UserType,
				RoleID:              roleID,
				PermissionsOverride: nil,
			})
			if err != nil {
				utils.HandleDBError(w, err, "CreateUser")
				return
			}
			usersCreated++

			fullName := row.FirstName + " " + row.LastName
			_, err = qtx.CreateUserProfile(r.Context(), database.CreateUserProfileParams{
				UserID:   user.ID,
				TenantID: tenantID,
				FullName: &fullName,
			})
			if err != nil {
				utils.HandleDBError(w, err, "CreateUserProfile")
				return
			}

			_, err = qtx.LinkUserToEmployee(r.Context(), database.LinkUserToEmployeeParams{
				ID:       emp.ID,
				UserID:   &user.ID,
				TenantID: tenantID,
			})
			if err != nil {
				utils.HandleDBError(w, err, "LinkUserToEmployee")
				return
			}

			_ = qtx.SetMustChangePassword(r.Context(), database.SetMustChangePasswordParams{
				ID:       user.ID,
				TenantID: tenantID,
			})

			h.queueWelcomeEmail(r.Context(), user, rawPassword)

			created = append(created, BulkImportUserBrief{
				ID:       user.ID,
				Username: user.Username,
				Email:    user.Email,
			})
			passwords = append(passwords, BulkImportPassword{
				Username: user.Username,
				Email:    user.Email,
				Password: rawPassword,
			})
		} else {
			employeesOnly++
			created = append(created, BulkImportUserBrief{
				ID:       emp.ID,
				Username: "",
				Email:    emp.Email,
			})
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		utils.HandleDBError(w, err, "CommitTransaction")
		return
	}

	// Update response with commit details
	resp.Created = len(created)
	resp.Failed = 0
	resp.Summary = &BulkImportSummary{
		EmployeesCreated: employeesCreated,
		UsersCreated:     usersCreated,
		EmployeesOnly:    employeesOnly,
	}
	resp.CreatedUsers = created
	resp.Passwords = passwords

	utils.SuccessResponse(w, http.StatusCreated, resp, "Bulk import completed successfully")
}

// ============================================================
// EMPLOYEES ONLY IMPORT
// ============================================================

// BulkImportEmployees handles POST /settings/bulk-import/employees
func (h *Handler) BulkImportEmployees(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userType, _ := r.Context().Value(constants.ContextKeyUserType).(string)
	if userType != constants.UserTypeTenantOwner {
		utils.ErrorResponse(w, http.StatusForbidden, "Only tenant owner can bulk import")
		return
	}

	dryRun := r.URL.Query().Get("dry_run") == "true"

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "File too large")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing 'file' field")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid CSV: "+err.Error())
		return
	}

	if len(records) < 2 {
		utils.ErrorResponse(w, http.StatusBadRequest, "CSV must have a header + at least one data row")
		return
	}

	if len(records) > 501 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Maximum 500 rows per import")
		return
	}

	headerMap := make(map[string]int)
	for i, col := range records[0] {
		headerMap[strings.TrimSpace(strings.ToLower(col))] = i
	}

	required := []string{"firstname", "lastname", "email", "hiredate", "positiontitle", "departmentname"}
	for _, col := range required {
		if _, ok := headerMap[col]; !ok {
			utils.ErrorResponse(w, http.StatusBadRequest, "Missing required column: "+col)
			return
		}
	}

	// Parse rows
	rows := make([]BulkEmployeeRow, 0, len(records)-1)
	for i, record := range records[1:] {
		rows = append(rows, BulkEmployeeRow{
			Row:                   i + 2,
			FirstName:             getCol(record, headerMap, "firstname"),
			LastName:              getCol(record, headerMap, "lastname"),
			Email:                 getCol(record, headerMap, "email"),
			Phone:                 getCol(record, headerMap, "phone"),
			HireDate:              getCol(record, headerMap, "hiredate"),
			PositionTitle:         getCol(record, headerMap, "positiontitle"),
			DepartmentName:        getCol(record, headerMap, "departmentname"),
			Salary:                getCol(record, headerMap, "salary"),
			Address:               getCol(record, headerMap, "address"),
			EmergencyContactName:  getCol(record, headerMap, "emergencycontactname"),
			EmergencyContactPhone: getCol(record, headerMap, "emergencycontactphone"),
			ProfileDescription:    getCol(record, headerMap, "profiledescription"),
		})
	}

	// ─────────────── VALIDATION ───────────────
	errorsMap := make([]map[string]string, len(rows))
	for i := range errorsMap {
		errorsMap[i] = make(map[string]string)
	}

	// Basic fields
	for i := range rows {
		row := &rows[i]
		errs := errorsMap[i]

		if row.FirstName == "" {
			errs["firstName"] = "First name is required"
		}
		if row.LastName == "" {
			errs["lastName"] = "Last name is required"
		}
		if row.Email == "" {
			errs["email"] = "Email is required"
		} else if !strings.Contains(row.Email, "@") {
			errs["email"] = "Invalid email format"
		}
		if row.HireDate == "" {
			errs["hireDate"] = "Hire date is required"
		} else if _, err := time.Parse("2006-01-02", row.HireDate); err != nil {
			errs["hireDate"] = "Hire date must be YYYY-MM-DD"
		}
		if row.PositionTitle == "" {
			errs["positionTitle"] = "Position is required"
		}
		if row.DepartmentName == "" {
			errs["departmentName"] = "Department is required"
		}
	}

	// Intra-file duplicates
	seenEmails := make(map[string]int)
	for i := range rows {
		if len(errorsMap[i]) > 0 {
			continue
		}
		if firstRow, exists := seenEmails[rows[i].Email]; exists {
			errorsMap[i]["email"] = fmt.Sprintf("Duplicate email in file (also row %d)", firstRow)
			continue
		}
		seenEmails[rows[i].Email] = rows[i].Row
	}

	// Fetch reference data
	deptNames := uniqueEmployeeField(rows, func(r BulkEmployeeRow) string { return r.DepartmentName })
	posTitles := uniqueEmployeeField(rows, func(r BulkEmployeeRow) string { return r.PositionTitle })
	emails := uniqueEmployeeField(rows, func(r BulkEmployeeRow) string { return r.Email })

	departments, _ := h.db.GetDepartmentsByNames(r.Context(), database.GetDepartmentsByNamesParams{
		TenantID: tenantID,
		Names:    deptNames,
	})
	deptMap := make(map[string]uuid.UUID)
	for _, d := range departments {
		deptMap[d.Name] = d.ID
	}

	positions, _ := h.db.GetPositionsByTitles(r.Context(), database.GetPositionsByTitlesParams{
		TenantID: tenantID,
		Titles:   posTitles,
	})
	posMap := make(map[string]uuid.UUID)
	for _, p := range positions {
		posMap[p.Title] = p.ID
	}

	existingEmails, _ := h.db.GetExistingEmails(r.Context(), database.GetExistingEmailsParams{
		TenantID: tenantID,
		Emails:   emails,
	})
	emailSet := make(map[string]bool, len(existingEmails))
	for _, e := range existingEmails {
		emailSet[e] = true
	}

	// Reference validation
	for i := range rows {
		row := &rows[i]
		errs := errorsMap[i]
		if len(errs) > 0 {
			continue
		}
		if _, ok := deptMap[row.DepartmentName]; !ok {
			errs["departmentName"] = "Department not found: " + row.DepartmentName
		}
		if _, ok := posMap[row.PositionTitle]; !ok {
			errs["positionTitle"] = "Position not found: " + row.PositionTitle
		}
		if emailSet[row.Email] {
			errs["email"] = "Email already exists"
		}
	}

	// Build response
	respRows := make([]BulkEmployeeRowResponse, len(rows))
	validCount := 0
	for i, row := range rows {
		errs := errorsMap[i]
		hasErr := len(errs) > 0
		if !hasErr {
			validCount++
		}
		respRows[i] = BulkEmployeeRowResponse{
			Row:    row.Row,
			Data:   row,
			Errors: errs,
			HasErr: hasErr,
		}
	}

	resp := BulkEmployeeResponse{
		DryRun:    dryRun,
		Total:     len(rows),
		Valid:     validCount,
		Invalid:   len(rows) - validCount,
		CanCommit: validCount == len(rows),
		Rows:      respRows,
	}

	// Dry run — return now
	if dryRun {
		msg := fmt.Sprintf("Validation complete: %d valid, %d invalid", resp.Valid, resp.Invalid)
		utils.SuccessResponse(w, http.StatusOK, resp, msg)
		return
	}

	// Commit — reject if any errors
	if resp.Invalid > 0 {
		utils.JsonResponse(w, http.StatusBadRequest, map[string]any{
			"data":    resp,
			"message": fmt.Sprintf("Cannot import: %d rows have errors", resp.Invalid),
		})
		return
	}

	// Insert in transaction
	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		utils.HandleDBError(w, err, "BeginTransaction")
		return
	}
	defer tx.Rollback(r.Context())

	qtx := database.New(h.Pool).WithTx(tx)
	createdCount := 0

	for _, row := range rows {
		hireDate, _ := time.Parse("2006-01-02", row.HireDate)

		var salary pgtype.Numeric
		if row.Salary != "" {
			_ = salary.Scan(row.Salary)
		}

		_, err = qtx.CreateEmployee(r.Context(), database.CreateEmployeeParams{
			TenantID:              tenantID,
			UserID:                nil,
			FirstName:             row.FirstName,
			LastName:              row.LastName,
			Email:                 row.Email,
			Phone:                 strPtrOrNil(row.Phone),
			Address:               strPtrOrNil(row.Address),
			HireDate:              hireDate,
			PositionID:            posMap[row.PositionTitle],
			DepartmentID:          deptMap[row.DepartmentName],
			Salary:                salary,
			EmergencyContactName:  strPtrOrNil(row.EmergencyContactName),
			EmergencyContactPhone: strPtrOrNil(row.EmergencyContactPhone),
			ProfileDescription:    strPtrOrNil(row.ProfileDescription),
			IsActive:              true,
		})
		if err != nil {
			utils.HandleDBError(w, err, "CreateEmployee")
			return
		}
		createdCount++
	}

	if err := tx.Commit(r.Context()); err != nil {
		utils.HandleDBError(w, err, "CommitTransaction")
		return
	}

	resp.Created = createdCount
	resp.Failed = 0

	utils.SuccessResponse(w, http.StatusCreated, resp, "Bulk employee import completed successfully")
}

func (h *Handler) DownloadBulkUsersTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="employees_users_template.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{
		"firstName", "lastName", "email", "phone", "hireDate",
		"positionTitle", "departmentName", "salary",
		"userType", "username", "roleName",
	})
	_ = cw.Write([]string{
		"John", "Doe", "john@example.com", "+1234567890", "2026-01-15",
		"Cook", "Kitchen", "45000",
		"staff", "john.doe", "Staff",
	})
	_ = cw.Write([]string{
		"Jane", "Smith", "jane@example.com", "", "2026-01-20",
		"Cashier", "Service", "38000",
		"", "", "",
	})
	cw.Flush()
}

func (h *Handler) DownloadBulkEmployeesTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="employees_only_template.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{
		"firstName", "lastName", "email", "phone", "hireDate",
		"positionTitle", "departmentName", "salary",
		"address", "emergencyContactName", "emergencyContactPhone", "profileDescription",
	})
	_ = cw.Write([]string{
		"John", "Doe", "john@example.com", "+1234567890", "2026-01-15",
		"Cook", "Kitchen", "45000",
		"123 Main St", "Mary Doe", "+1234567899", "Experienced chef",
	})
	cw.Flush()
}

// validateBasicFields — inline basic field validation for employees+users rows
func validateBasicFields(rows []BulkImportRow, errorsMap []map[string]string) {
	for i := range rows {
		row := &rows[i]
		errs := errorsMap[i]

		if row.FirstName == "" {
			errs["firstName"] = "First name is required"
		} else if len(row.FirstName) > 100 {
			errs["firstName"] = "First name too long (max 100)"
		}

		if row.LastName == "" {
			errs["lastName"] = "Last name is required"
		} else if len(row.LastName) > 100 {
			errs["lastName"] = "Last name too long (max 100)"
		}

		if row.Email == "" {
			errs["email"] = "Email is required"
		} else if !strings.Contains(row.Email, "@") || !strings.Contains(row.Email, ".") {
			errs["email"] = "Invalid email format"
		}

		if row.HireDate == "" {
			errs["hireDate"] = "Hire date is required"
		} else if _, err := time.Parse("2006-01-02", row.HireDate); err != nil {
			errs["hireDate"] = "Hire date must be YYYY-MM-DD"
		}

		if row.PositionTitle == "" {
			errs["positionTitle"] = "Position title is required"
		}

		if row.DepartmentName == "" {
			errs["departmentName"] = "Department name is required"
		}

		if row.Salary != "" {
			if !isNumeric(row.Salary) {
				errs["salary"] = "Salary must be a number"
			}
		}

		if row.UserType != "" {
			allowed := map[string]bool{"staff": true, "external": true}
			if !allowed[row.UserType] {
				errs["userType"] = "User type must be 'staff' or 'external'"
			}
			if row.Username == "" {
				errs["username"] = "Username required when user type is set"
			} else if len(row.Username) < 3 || len(row.Username) > 50 {
				errs["username"] = "Username must be 3-50 characters"
			}
		} else if row.Username != "" {
			errs["userType"] = "User type required when username is set"
		}
	}
}

func detectIntraFileDuplicates(rows []BulkImportRow, errorsMap []map[string]string) {
	seenEmails := make(map[string]int)
	seenUsernames := make(map[string]int)

	for i := range rows {
		row := &rows[i]
		errs := errorsMap[i]
		if len(errs) > 0 {
			continue
		}

		if firstRow, exists := seenEmails[row.Email]; exists {
			errs["email"] = fmt.Sprintf("Duplicate email in file (also row %d)", firstRow)
			continue
		}
		seenEmails[row.Email] = row.Row

		if row.Username != "" {
			if firstRow, exists := seenUsernames[row.Username]; exists {
				errs["username"] = fmt.Sprintf("Duplicate username in file (also row %d)", firstRow)
				continue
			}
			seenUsernames[row.Username] = row.Row
		}
	}
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	dots := 0
	for _, ch := range s {
		if ch == '.' {
			dots++
			if dots > 1 {
				return false
			}
			continue
		}
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func getCol(record []string, headerMap map[string]int, col string) string {
	idx, ok := headerMap[col]
	if !ok || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func uniqueField(rows []BulkImportRow, get func(BulkImportRow) string) []string {
	seen := make(map[string]bool)
	out := []string{}
	for _, row := range rows {
		v := get(row)
		if v != "" && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func uniqueEmployeeField(rows []BulkEmployeeRow, get func(BulkEmployeeRow) string) []string {
	seen := make(map[string]bool)
	out := []string{}
	for _, row := range rows {
		v := get(row)
		if v != "" && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

// generateInitialPassword — deterministic initial password
// Format: <FirstName>@<last4-of-phone>  or  <FirstName>@<current-year>
func generateInitialPassword(firstName, phone string) string {
	name := strings.TrimSpace(firstName)
	if name == "" {
		name = "User"
	}
	// Capitalize first letter, lowercase rest
	name = strings.ToUpper(name[:1]) + strings.ToLower(name[1:])

	digits := ""
	for _, ch := range phone {
		if ch >= '0' && ch <= '9' {
			digits += string(ch)
		}
	}

	if len(digits) >= 4 {
		return fmt.Sprintf("%s@%s", name, digits[len(digits)-4:])
	}
	return fmt.Sprintf("%s@%d", name, time.Now().Year())
}

// ---------- Reference data fetch helpers ----------

func (h *Handler) fetchDepartments(r *http.Request, tenantID uuid.UUID, names []string) (map[string]uuid.UUID, error) {
	if len(names) == 0 {
		return map[string]uuid.UUID{}, nil
	}
	rows, err := h.db.GetDepartmentsByNames(r.Context(), database.GetDepartmentsByNamesParams{
		TenantID: tenantID,
		Names:    names,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[string]uuid.UUID, len(rows))
	for _, d := range rows {
		out[d.Name] = d.ID
	}
	return out, nil
}

func (h *Handler) fetchPositions(r *http.Request, tenantID uuid.UUID, titles []string) (map[string]uuid.UUID, error) {
	if len(titles) == 0 {
		return map[string]uuid.UUID{}, nil
	}
	rows, err := h.db.GetPositionsByTitles(r.Context(), database.GetPositionsByTitlesParams{
		TenantID: tenantID,
		Titles:   titles,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[string]uuid.UUID, len(rows))
	for _, p := range rows {
		out[p.Title] = p.ID
	}
	return out, nil
}

func (h *Handler) fetchRoles(r *http.Request, tenantID uuid.UUID, names []string) (map[string]uuid.UUID, error) {
	if len(names) == 0 {
		return map[string]uuid.UUID{}, nil
	}
	rows, err := h.db.GetRolesByNames(r.Context(), database.GetRolesByNamesParams{
		TenantID: tenantID,
		Names:    names,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[string]uuid.UUID, len(rows))
	for _, ro := range rows {
		out[ro.Name] = ro.ID
	}
	return out, nil
}

// queueWelcomeEmail — placeholder; wire to your mailer/job queue
func (h *Handler) queueWelcomeEmail(_ context.Context, _ database.User, _ string) {
	// TODO: send email with username + temporary password
}
