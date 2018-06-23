--- main bet lucky number for walletb
-- @gas_limit 1000000
-- @gas_price 0.0001
-- @param_cnt 0
-- @return_cnt 0
-- @publisher walletb
function main()
	nonce = 0
	tx = "4TKrbjmyoY77F2WdpVn3cxqGWipkMGzGKQ7YcAQ45LtJ"
	-- tx = "main"
	b = "2538yUDuKTLaXqCTFS1tfVmMEL4dVnzLDWChoMdoxgCa4"
	-- b = "walletb"
	ok, r = Call(tx, "QueryLastLuckyBlock")
	Log(string.format("last lucky block %s", tostring(ok)))
	Log(string.format("last lucky block r = %s", tostring(r)))
	Assert(ok)
	Assert(r == -1)
    ok, r = Call(tx, "QueryRound")
	Log(string.format("query round %s", tostring(ok)))
	Log(string.format("query round r = %s", tostring(r)))
	Assert(ok)
	Assert(r == 0)

    for i=0,0 do
		ok, r = Call(tx, "Bet", b, 0, 1, nonce)
		Log(string.format("bet %s", tostring(ok)))
		Log(string.format("bet r = %s", tostring(r)))
		Assert(ok)
		Assert(r == 0)
    end
    ok, r = Call(tx, "QueryUserNumber")
	Assert(ok)
	Assert(r == 0)
    ok, r = Call(tx, "QueryTotalCoins")
	Assert(ok)
	Assert(r == 0)
    ok, r = Call(tx, "QueryLastLuckyBlock")
	Assert(ok)
	Assert(r >= 0)
    ok, r = Call(tx, "QueryRound")
	Assert(ok)
	Assert(r == 1)
end--f
