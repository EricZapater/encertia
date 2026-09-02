package user_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/encertia/backend/internal/shared"
	"github.com/encertia/backend/internal/user"
	"github.com/google/uuid"
)

func setupUserService() (user.Service, *mockRepository) {
	repo := newMockRepository()
	svc := user.NewService(repo)
	return svc, repo
}

// HU-USER-01: ListUsers
func TestListUsers_AdminCanListAll(t *testing.T) {
	svc, _ := setupUserService()
	ctx := context.Background()

	// Seed users
	_, _ = svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "admin@encertia.cat",
		Password:  "Password123!",
		FirstName: "Admin",
		LastName:  "System",
		Role:      user.RoleAdmin,
	})
	_, _ = svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "teacher@encertia.cat",
		Password:  "Password123!",
		FirstName: "Teacher",
		LastName:  "One",
		Role:      user.RoleTeacher,
	})
	_, _ = svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "student@encertia.cat",
		Password:  "Password123!",
		FirstName: "Student",
		LastName:  "One",
		Role:      user.RoleStudent,
	})

	res, appErr := svc.ListUsers(ctx, uuid.New(), "admin", user.ListUsersFilter{
		Page:     1,
		PageSize: 20,
	})
	if appErr != nil {
		t.Fatalf("unexpected error listing users as admin: %v", appErr)
	}

	if res.Pagination.TotalCount != 3 {
		t.Errorf("expected 3 total users, got %d", res.Pagination.TotalCount)
	}
	if len(res.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(res.Items))
	}
}

func TestListUsers_TeacherCanOnlyListStudents(t *testing.T) {
	svc, _ := setupUserService()
	ctx := context.Background()

	_, _ = svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "admin@encertia.cat",
		Password:  "Password123!",
		FirstName: "Admin",
		LastName:  "System",
		Role:      user.RoleAdmin,
	})
	_, _ = svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "student1@encertia.cat",
		Password:  "Password123!",
		FirstName: "Student",
		LastName:  "One",
		Role:      user.RoleStudent,
	})

	res, appErr := svc.ListUsers(ctx, uuid.New(), "teacher", user.ListUsersFilter{
		Page:     1,
		PageSize: 20,
	})
	if appErr != nil {
		t.Fatalf("unexpected error listing users as teacher: %v", appErr)
	}

	if res.Pagination.TotalCount != 1 {
		t.Errorf("expected 1 student user, got %d", res.Pagination.TotalCount)
	}
	if res.Items[0].Role != user.RoleStudent {
		t.Errorf("expected role student, got %s", res.Items[0].Role)
	}
}

func TestListUsers_StudentForbidden(t *testing.T) {
	svc, _ := setupUserService()
	ctx := context.Background()

	_, appErr := svc.ListUsers(ctx, uuid.New(), "student", user.ListUsersFilter{})
	if appErr == nil {
		t.Fatal("expected forbidden error for student, got nil")
	}
	if appErr.Code != shared.ErrCodeForbidden {
		t.Errorf("expected FORBIDDEN code, got %s", appErr.Code)
	}
	if appErr.StatusCode != 403 {
		t.Errorf("expected 403 status, got %d", appErr.StatusCode)
	}
}

func TestListUsers_FiltersAndPagination(t *testing.T) {
	svc, repo := setupUserService()
	ctx := context.Background()

	// Create 15 users
	for i := 1; i <= 15; i++ {
		email := fmt.Sprintf("student%d@encertia.cat", i)
		name := fmt.Sprintf("Student%d", i)
		if i == 5 {
			name = "SpecialMarc"
		}
		_, _ = svc.CreateUser(ctx, "admin", user.CreateUserInput{
			Email:     email,
			Password:  "Password123!",
			FirstName: name,
			LastName:  "Test",
			Role:      user.RoleStudent,
		})
	}

	// Deactivate user 1
	u1, _ := repo.GetUserByEmail(ctx, "student1@encertia.cat")
	u1.IsActive = false

	// Test pagination page 1 (pageSize 5)
	resP1, appErr := svc.ListUsers(ctx, uuid.New(), "admin", user.ListUsersFilter{
		Page:     1,
		PageSize: 5,
		Status:   "active",
	})
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if resP1.Pagination.TotalCount != 14 { // 15 - 1 inactive
		t.Errorf("expected 14 active users, got %d", resP1.Pagination.TotalCount)
	}
	if resP1.Pagination.TotalPages != 3 {
		t.Errorf("expected 3 total pages, got %d", resP1.Pagination.TotalPages)
	}
	if len(resP1.Items) != 5 {
		t.Errorf("expected 5 items on page 1, got %d", len(resP1.Items))
	}

	// Test search filter
	resSearch, _ := svc.ListUsers(ctx, uuid.New(), "admin", user.ListUsersFilter{
		Search: "SpecialMarc",
	})
	if resSearch.Pagination.TotalCount != 1 {
		t.Errorf("expected 1 match for search SpecialMarc, got %d", resSearch.Pagination.TotalCount)
	}

	// Test status = "inactive"
	resInactive, _ := svc.ListUsers(ctx, uuid.New(), "admin", user.ListUsersFilter{
		Status: "inactive",
	})
	if resInactive.Pagination.TotalCount != 1 {
		t.Errorf("expected 1 inactive user, got %d", resInactive.Pagination.TotalCount)
	}
}

// HU-USER-02: CreateUser
func TestCreateUser_Success_Admin(t *testing.T) {
	svc, _ := setupUserService()
	ctx := context.Background()

	// Admin creating teacher
	res, appErr := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "prof.nou@encertia.cat",
		Password:  "TeacherSecret123!",
		FirstName: "Jordi",
		LastName:  "Roca",
		Role:      user.RoleTeacher,
	})
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if res.User.Email != "prof.nou@encertia.cat" {
		t.Errorf("expected email prof.nou@encertia.cat, got %s", res.User.Email)
	}
	if res.User.Role != user.RoleTeacher {
		t.Errorf("expected role teacher, got %s", res.User.Role)
	}
	if !res.User.IsActive {
		t.Error("expected newly created user to be active")
	}
}

func TestCreateUser_TeacherRestrictions(t *testing.T) {
	svc, _ := setupUserService()
	ctx := context.Background()

	// Teacher creating student -> Allowed
	res, appErr := svc.CreateUser(ctx, "teacher", user.CreateUserInput{
		Email:     "alumne.nou@encertia.cat",
		Password:  "StudentSecret123!",
		FirstName: "Carla",
		LastName:  "Serra",
		Role:      user.RoleStudent,
	})
	if appErr != nil {
		t.Fatalf("unexpected error for teacher creating student: %v", appErr)
	}
	if res.User.Role != user.RoleStudent {
		t.Errorf("expected student, got %s", res.User.Role)
	}

	// Teacher creating teacher -> Forbidden
	_, appErr2 := svc.CreateUser(ctx, "teacher", user.CreateUserInput{
		Email:     "altre.prof@encertia.cat",
		Password:  "TeacherSecret123!",
		FirstName: "Pere",
		LastName:  "Mas",
		Role:      user.RoleTeacher,
	})
	if appErr2 == nil {
		t.Fatal("expected forbidden error when teacher creates teacher, got nil")
	}
	if appErr2.StatusCode != 403 {
		t.Errorf("expected status 403, got %d", appErr2.StatusCode)
	}

	// Teacher creating admin -> Forbidden
	_, appErr3 := svc.CreateUser(ctx, "teacher", user.CreateUserInput{
		Email:     "admin.nou@encertia.cat",
		Password:  "AdminSecret123!",
		FirstName: "Super",
		LastName:  "Admin",
		Role:      user.RoleAdmin,
	})
	if appErr3 == nil {
		t.Fatal("expected forbidden error when teacher creates admin, got nil")
	}
	if appErr3.StatusCode != 403 {
		t.Errorf("expected status 403, got %d", appErr3.StatusCode)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	svc, _ := setupUserService()
	ctx := context.Background()

	_, _ = svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "duplicat@encertia.cat",
		Password:  "Password123!",
		FirstName: "Marc",
		LastName:  "Vila",
		Role:      user.RoleStudent,
	})

	_, appErr := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "duplicat@encertia.cat",
		Password:  "Password123!",
		FirstName: "Marc",
		LastName:  "Vila",
		Role:      user.RoleStudent,
	})
	if appErr == nil {
		t.Fatal("expected conflict error, got nil")
	}
	if appErr.StatusCode != 409 {
		t.Errorf("expected status 409, got %d", appErr.StatusCode)
	}
	if appErr.Code != shared.ErrCodeEmailAlreadyExists {
		t.Errorf("expected EMAIL_ALREADY_EXISTS, got %s", appErr.Code)
	}
}

func TestCreateUser_ValidationErrors(t *testing.T) {
	svc, _ := setupUserService()
	ctx := context.Background()

	tests := []struct {
		name  string
		input user.CreateUserInput
	}{
		{"empty email", user.CreateUserInput{Email: "", Password: "Password123!", FirstName: "A", LastName: "B", Role: user.RoleStudent}},
		{"invalid email", user.CreateUserInput{Email: "invalid-email", Password: "Password123!", FirstName: "A", LastName: "B", Role: user.RoleStudent}},
		{"short password", user.CreateUserInput{Email: "valid@encertia.cat", Password: "short", FirstName: "A", LastName: "B", Role: user.RoleStudent}},
		{"empty first name", user.CreateUserInput{Email: "valid@encertia.cat", Password: "Password123!", FirstName: "", LastName: "B", Role: user.RoleStudent}},
		{"empty last name", user.CreateUserInput{Email: "valid@encertia.cat", Password: "Password123!", FirstName: "A", LastName: "", Role: user.RoleStudent}},
		{"invalid role", user.CreateUserInput{Email: "valid@encertia.cat", Password: "Password123!", FirstName: "A", LastName: "B", Role: "invalid_role"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, appErr := svc.CreateUser(ctx, "admin", tt.input)
			if appErr == nil {
				t.Fatalf("expected validation error for %s, got nil", tt.name)
			}
			if appErr.StatusCode != 400 {
				t.Errorf("expected status 400, got %d", appErr.StatusCode)
			}
		})
	}
}

// HU-USER-03: BatchCreateUsers
func TestBatchCreateUsers_MixedBatch(t *testing.T) {
	svc, _ := setupUserService()
	ctx := context.Background()

	// Pre-seed an existing user
	_, _ = svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "existent@encertia.cat",
		Password:  "Password123!",
		FirstName: "Existent",
		LastName:  "User",
		Role:      user.RoleStudent,
	})

	pwdValid := "Password123!"
	pwdShort := "123"

	req := user.BatchCreateUsersRequest{
		Users: []user.BatchUserItem{
			{Email: "valid1@encertia.cat", FirstName: "Valid", LastName: "One", Role: user.RoleStudent, Password: &pwdValid},
			{Email: "existent@encertia.cat", FirstName: "Duplicate", LastName: "DB", Role: user.RoleStudent},
			{Email: "valid2@encertia.cat", FirstName: "Valid", LastName: "Two", Role: user.RoleStudent},
			{Email: "valid2@encertia.cat", FirstName: "Duplicate", LastName: "Batch", Role: user.RoleStudent},
			{Email: "invalid-email", FirstName: "Bad", LastName: "Email", Role: user.RoleStudent},
			{Email: "shortpwd@encertia.cat", FirstName: "Short", LastName: "Pwd", Role: user.RoleStudent, Password: &pwdShort},
			{Email: "teacher@encertia.cat", FirstName: "Teacher", LastName: "Bad", Role: user.RoleTeacher}, // When teacher imports
		},
	}

	// Process batch as teacher
	res, appErr := svc.BatchCreateUsers(ctx, "teacher", req)
	if appErr != nil {
		t.Fatalf("unexpected error processing batch: %v", appErr)
	}

	if res.TotalRequested != 7 {
		t.Errorf("expected totalRequested 7, got %d", res.TotalRequested)
	}
	if res.CreatedCount != 2 { // valid1, valid2
		t.Errorf("expected createdCount 2, got %d", res.CreatedCount)
	}
	if res.FailedCount != 5 {
		t.Errorf("expected failedCount 5, got %d", res.FailedCount)
	}
	if len(res.CreatedUsers) != 2 {
		t.Errorf("expected 2 created users, got %d", len(res.CreatedUsers))
	}
	if len(res.Errors) != 5 {
		t.Errorf("expected 5 errors, got %d", len(res.Errors))
	}
}

func TestBatchCreateUsers_StudentForbidden(t *testing.T) {
	svc, _ := setupUserService()
	ctx := context.Background()

	_, appErr := svc.BatchCreateUsers(ctx, "student", user.BatchCreateUsersRequest{
		Users: []user.BatchUserItem{
			{Email: "student@encertia.cat", FirstName: "A", LastName: "B", Role: user.RoleStudent},
		},
	})
	if appErr == nil {
		t.Fatal("expected forbidden error for student batch import, got nil")
	}
	if appErr.StatusCode != 403 {
		t.Errorf("expected 403, got %d", appErr.StatusCode)
	}
}

// HU-USER-04: GetUserByID
func TestGetUserByID_Permissions(t *testing.T) {
	svc, _ := setupUserService()
	ctx := context.Background()

	adminUser, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "admin@encertia.cat",
		Password:  "Password123!",
		FirstName: "Admin",
		LastName:  "User",
		Role:      user.RoleAdmin,
	})
	teacherUser, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "teacher@encertia.cat",
		Password:  "Password123!",
		FirstName: "Teacher",
		LastName:  "User",
		Role:      user.RoleTeacher,
	})
	studentUser, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "student@encertia.cat",
		Password:  "Password123!",
		FirstName: "Student",
		LastName:  "User",
		Role:      user.RoleStudent,
	})

	// 1. Admin can view any user
	_, errAdmin := svc.GetUserByID(ctx, adminUser.User.ID, "admin", studentUser.User.ID)
	if errAdmin != nil {
		t.Errorf("admin should be able to view student: %v", errAdmin)
	}

	// 2. Teacher can view student
	_, errTeacherViewStudent := svc.GetUserByID(ctx, teacherUser.User.ID, "teacher", studentUser.User.ID)
	if errTeacherViewStudent != nil {
		t.Errorf("teacher should be able to view student: %v", errTeacherViewStudent)
	}

	// 3. Teacher cannot view admin
	_, errTeacherViewAdmin := svc.GetUserByID(ctx, teacherUser.User.ID, "teacher", adminUser.User.ID)
	if errTeacherViewAdmin == nil {
		t.Error("teacher should not be able to view admin")
	}

	// 4. Student can view self
	_, errStudentSelf := svc.GetUserByID(ctx, studentUser.User.ID, "student", studentUser.User.ID)
	if errStudentSelf != nil {
		t.Errorf("student should be able to view self: %v", errStudentSelf)
	}

	// 5. Student cannot view other user
	_, errStudentOther := svc.GetUserByID(ctx, studentUser.User.ID, "student", teacherUser.User.ID)
	if errStudentOther == nil {
		t.Error("student should not be able to view other user")
	}

	// 6. Non-existent user
	_, errNotFound := svc.GetUserByID(ctx, adminUser.User.ID, "admin", uuid.New())
	if errNotFound == nil || errNotFound.StatusCode != 404 {
		t.Errorf("expected 404 for non-existent user, got %v", errNotFound)
	}
}

// HU-USER-05: UpdateUser
func TestUpdateUser_SelfEditActiveUser(t *testing.T) {
	svc, _ := setupUserService()
	ctx := context.Background()

	created, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "student@encertia.cat",
		Password:  "Password123!",
		FirstName: "Original",
		LastName:  "Name",
		Role:      user.RoleStudent,
	})

	newName := "UpdatedName"
	newEmail := "student.new@encertia.cat"
	res, appErr := svc.UpdateUser(ctx, created.User.ID, "student", created.User.ID, user.UpdateUserInput{
		FirstName: &newName,
		Email:     &newEmail,
	})
	if appErr != nil {
		t.Fatalf("unexpected error on self-edit: %v", appErr)
	}

	if res.User.FirstName != "UpdatedName" {
		t.Errorf("expected updated first name, got %s", res.User.FirstName)
	}
	if res.User.Email != "student.new@encertia.cat" {
		t.Errorf("expected updated email, got %s", res.User.Email)
	}
}

func TestUpdateUser_RoleAndActiveRestrictions(t *testing.T) {
	svc, repo := setupUserService()
	ctx := context.Background()

	student, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "student@encertia.cat",
		Password:  "Password123!",
		FirstName: "Student",
		LastName:  "User",
		Role:      user.RoleStudent,
	})

	// 1. Non-admin attempting to change role -> 403
	newRole := user.RoleAdmin
	_, appErrRole := svc.UpdateUser(ctx, student.User.ID, "student", student.User.ID, user.UpdateUserInput{
		Role: &newRole,
	})
	if appErrRole == nil || appErrRole.StatusCode != 403 {
		t.Errorf("expected 403 when non-admin changes role, got %v", appErrRole)
	}

	// 2. Non-admin attempting to change isActive -> 403
	newActive := false
	_, appErrActive := svc.UpdateUser(ctx, student.User.ID, "student", student.User.ID, user.UpdateUserInput{
		IsActive: &newActive,
	})
	if appErrActive == nil || appErrActive.StatusCode != 403 {
		t.Errorf("expected 403 when non-admin changes isActive, got %v", appErrActive)
	}

	// 3. Non-admin attempting to edit another user -> 403
	otherUser, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "other@encertia.cat",
		Password:  "Password123!",
		FirstName: "Other",
		LastName:  "User",
		Role:      user.RoleStudent,
	})
	fn := "Hacked"
	_, appErrOther := svc.UpdateUser(ctx, student.User.ID, "student", otherUser.User.ID, user.UpdateUserInput{
		FirstName: &fn,
	})
	if appErrOther == nil || appErrOther.StatusCode != 403 {
		t.Errorf("expected 403 when editing other user, got %v", appErrOther)
	}

	// 4. Inactive user attempting self-edit -> 403
	uDB, _ := repo.GetUserByID(ctx, student.User.ID)
	uDB.IsActive = false
	_, appErrInactive := svc.UpdateUser(ctx, student.User.ID, "student", student.User.ID, user.UpdateUserInput{
		FirstName: &fn,
	})
	if appErrInactive == nil || appErrInactive.StatusCode != 403 {
		t.Errorf("expected 403 when inactive user edits self, got %v", appErrInactive)
	}

	// 5. Admin updating role and isActive on any user -> 200
	adminActive := true
	adminRole := user.RoleTeacher
	resAdmin, appErrAdmin := svc.UpdateUser(ctx, uuid.New(), "admin", student.User.ID, user.UpdateUserInput{
		IsActive: &adminActive,
		Role:     &adminRole,
	})
	if appErrAdmin != nil {
		t.Fatalf("unexpected error when admin updates user: %v", appErrAdmin)
	}
	if resAdmin.User.Role != user.RoleTeacher {
		t.Errorf("expected updated role teacher, got %s", resAdmin.User.Role)
	}
	if !resAdmin.User.IsActive {
		t.Error("expected isActive to be reactivated to true")
	}
}

func TestUpdateUser_EmailConflict(t *testing.T) {
	svc, _ := setupUserService()
	ctx := context.Background()

	u1, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "user1@encertia.cat",
		Password:  "Password123!",
		FirstName: "User",
		LastName:  "One",
		Role:      user.RoleStudent,
	})
	u2, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "user2@encertia.cat",
		Password:  "Password123!",
		FirstName: "User",
		LastName:  "Two",
		Role:      user.RoleStudent,
	})

	emailConflict := u1.User.Email
	_, appErr := svc.UpdateUser(ctx, u2.User.ID, "student", u2.User.ID, user.UpdateUserInput{
		Email: &emailConflict,
	})
	if appErr == nil || appErr.StatusCode != 409 {
		t.Errorf("expected 409 email conflict, got %v", appErr)
	}
}

// HU-USER-06: ResetPassword
func TestResetPassword_PermissionsAndValidation(t *testing.T) {
	svc, repo := setupUserService()
	ctx := context.Background()

	admin, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "admin@encertia.cat",
		Password:  "Password123!",
		FirstName: "Admin",
		LastName:  "User",
		Role:      user.RoleAdmin,
	})
	teacher, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "teacher@encertia.cat",
		Password:  "Password123!",
		FirstName: "Teacher",
		LastName:  "User",
		Role:      user.RoleTeacher,
	})
	student, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "student@encertia.cat",
		Password:  "Password123!",
		FirstName: "Student",
		LastName:  "User",
		Role:      user.RoleStudent,
	})

	// 1. Admin resets student password -> 200
	msgAdmin, appErr := svc.ResetPassword(ctx, admin.User.ID, "admin", student.User.ID, user.ResetPasswordInput{
		NewPassword: "BrandNewPassword123!",
	})
	if appErr != nil {
		t.Fatalf("unexpected error when admin resets password: %v", appErr)
	}
	if msgAdmin.Message != "Contrasenya actualitzada correctament." {
		t.Errorf("expected success message, got %s", msgAdmin.Message)
	}

	// 2. Teacher resets student password -> 200
	_, appErrTeacher := svc.ResetPassword(ctx, teacher.User.ID, "teacher", student.User.ID, user.ResetPasswordInput{
		NewPassword: "BrandNewPassword456!",
	})
	if appErrTeacher != nil {
		t.Fatalf("unexpected error when teacher resets student password: %v", appErrTeacher)
	}

	// 3. Teacher attempts to reset teacher password -> 403
	_, appErrTeacherOnTeacher := svc.ResetPassword(ctx, teacher.User.ID, "teacher", teacher.User.ID, user.ResetPasswordInput{
		NewPassword: "BrandNewPassword789!",
	})
	if appErrTeacherOnTeacher == nil || appErrTeacherOnTeacher.StatusCode != 403 {
		t.Errorf("expected 403 when teacher resets non-student password, got %v", appErrTeacherOnTeacher)
	}

	// 4. Student attempts to reset password -> 403
	_, appErrStudent := svc.ResetPassword(ctx, student.User.ID, "student", student.User.ID, user.ResetPasswordInput{
		NewPassword: "BrandNewPassword000!",
	})
	if appErrStudent == nil || appErrStudent.StatusCode != 403 {
		t.Errorf("expected 403 when student uses reset password endpoint, got %v", appErrStudent)
	}

	// 5. Short password -> 400
	_, appErrShort := svc.ResetPassword(ctx, admin.User.ID, "admin", student.User.ID, user.ResetPasswordInput{
		NewPassword: "short",
	})
	if appErrShort == nil || appErrShort.StatusCode != 400 {
		t.Errorf("expected 400 when password is short, got %v", appErrShort)
	}

	// Check token revocation recorded
	if _, revoked := repo.revokedUsers[student.User.ID]; !revoked {
		t.Error("expected tokens to be revoked on password reset")
	}
}

// HU-USER-07: DeleteUser
func TestDeleteUser_SoftDeleteAndPermissions(t *testing.T) {
	svc, repo := setupUserService()
	ctx := context.Background()

	admin, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "admin@encertia.cat",
		Password:  "Password123!",
		FirstName: "Admin",
		LastName:  "User",
		Role:      user.RoleAdmin,
	})
	teacher, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "teacher@encertia.cat",
		Password:  "Password123!",
		FirstName: "Teacher",
		LastName:  "User",
		Role:      user.RoleTeacher,
	})
	student, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "student@encertia.cat",
		Password:  "Password123!",
		FirstName: "Student",
		LastName:  "User",
		Role:      user.RoleStudent,
	})

	// 1. Teacher cannot delete user -> 403
	_, appErrTeacher := svc.DeleteUser(ctx, teacher.User.ID, "teacher", student.User.ID)
	if appErrTeacher == nil || appErrTeacher.StatusCode != 403 {
		t.Errorf("expected 403 when teacher deletes user, got %v", appErrTeacher)
	}

	// 2. Student cannot delete user -> 403
	_, appErrStudent := svc.DeleteUser(ctx, student.User.ID, "student", student.User.ID)
	if appErrStudent == nil || appErrStudent.StatusCode != 403 {
		t.Errorf("expected 403 when student deletes user, got %v", appErrStudent)
	}

	// 3. Admin soft-deletes user -> 200
	msg, appErrAdmin := svc.DeleteUser(ctx, admin.User.ID, "admin", student.User.ID)
	if appErrAdmin != nil {
		t.Fatalf("unexpected error when admin deletes user: %v", appErrAdmin)
	}
	if msg.Message != "Usuari donat de baixa correctament." {
		t.Errorf("expected success message, got %s", msg.Message)
	}

	// User should now return 404
	_, appErrGet := svc.GetUserByID(ctx, admin.User.ID, "admin", student.User.ID)
	if appErrGet == nil || appErrGet.StatusCode != 404 {
		t.Errorf("expected 404 for soft-deleted user, got %v", appErrGet)
	}

	// Tokens should be revoked
	if _, revoked := repo.revokedUsers[student.User.ID]; !revoked {
		t.Error("expected tokens to be revoked on user deletion")
	}

	// Deleting non-existent user -> 404
	_, appErrNotFound := svc.DeleteUser(ctx, admin.User.ID, "admin", uuid.New())
	if appErrNotFound == nil || appErrNotFound.StatusCode != 404 {
		t.Errorf("expected 404 when deleting non-existent user, got %v", appErrNotFound)
	}
}

func TestUser_LanguageOptionsAndUpdates(t *testing.T) {
	svc, _ := setupUserService()
	ctx := context.Background()

	// 1. CreateUser with default language
	u1, appErr := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "lang1@encertia.cat",
		Password:  "Password123!",
		FirstName: "User",
		LastName:  "One",
		Role:      user.RoleStudent,
	})
	if appErr != nil {
		t.Fatalf("unexpected error creating user: %v", appErr)
	}
	if u1.User.Language != "ca" {
		t.Errorf("expected default language 'ca', got %s", u1.User.Language)
	}

	// 2. CreateUser with custom valid language 'es'
	u2, appErrEs := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "lang2@encertia.cat",
		Password:  "Password123!",
		FirstName: "User",
		LastName:  "Two",
		Role:      user.RoleStudent,
		Language:  "es",
	})
	if appErrEs != nil {
		t.Fatalf("unexpected error: %v", appErrEs)
	}
	if u2.User.Language != "es" {
		t.Errorf("expected language 'es', got %s", u2.User.Language)
	}

	// 3. CreateUser with invalid language 'fr'
	_, appErrFr := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "lang3@encertia.cat",
		Password:  "Password123!",
		FirstName: "User",
		LastName:  "Three",
		Role:      user.RoleStudent,
		Language:  "fr",
	})
	if appErrFr == nil || appErrFr.StatusCode != 400 {
		t.Errorf("expected 400 for invalid language, got %v", appErrFr)
	}

	// 4. UpdateUser language to 'en'
	newLang := "en"
	uUpdated, appErrUpd := svc.UpdateUser(ctx, u1.User.ID, "student", u1.User.ID, user.UpdateUserInput{
		Language: &newLang,
	})
	if appErrUpd != nil {
		t.Fatalf("unexpected error updating language: %v", appErrUpd)
	}
	if uUpdated.User.Language != "en" {
		t.Errorf("expected updated language 'en', got %s", uUpdated.User.Language)
	}

	// 5. UpdateUser with invalid language
	invalidLang := "de"
	_, appErrInvalidUpd := svc.UpdateUser(ctx, u1.User.ID, "student", u1.User.ID, user.UpdateUserInput{
		Language: &invalidLang,
	})
	if appErrInvalidUpd == nil || appErrInvalidUpd.StatusCode != 400 {
		t.Errorf("expected 400 when updating invalid language, got %v", appErrInvalidUpd)
	}
}
