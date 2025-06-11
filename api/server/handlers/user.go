package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/AlyxPink/meower/api/db"
	userV1 "github.com/AlyxPink/meower/api/proto/user/v1"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userServiceServer struct {
	userV1.UnimplementedUserServiceServer
	db *pgxpool.Pool
}

func NewUserServer(db *pgxpool.Pool) userV1.UserServiceServer {
	return &userServiceServer{db: db}
}

// Helper function to convert DB user to proto user
func (s *userServiceServer) dbUserToProto(user db.User) *userV1.User {
	protoUser := &userV1.User{
		Id:                  hex.EncodeToString(user.ID.Bytes[:]),
		Username:            user.Username,
		DisplayName:         user.DisplayName,
		Email:               user.Email,
		EmailVerified:       user.EmailVerified.Bool,
		FailedLoginAttempts: user.FailedLoginAttempts.Int32,
		AccountLocked:       user.AccountLocked.Bool,
	}

	if user.CreatedAt.Valid {
		protoUser.CreatedAt = timestamppb.New(user.CreatedAt.Time)
	}

	if user.LastLoginAt.Valid {
		protoUser.LastLoginAt = timestamppb.New(user.LastLoginAt.Time)
	}

	return protoUser
}

// Helper function to hash password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Helper function to check password
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Helper function to generate random token
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Helper function to parse UUID string to pgtype.UUID
func parseUUID(uuidStr string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(uuidStr)
	return uuid, err
}

// CreateUser creates a new user
func (s *userServiceServer) CreateUser(ctx context.Context, req *userV1.CreateUserRequest) (*userV1.CreateUserResponse, error) {
	// Hash the password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	// Create user in database
	user, err := db.New(s.db).CreateUser(ctx, db.CreateUserParams{
		Username:     req.Username,
		DisplayName:  req.DisplayName,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			if strings.Contains(err.Error(), "username") {
				return nil, status.Errorf(codes.AlreadyExists, "username already exists")
			}
			if strings.Contains(err.Error(), "email") {
				return nil, status.Errorf(codes.AlreadyExists, "email already exists")
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &userV1.CreateUserResponse{
		User: s.dbUserToProto(user),
	}, nil
}

// GetUser gets a user by ID
func (s *userServiceServer) GetUser(ctx context.Context, req *userV1.GetUserRequest) (*userV1.GetUserResponse, error) {
	uuid, err := parseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	user, err := db.New(s.db).GetUserById(ctx, uuid)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &userV1.GetUserResponse{
		User: s.dbUserToProto(user),
	}, nil
}

// GetUserByEmail gets a user by email
func (s *userServiceServer) GetUserByEmail(ctx context.Context, req *userV1.GetUserByEmailRequest) (*userV1.GetUserByEmailResponse, error) {
	user, err := db.New(s.db).GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &userV1.GetUserByEmailResponse{
		User: s.dbUserToProto(user),
	}, nil
}

// GetUserByUsername gets a user by username
func (s *userServiceServer) GetUserByUsername(ctx context.Context, req *userV1.GetUserByUsernameRequest) (*userV1.GetUserByUsernameResponse, error) {
	user, err := db.New(s.db).GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &userV1.GetUserByUsernameResponse{
		User: s.dbUserToProto(user),
	}, nil
}

// UpdateUser updates a user
func (s *userServiceServer) UpdateUser(ctx context.Context, req *userV1.UpdateUserRequest) (*userV1.UpdateUserResponse, error) {
	uuid, err := parseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	// Prepare update parameters
	updateParams := db.UpdateUserParams{
		ID: uuid,
	}

	// Handle optional fields
	if req.DisplayName != nil {
		updateParams.DisplayName = req.DisplayName.Value
	}

	if req.Email != nil {
		updateParams.Email = req.Email.Value
	}

	if req.EmailVerified != nil {
		updateParams.EmailVerified = pgtype.Bool{Bool: req.EmailVerified.Value, Valid: true}
	}

	if req.Password != nil {
		hashedPassword, err := hashPassword(req.Password.Value)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
		}
		updateParams.PasswordHash = hashedPassword
	}

	user, err := db.New(s.db).UpdateUser(ctx, updateParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &userV1.UpdateUserResponse{
		User: s.dbUserToProto(user),
	}, nil
}

// DeleteUser deletes a user
func (s *userServiceServer) DeleteUser(ctx context.Context, req *userV1.DeleteUserRequest) (*userV1.DeleteUserResponse, error) {
	uuid, err := parseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	err = db.New(s.db).DeleteUser(ctx, uuid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &userV1.DeleteUserResponse{}, nil
}

// ListUsers lists users with pagination
func (s *userServiceServer) ListUsers(ctx context.Context, req *userV1.ListUsersRequest) (*userV1.ListUsersResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	users, err := db.New(s.db).IndexUsers(ctx, db.IndexUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	var protoUsers []*userV1.User
	for _, user := range users {
		protoUsers = append(protoUsers, s.dbUserToProto(user))
	}

	return &userV1.ListUsersResponse{
		Users: protoUsers,
	}, nil
}

// Login authenticates a user
func (s *userServiceServer) Login(ctx context.Context, req *userV1.LoginRequest) (*userV1.LoginResponse, error) {
	var user db.User
	var err error

	// Try to find user by email or username
	if strings.Contains(req.UsernameOrEmail, "@") {
		user, err = db.New(s.db).GetUserByEmail(ctx, req.UsernameOrEmail)
	} else {
		user, err = db.New(s.db).GetUserByUsername(ctx, req.UsernameOrEmail)
	}

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	// Check if account is locked
	if user.AccountLocked.Bool {
		return nil, status.Errorf(codes.PermissionDenied, "account is locked")
	}

	// Check password
	if !checkPasswordHash(req.Password, user.PasswordHash) {
		// Increment failed login attempts
		attempts := user.FailedLoginAttempts.Int32 + 1
		locked := attempts >= 5

		_, err = db.New(s.db).UpdateLoginAttempts(ctx, db.UpdateLoginAttemptsParams{
			ID:                  user.ID,
			FailedLoginAttempts: pgtype.Int4{Int32: attempts, Valid: true},
			AccountLocked:       pgtype.Bool{Bool: locked, Valid: true},
		})
		if err != nil {
			// Log error but don't expose it
			fmt.Printf("Failed to update login attempts: %v\n", err)
		}

		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	// Reset failed login attempts and update last login
	_, err = db.New(s.db).UpdateLoginAttempts(ctx, db.UpdateLoginAttemptsParams{
		ID:                  user.ID,
		FailedLoginAttempts: pgtype.Int4{Int32: 0, Valid: true},
		AccountLocked:       pgtype.Bool{Bool: false, Valid: true},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update login attempts: %v", err)
	}

	// Update last login time
	_, err = db.New(s.db).UpdateUser(ctx, db.UpdateUserParams{
		ID:          user.ID,
		LastLoginAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	})
	if err != nil {
		// Log error but don't fail the login
		fmt.Printf("Failed to update last login time: %v\n", err)
	}

	// Generate a simple token (in production, use JWT)
	token, err := generateToken(32)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	return &userV1.LoginResponse{
		Token: token,
		User:  s.dbUserToProto(user),
	}, nil
}

// Logout logs out a user
func (s *userServiceServer) Logout(ctx context.Context, req *userV1.LogoutRequest) (*userV1.LogoutResponse, error) {
	// In a real implementation, you would invalidate the token
	// For now, just return success
	return &userV1.LogoutResponse{
		Success: true,
	}, nil
}

// RequestPasswordReset initiates a password reset
func (s *userServiceServer) RequestPasswordReset(ctx context.Context, req *userV1.RequestPasswordResetRequest) (*userV1.RequestPasswordResetResponse, error) {
	user, err := db.New(s.db).GetUserByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists or not
		return &userV1.RequestPasswordResetResponse{
			Success: true,
		}, nil
	}

	// Generate reset token
	token, err := generateToken(32)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate reset token: %v", err)
	}

	// Set token expiration to 1 hour
	expires := time.Now().Add(time.Hour)

	_, err = db.New(s.db).UpdatePasswordResetToken(ctx, db.UpdatePasswordResetTokenParams{
		ID:                   user.ID,
		ResetPasswordToken:   pgtype.Text{String: token, Valid: true},
		ResetPasswordExpires: pgtype.Timestamp{Time: expires, Valid: true},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save reset token: %v", err)
	}

	// In a real implementation, send email with token
	fmt.Printf("Password reset token for %s: %s\n", req.Email, token)

	return &userV1.RequestPasswordResetResponse{
		Success: true,
	}, nil
}

// ResetPassword resets a user's password using a token
func (s *userServiceServer) ResetPassword(ctx context.Context, req *userV1.ResetPasswordRequest) (*userV1.ResetPasswordResponse, error) {
	user, err := db.New(s.db).GetUserByResetToken(ctx, pgtype.Text{String: req.Token, Valid: true})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired reset token")
	}

	// Hash new password
	hashedPassword, err := hashPassword(req.NewPassword)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	// Update password and clear reset token
	_, err = db.New(s.db).UpdateUser(ctx, db.UpdateUserParams{
		ID:           user.ID,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update password: %v", err)
	}

	// Clear reset token
	_, err = db.New(s.db).UpdatePasswordResetToken(ctx, db.UpdatePasswordResetTokenParams{
		ID:                   user.ID,
		ResetPasswordToken:   pgtype.Text{Valid: false},
		ResetPasswordExpires: pgtype.Timestamp{Valid: false},
	})
	if err != nil {
		// Log error but don't fail the reset
		fmt.Printf("Failed to clear reset token: %v\n", err)
	}

	return &userV1.ResetPasswordResponse{
		Success: true,
	}, nil
}

// VerifyEmail verifies a user's email
func (s *userServiceServer) VerifyEmail(ctx context.Context, req *userV1.VerifyEmailRequest) (*userV1.VerifyEmailResponse, error) {
	uuid, err := parseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	_, err = db.New(s.db).VerifyEmail(ctx, uuid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify email: %v", err)
	}

	return &userV1.VerifyEmailResponse{
		Success: true,
	}, nil
}
