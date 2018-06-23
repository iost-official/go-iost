--- main 一元夺宝
-- snatch treasure with 1 coin !
-- @gas_limit 100000
-- @gas_price 0.001
-- @param_cnt 0
-- @return_cnt 0
-- @publisher walleta
function main()
	Put("max_user_number", 20)
	Put("user_number", 0)
	Put("winner", "")
	Put("claimed", "false")
end--f

--- BuyCoin buy coins
-- buy some coins
-- @param_cnt 2
-- @return_cnt 1
function BuyCoin(account, buyNumber)
	if (buyNumber <= 0)
	then
	    return "buy number should be more than zero"
	end

	maxUserNumber = Get("max_user_number")
    number = Get("user_number")
	if (number >= maxUserNumber or number + buyNumber > maxUserNumber)
	then
	    return string.format("max user number exceed, only %d coins left", maxUserNumber - number)
	end

	-- print(string.format("deposit account = %s, number = %d", account, buyNumber))
	Assert(Deposit(account, buyNumber) == 1)

	win = false
	for i = 0, buyNumber - 1, 1 do
	    win = win or winAfterBuyOne(number)
	    number = number + 1
	end
	Put("user_number", number)

	if (win)
	then
	    Put("winner", account)
	end

	return 0
end--f

--- winAfterBuyOne win after buy one
-- @param_cnt 1
-- @return_cnt 1
function winAfterBuyOne(number)
	win = Random(1 - 1.0 / (number + 1))
	return win
end--f

--- QueryWinner query winner
-- @param_cnt 0
-- @return_cnt 1
function QueryWinner()
	return Get("winner")
end--f

--- QueryClaimed query claimed
-- @param_cnt 0
-- @return_cnt 1
function QueryClaimed()
	return Get("claimed")
end--f

--- QueryUserNumber query user number 
-- @param_cnt 0
-- @return_cnt 1
function QueryUserNumber()
	return Get("user_number")
end--f

--- QueryMaxUserNumber query max user number 
-- @param_cnt 0
-- @return_cnt 1
function QueryMaxUserNumber()
	return Get("max_user_number")
end--f

--- Claim claim prize
-- @param_cnt 0
-- @return_cnt 1
function Claim()
	claimed = Get("claimed")
	if (claimed == "true")
	then
		return "price has been claimed"
	end
	number = Get("user_number")
	maxUserNumber = Get("max_user_number")
	if (number < maxUserNumber)
	then
		return string.format("game not end yet! user_number = %d, max_user_number = %d", number, maxUserNumber)
	end
	winner = Get("winner")

	Put("claimed", "true")

	Assert(Withdraw(winner, number) == 1)
	return 0
end--f
