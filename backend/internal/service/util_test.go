package service

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func objectIDZero() primitive.ObjectID { return primitive.NilObjectID }

func errorsAs(err error, target interface{}) bool {
	return errors.As(err, target)
}
