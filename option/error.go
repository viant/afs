package option

const (
	//ErrorTypeDownload download error type
	ErrorTypeDownload = "download"
	//ErrorTypeUpload upload error type
	ErrorTypeUpload = "upload"
	//ErrorTypeReader reader error type
	ErrorTypeReader = "reader"
)

//Error represents a simulation error
type Error struct {
	Type  string
	Error error
}

//Errors represents simulation errors
type Errors []*Error

//NewUploadError creates an upload error
func NewUploadError(err error) *Error {
	return &Error{
		Type:  ErrorTypeUpload,
		Error: err,
	}
}

//NewDownloadError creates a download error
func NewDownloadError(err error) *Error {
	return &Error{
		Type:  ErrorTypeDownload,
		Error: err,
	}
}

//NewReaderError creates a reader error
func NewReaderError(err error) *Error {
	return &Error{
		Type:  ErrorTypeReader,
		Error: err,
	}
}

//NewErrors creates an error slice for supplied errors
func NewErrors(errors ...*Error) []*Error {
	return errors
}
