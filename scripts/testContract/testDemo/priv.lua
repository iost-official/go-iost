--- main bet lucky number for walleta
-- @gas_limit 1000000
-- @gas_price 0.0001
-- @param_cnt 0
-- @return_cnt 0
-- @publisher walleta
function main()
	tx = "BDQgPw2o2fTYXXaNCr5Mwq86meFC6dFZC29F5cZsjb3i"
	-- tx = "main"
	a = "gvCQNmkuA6AwdddRMSUg6jr8W7swKWAnhEY3cAthj9bX"
	-- a = "walleta"
	ok, r = Call(tx, "getReward", 0, 1000, 1)
	Log(string.format("get reward %s", tostring(ok)))
	Log(string.format("get reward r = %s", tostring(r)))
	Assert(ok == false)
end--f
