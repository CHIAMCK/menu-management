package repository

import (
	"errors"

	"gorm.io/gorm"
)

func mapRecordNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}

	return err
}
