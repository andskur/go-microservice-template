// Package grpc provides the gRPC module implementation.
package grpc

import (
	"microservice-template/internal/models"
	userProto "microservice-template/protocols/user"

	"github.com/gofrs/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func userToProto(u *models.User) *userProto.User {
	if u == nil {
		return nil
	}

	return &userProto.User{
		Uuid:      u.UUID.Bytes(),
		Email:     u.Email,
		Name:      u.Name,
		Status:    userStatusToProto(u.Status),
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

func userFromProto(pb *userProto.User) *models.User {
	if pb == nil {
		return nil
	}

	return &models.User{
		UUID:      uuid.FromBytesOrNil(pb.Uuid),
		Email:     pb.Email,
		Name:      pb.Name,
		Status:    userStatusFromProto(pb.Status),
		CreatedAt: pb.CreatedAt.AsTime(),
		UpdatedAt: pb.UpdatedAt.AsTime(),
	}
}

func userStatusToProto(s models.UserStatus) userProto.UserStatus {
	switch s {
	case models.UserActive:
		return userProto.UserStatus_USER_STATUS_ACTIVE
	case models.UserDeleted:
		return userProto.UserStatus_USER_STATUS_DELETED
	default:
		return userProto.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

func userStatusFromProto(s userProto.UserStatus) models.UserStatus {
	switch s {
	case userProto.UserStatus_USER_STATUS_ACTIVE:
		return models.UserActive
	case userProto.UserStatus_USER_STATUS_DELETED:
		return models.UserDeleted
	case userProto.UserStatus_USER_STATUS_UNSPECIFIED:
		return models.UserActive
	default:
		return models.UserActive
	}
}
