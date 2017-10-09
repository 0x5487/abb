package identity

import (
	"context"
)

func ValidateModule(ctx context.Context, code string) bool {
	// claim, found := FromContext(ctx)
	// if !found {
	// 	appErr := app.AppError{ErrorCode: "invalid_input", Message: "claim was invalid"}
	// 	panic(appErr)
	// }
	// moduleID, exist := _auditSvc.ModulesDic[code]
	// if !exist {
	// 	return false
	// }
	// for _, m := range claim.Modules {

	// 	mID, err := strconv.Atoi(m)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	if mID == moduleID {
	// 		return true
	// 	}
	// }
	// return false
	return false
}
