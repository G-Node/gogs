package dmp_schema

type MetiDmp struct {
	Index         int
	Title         string
	Description   string
	Manager       string
	DataType      string
	ReleaseLevel  int
	ConcealReason string
	ConcealPeriod string
	Acquirer      string
	AcquireMethod string
}
type MetiDmpInfo struct {
	Schema         string
	DmpType        string
	AgreementTitle string
	AgreementDate  string // FIXME: to date type
	SubmitDate     string // FIXME: to date type
	CorporateName  string
	Researches     []MetiDmp
}
