package logic

import (
	"errors"
	"net/http"

	"history/internal/errx"
	"history/model"
)

func validateUserID(userID int64) error {
	if userID <= 0 {
		return errx.RPCError(http.StatusBadRequest, errx.CodeUserIDInvalid, "user_id invalid")
	}
	return nil
}

func validateMedia(mediaType int64, mediaID int64) error {
	if mediaType != model.MediaTypePost && mediaType != model.MediaTypeVideo {
		return errx.RPCError(http.StatusBadRequest, errx.CodeMediaTypeInvalid, "media_type invalid")
	}
	if mediaID <= 0 {
		return errx.RPCError(http.StatusBadRequest, errx.CodeMediaIDInvalid, "media_id invalid")
	}
	return nil
}

func mapModelError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, model.ErrInvalidProgress) {
		return errx.RPCError(http.StatusBadRequest, errx.CodeProgressInvalid, "progress invalid")
	}
	return errx.RPCError(http.StatusInternalServerError, errx.CodeInternalDefault, err.Error())
}
