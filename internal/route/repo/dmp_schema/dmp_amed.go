package dmp_schema

type AmedDmpInfo struct {
	Schema         string             `json:"schema"`
	CreateDate     string             `json:"createDate"`
	Project        AmedProject        `json:"project"`
	Required       AmedRequired       `json:"required"`
	Researches     AmedResearches     `json:"researches"`
	ForPublication AmedForPublication `json:"forPublication"`
	Researchers    AmedResearchers    `json:"researchers"`
}
type AmedRepresentative struct {
	BelongTo string `json:"belongTo"`
	Post     string `json:"post"`
	Name     string `json:"name"`
}
type AmedProject struct {
	FiscalYear     int                `json:"fiscalYear"`
	Title          string             `json:"title"`
	ProblemName    string             `json:"problemName"`
	Representative AmedRepresentative `json:"representative"`
}
type AmedRequired struct {
	HasRegistNecessity bool   `json:"hasRegistNecessity"`
	NoRegistReason     string `json:"noRegistReason"`
}
type AmedData struct {
	Title          string `json:"title"`
	ReleasePolicy  string `json:"releasePolicy"`
	ConcealReason  string `json:"concealReason"`
	RepositoryType string `json:"repositoryType"`
	RepositoryName string `json:"repositoryName"`
	DataAmount     string `json:"dataAmount"`
}
type AmedResearches struct {
	Description string     `json:"description"`
	Data        []AmedData `json:"data"`
}
type AmedForPublication struct {
	HasOfferPolicy bool   `json:"hasOfferPolicy"`
	PolicyName     string `json:"policyName"`
}
type AmedPersonal struct {
	BelongTo string `json:"belongTo"`
	Post     string `json:"post"`
	Name     string `json:"name"`
}
type AmedManager struct {
	IsConcurrent bool         `json:"isConcurrent"`
	Personal     AmedPersonal `json:"personal"`
}
type AmedStaff struct {
	BelongTo          string `json:"belongTo"`
	Post              string `json:"post"`
	Name              string `json:"name"`
	ERad              string `json:"e-Rad"`
	CanPublished      bool   `json:"canPublished"`
	PostType          string `json:"postType"`
	FinancialResource string `json:"financialResource"`
	EmploymentStatus  string `json:"employmentStatus"`
	Roles             string `json:"roles"`
	Remarks           string `json:"remarks"`
}
type AmedResearchers struct {
	NumberOfPeople int         `json:"numberOfPeople"`
	Manager        AmedManager `json:"manager"`
	Staff          []AmedStaff `json:"staff"`
}
