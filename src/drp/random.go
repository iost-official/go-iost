package drp

import {
	"bytes"
	"errors"
	"math/big"

	"golang.org/x/crypto/sha3"
}

// globals: block, msg

type Address [AddressLength]byte

type Participant struct {
	secret *big.Int
	commitment [32]byte
	reward *big.Int
	revealed bool
	rewarded bool
}

type Consumer struct {
	address Address
	bountyPot *big.Int
}

type Campaign struct {
	blockNumber uint32
	deposit *big.Int
	commitBalkline uint16
	commitDeadline uint16
	
	random *big.Int
	settled bool
	bountyPot *big.Int
	numCommits uint32
	numReveals uint32

	consumers map[Address]*Consumer
	participants map[Address]*Participant
}

// TODO: Persist this to the blockchain
[]Campaign campaigns

func newCampaign(blockNumber uint32, deposit *big.Int, commitBalkline uint16, commitDeadline uint16) *big.Int {
	if block.number >= blockNumber ||
			commitBalkline <= 0 ||
			commitDeadline <= 0 ||
			commitDeadline >= commitBalkline ||
			block.number >= blockNumber - commitBalkline ||
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
	campaign = campaigns[campaignID]
	consumer = campaign.consumers[msg.sender]
	if block.number > campaign.blockNumber - campaign.commitDeadline ||
	    consumer.address == 0 {
		return errors.New("")
	}
	campaign.bountyBot += msg.value;
	campaign.consumers[msg.sender] = &Consumer{address: msg.sender, bountyPot: msg.value}
	// Log follow
	return true; 
}

func commitCampaign(campaignID *big.Int, hash [32]byte) {
	campaign = campaigns[campaignID]
	if msg.value != deposit ||
			block.number < blockNumber - commitBalkline ||
			block.number > blockNumber - commitDeadline ||
			!bytes.Equal(campaign.participants[msg.sender].commitment, make([]byte, 32)) {
	  return errors.New("")			
	}
  campaign.participants[msg.sender] = &Participant{secret: 0, commitment: hash, reward: 0, revealed: false, rewarded: false}
	campaign.numCommits++;
	// Log commit
}

func revealCampaign(campaignID *big.Int, secret *big.Int, campaign *Campaign, participant *Participant) {
	
}

func returnReward(share *big.Int, campaign Campaign, participant Participant) {	
}

func refundBounty(campaignID *big.Int) {
	campaign := campaigns[campaignID];
	if block.number < campaign.blockNumber ||
			(campaign.numCommits == campaign.numReveals && campaign.numCommits != 0) ||
	   campaign.consumers[msg.sender].address != msg.sender {
	  return errors.New("")
	}
}