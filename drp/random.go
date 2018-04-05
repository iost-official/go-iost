package drp

import (
	"bytes"
	"errors"
	"math/big"

	"golang.org/x/crypto/sha3"
)

// globals: block, msg

type Address [32]byte

type Participant struct {
	secret     *big.Int
	commitment [32]byte
	reward     *big.Int
	revealed   bool
	rewarded   bool
}

type Consumer struct {
	address   Address
	bountyPot *big.Int
}

type Campaign struct {
	blockNumber    uint32
	deposit        *big.Int
	commitBalkline uint16
	commitDeadline uint16

	random     *big.Int
	settled    bool
	bountyPot  *big.Int
	numCommits uint32
	numReveals uint32

	consumers    map[Address]*Consumer
	participants map[Address]*Participant
}

// TODO: Persist this to LevelDB
var campaigns []Campaign

func newCampaign(blockNumber uint32, deposit *big.Int, commitBalkline uint16, commitDeadline uint16) *big.Int {
	if block.number >= blockNumber ||
		commitBalkline <= 0 ||
		commitDeadline <= 0 ||
		commitDeadline >= commitBalkline ||
		block.number >= blockNumber-commitBalkline ||
		deposit <= 0 {
		return errors.New("")
	}
	campaignID := big.NewInt(len(campaigns) + 1)
	campaign := campaigns[campaignID]
	campaign.blockNumber = blockNumber
	campaign.deposit = deposit
	campaign.commitBalkline = commitBalkline
	campaign.commitDeadline = commitDeadline
	campaign.bountyPot = msg.value
	campaign.consumers[msg.sender] = &Consumer{address: msg.sender, bountyPot: msg.value}
	// Log campaign added
	return campaignID
}

func followCampaign(campaignID *big.Int) bool {
	campaign := campaigns[campaignID]
	consumer := campaign.consumers[msg.sender]
	if block.number > campaign.blockNumber-campaign.commitDeadline ||
		consumer.address == 0 {
		return errors.New("")
	}
	campaign.bountyBot += msg.value
	campaign.consumers[msg.sender] = &Consumer{address: msg.sender, bountyPot: msg.value}
	// Log follow
	return true
}

func commitCampaign(campaignID *big.Int, hash [32]byte) {
	campaign := campaigns[campaignID]
	if msg.value != deposit ||
		block.number < blockNumber-commitBalkline ||
		block.number > blockNumber-commitDeadline ||
		!bytes.Equal(campaign.participants[msg.sender].commitment, make([]byte, 32)) {
		return errors.New("")
	}
	campaign.participants[msg.sender] = &Participant{secret: 0, commitment: hash, reward: 0, revealed: false, rewarded: false}
	campaign.numCommits++
	// Log commit
}

func revealCampaign(campaignID *big.Int, secret *big.Int, campaign *Campaign, participant *Participant) {
	if block.number <= campaign.blockNumber-campaign.commitDeadline ||
		block.number >= campaign.blockNumber ||
		sha3.NewKeccak256().Sum(secret) != participant.commitment ||
		participnt.revealed {
		return errors.New("")
	}
	participant.secret = secret
	participant.revealed = true
	campaign.numReveals++
	campaign.random ^= participant.secret
	// Log reveal
}

func getRandom(campaignID *big.Int) *big.Int {
	campaign := campaigns[campaignID]
	if campaign.numReveals == campaign.numCommits {
		campaign.settled = true
		return campaign.random
	}
	return errors.New("")
}

func getMyBounty(campaignID *big.Int) {
	campaign := campaigns[campaignID]
	participant := campaign.participants[msg.sender]
	if block.number < campaign.blockNumber ||
		participant.rewarded {
		return errors.New("")
	}
	if campaign.numReveals > 0 {
		if participant.revealed {
			share := calculateShare(campaign)
			returnReward(share, campaign, participant)
		}
	} else {
		returnReward(0, campaign, participant)
	}
}

func calculateShare(campaign Campaign) *big.Int {
	if campaign.numCommits > campaign.numReveals {
		share := fines(campaign) / campaign.numReveals
	} else {
		share := campaign.bountyPot / campaign.numReveals
	}
}

func fines(campaign Campaign) *big.Int {
	return (campaign.numCommits - campaign.numReveals) * campaign.deposit
}

func returnReward(share *big.Int, campaign Campaign, participant Participant) {
	participant.reward = share
	participant.rewarted = true
	if !msg.sender.send(share + campaign.deposit) {
		participant.reward = 0
		participant.rewarded = false
	}
}

func refundBounty(campaignID *big.Int) {
	campaign := campaigns[campaignID]
	if block.number < campaign.blockNumber ||
		(campaign.numCommits == campaign.numReveals && campaign.numCommits != 0) ||
		campaign.consumers[msg.sender].address != msg.sender {
		return errors.New("")
	}
	bountyPot := campaign.consumers[msg.sender].bountyPo
	campaign.consumers[msg.sender].bountyPot = 0
	if !msg.sender.send(bountyPot) {
		campaign.consumers[msg.sender].bountyPot = bountyPot
	}
}
