package api

type BusinessActionContext struct {
	Xid           string
	BranchId      string
	ActionName    string
	ActionContext map[string]interface{}
}
