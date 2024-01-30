package s3mock

import (
	"errors"
	"io"
	"multi-tenant-HR-information-system-backend/httperror"
)

type s3Mock struct {
	
}

func NewS3Mock() *s3Mock {
	return &s3Mock{}
}

func (s* s3Mock) UploadResume(file io.Reader, jobApplicationId string, firstName string, lastName string, fileExt string) (url string, err error) {
	return "", httperror.NewInternalServerError(errors.New("this is a mocked error"))
}