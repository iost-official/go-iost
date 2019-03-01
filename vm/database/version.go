package database

import "strconv"

func getHeight(db database) int64 {
	str := db.Get("reserve.height")
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 9223372036854775807
	}
	return i
}
