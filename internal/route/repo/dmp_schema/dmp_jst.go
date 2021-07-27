package dmp_schema

type JstDmpInfo struct {
	Schema         string            `json:"schema"`
	CreateDate     string            `json:"createDate"`
	AmedNumber     string            `json:"amedNumber"`
	Project        JstProject        `json:"project"`
	ForPublication JstForPublication `json:"forPublication"`
	Researches     []JstResearches   `json:"researches"`
	Researchers    JstResearchers    `json:"researchers"`
}
type JstRepresentative struct {
	BelongTo string `json:"belongTo"`
	Post     string `json:"post"`
	Name     string `json:"name"`
}
type JstProject struct {
	FiscalYear     int               `json:"fiscalYear"`
	Title          string            `json:"title"`
	ProblemName    string            `json:"problemName"`
	Representative JstRepresentative `json:"representative"`
}
type JstForPublication struct {
	HasUsed       bool   `json:"hasUsed"`
	UnwriteReason string `json:"unwriteReason"`
}
type JstResearches struct {
	Title            string   `json:"title"`
	Type             []string `json:"type"`
	Description      string   `json:"description"`
	ReleasePolicy    string   `json:"releasePolicy"`
	ConcealReason    string   `json:"concealReason"`
	HasOfferPolicy   bool     `json:"hasOfferPolicy"`
	PolicyName       string   `json:"policyName"`
	RepositoryType   string   `json:"repositoryType"`
	RepositoryName   string   `json:"repositoryName"`
	DataAmount       string   `json:"dataAmount"`
	DataSchema       string   `json:"dataSchema"`
	ProcessingPolicy string   `json:"processingPolicy"`
	IsRegistered     bool     `json:"isRegistered"`
	RegisteredInfo   string   `json:"registeredInfo"`
}
type JstPersonal struct {
	BelongTo string `json:"belongTo"`
	Post     string `json:"post"`
	Name     string `json:"name"`
}
type JstManager struct {
	IsConcurrent bool        `json:"isConcurrent"`
	Personal     JstPersonal `json:"personal"`
}
type JstStaff struct {
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
type JstResearchers struct {
	NumberOfPeople int        `json:"numberOfPeople"`
	Manager        JstManager `json:"manager"`
	Staff          []JstStaff `json:"staff"`
}
