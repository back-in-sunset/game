package logic

import "fmt"

func formatCommentIDsKey(objID int64, objType int64, rootID int64, sortType int64) string {
	return fmt.Sprintf(prefixCommentIDs, objID, objType, rootID, sortType)
}

func formatCommentObjSortTypeKey(objID int64, objType int64, rootID int64, sortType int64) string {
	return fmt.Sprintf(prefixCommentObjSortType, objID, objType, rootID, sortType)
}
