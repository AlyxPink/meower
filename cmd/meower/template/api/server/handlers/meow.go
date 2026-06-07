package handlers

import (
	"context"
	"encoding/hex"

	"TEMPLATE_MODULE_PATH/api/db"
	meowV1 "TEMPLATE_MODULE_PATH/api/proto/meow/v1"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type meowServiceServer struct {
	meowV1.UnimplementedMeowServiceServer
	db *pgxpool.Pool
}

func NewMeowerServer(db *pgxpool.Pool) meowV1.MeowServiceServer {
	return &meowServiceServer{db: db}
}

// uuidHex renders a pgtype.UUID as the 32-char hex string used in URLs and
// proto ids. It round-trips through parseUUID (see user.go).
func uuidHex(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return hex.EncodeToString(u.Bytes[:])
}

// pgText returns the string value of a nullable text column, or "" if NULL.
func pgText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// authorOf looks up a user's handle and display name for embedding in a Meow
// response. It is best-effort: an unknown or anonymous author yields empty
// strings rather than an error, so posting never fails on author lookup.
func (s *meowServiceServer) authorOf(ctx context.Context, userID pgtype.UUID) (username, displayName string) {
	if !userID.Valid {
		return "", ""
	}
	user, err := db.New(s.db).GetUserById(ctx, userID)
	if err != nil {
		return "", ""
	}
	return user.Username, user.DisplayName
}

func (s *meowServiceServer) CreateMeow(ctx context.Context, req *meowV1.CreateMeowRequest) (*meowV1.CreateMeowResponse, error) {
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}

	// author_id is optional (anonymous meows allowed). When present it must be a
	// valid UUID; an invalid one is a client error, not a silent anonymous post.
	var authorID pgtype.UUID
	if req.AuthorId != "" {
		parsed, err := parseUUID(req.AuthorId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid author_id: %v", err)
		}
		authorID = parsed
	}

	meow, err := db.New(s.db).CreateMeow(ctx, db.CreateMeowParams{
		UserID:  authorID,
		Content: req.Content,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create meow: %v", err)
	}

	username, displayName := s.authorOf(ctx, meow.UserID)
	return &meowV1.CreateMeowResponse{
		Meow: &meowV1.Meow{
			Id:                uuidHex(meow.ID),
			Content:           meow.Content,
			CreatedAt:         timestamppb.New(meow.CreatedAt.Time),
			AuthorId:          uuidHex(meow.UserID),
			AuthorUsername:    username,
			AuthorDisplayName: displayName,
		},
	}, nil
}

func (s *meowServiceServer) GetMeow(ctx context.Context, req *meowV1.GetMeowRequest) (*meowV1.GetMeowResponse, error) {
	id, err := parseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid meow ID: %v", err)
	}

	meow, err := db.New(s.db).ShowMeow(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "meow not found: %v", err)
	}

	return &meowV1.GetMeowResponse{
		Meow: &meowV1.Meow{
			Id:                uuidHex(meow.ID),
			Content:           meow.Content,
			CreatedAt:         timestamppb.New(meow.CreatedAt.Time),
			AuthorId:          uuidHex(meow.UserID),
			AuthorUsername:    pgText(meow.AuthorUsername),
			AuthorDisplayName: pgText(meow.AuthorDisplayName),
		},
	}, nil
}

func (s *meowServiceServer) IndexMeow(ctx context.Context, req *meowV1.IndexMeowRequest) (*meowV1.IndexMeowResponse, error) {
	meows, err := db.New(s.db).IndexMeows(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list meows: %v", err)
	}

	resp := make([]*meowV1.Meow, 0, len(meows))
	for _, meow := range meows {
		resp = append(resp, &meowV1.Meow{
			Id:                uuidHex(meow.ID),
			Content:           meow.Content,
			CreatedAt:         timestamppb.New(meow.CreatedAt.Time),
			AuthorId:          uuidHex(meow.UserID),
			AuthorUsername:    pgText(meow.AuthorUsername),
			AuthorDisplayName: pgText(meow.AuthorDisplayName),
		})
	}

	return &meowV1.IndexMeowResponse{Meows: resp}, nil
}

func (s *meowServiceServer) UpdateMeow(ctx context.Context, req *meowV1.UpdateMeowRequest) (*meowV1.UpdateMeowResponse, error) {
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}

	id, err := parseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid meow ID: %v", err)
	}

	meow, err := db.New(s.db).UpdateMeow(ctx, db.UpdateMeowParams{
		ID:      id,
		Content: req.Content,
	})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "meow not found: %v", err)
	}

	username, displayName := s.authorOf(ctx, meow.UserID)
	return &meowV1.UpdateMeowResponse{
		Meow: &meowV1.Meow{
			Id:                uuidHex(meow.ID),
			Content:           meow.Content,
			CreatedAt:         timestamppb.New(meow.CreatedAt.Time),
			AuthorId:          uuidHex(meow.UserID),
			AuthorUsername:    username,
			AuthorDisplayName: displayName,
		},
	}, nil
}

func (s *meowServiceServer) DeleteMeow(ctx context.Context, req *meowV1.DeleteMeowRequest) (*meowV1.DeleteMeowResponse, error) {
	id, err := parseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid meow ID: %v", err)
	}

	if err := db.New(s.db).DeleteMeow(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete meow: %v", err)
	}

	return &meowV1.DeleteMeowResponse{}, nil
}
