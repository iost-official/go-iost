--- main bet lucky number for walleta
-- @gas_limit 1000000
-- @gas_price 0.0001
-- @param_cnt 0
-- @return_cnt 0
-- @publisher walleta
function main()
	tx = "2RQqYfcfviN349Zx9q4mTRui6CWCHf3Rihkm7VXWmxfo"
	-- tx = "main"
	a = "gvCQNmkuA6AwdddRMSUg6jr8W7swKWAnhEY3cAthj9bX"
	-- a = "walleta"
	ok, r = Call(tx, "QueryLastLuckyBlock")
	Log(string.format("last lucky block %s", tostring(ok)))
	Log(string.format("last lucky block r = %s", tostring(r)))
	Assert(ok)
	Assert(r == -1)

    for i=0,48 do
		ok, r = Call(tx, "Bet", a, i % 10, 1)
		Log(string.format("bet %s", tostring(ok)))
		Log(string.format("bet r = %s", tostring(r)))
		Assert(ok)
		Assert(r == 0)
    end
    ok, r = Call(tx, "QueryUserNumber")
	Assert(ok)
	Assert(r == 49)
    ok, r = Call(tx, "QueryTotalCoins")
	Assert(ok)
	Assert(r == 49)

    for i=0,49 do
		ok, r = Call(tx, "Bet", a, i % 10, 2)
		Assert(ok)
		Assert(r == 0)
    end
    ok, r = Call(tx, "QueryUserNumber")
	Assert(ok)
	Assert(r == 99)
    ok, r = Call(tx, "QueryTotalCoins")
	Assert(ok)
	Assert(r == 149)
end--f
