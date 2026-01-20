package grpcclient

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"

	"microservice-template/internal/models"
	proto "microservice-template/protocols/userservice"
)

// UserToProto converts domain User to proto User.
func UserToProto(user *models.User) *proto.User {
	if user == nil {
		return nil
	}

	return &proto.User{
		Uuid:      user.UUID.Bytes(),
		Email:     user.Email,
		Name:      user.Name,
		Status:    UserStatusToProto(user.Status),
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}
}

// UserFromProto converts proto User to domain User.
func UserFromProto(pb *proto.User) (*models.User, error) {
	if pb == nil {
		return nil, fmt.Errorf("proto user is nil")
	}

	userUUID, err := uuid.FromBytes(pb.Uuid)
	if err != nil {
		return nil, fmt.Errorf("parse uuid: %w", err)
	}

	status, err := UserStatusFromProto(pb.Status)
	if err != nil {
		return nil, fmt.Errorf("parse status: %w", err)
	}

	return &models.User{
		UUID:      userUUID,
		Email:     pb.Email,
		Name:      pb.Name,
		Status:    status,
		CreatedAt: time.Unix(pb.CreatedAt, 0),
		UpdatedAt: time.Unix(pb.UpdatedAt, 0),
	}, nil
}

// UserStatusToProto converts domain UserStatus to proto UserStatus.
func UserStatusToProto(status models.UserStatus) proto.UserStatus {
	switch status {
	case models.UserActive:
		return proto.UserStatus_USER_STATUS_ACTIVE
	case models.UserDeleted:
		return proto.UserStatus_USER_STATUS_DELETED
	default:
		return proto.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

// UserStatusFromProto converts proto UserStatus to domain UserStatus.
func UserStatusFromProto(status proto.UserStatus) (models.UserStatus, error) {
	switch status {
	case proto.UserStatus_USER_STATUS_ACTIVE:
		return models.UserActive, nil
	case proto.UserStatus_USER_STATUS_DELETED:
		return models.UserDeleted, nil
	case proto.UserStatus_USER_STATUS_UNSPECIFIED:
		return models.UserActive, nil // Default to active
	default:
		return 0, fmt.Errorf("unknown user status: %d", status)
	}
}
