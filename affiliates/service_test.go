package affiliates_test

import (
	"encoding/hex"
	"github.com/notegio/openrelay/affiliates"
	"github.com/notegio/openrelay/config"
	"github.com/notegio/openrelay/types"
	"gopkg.in/redis.v3"
	"math/big"
	"os"
	"testing"
	"bytes"
	// "time"
)

func getRedisClient(t *testing.T) *redis.Client {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Errorf("Please set the REDIS_URL environment variable")
		return nil
	}
	return redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
}

func TestGetMissingAffiliate(t *testing.T) {
	redisClient := getRedisClient(t)
	if redisClient == nil {
		return
	}
	service := affiliates.NewRedisAffiliateService(redisClient)
	address, _ := hex.DecodeString("1000000000000000000000000000000000000000")
	addressArray := &types.Address{}
	copy(addressArray[:], address[:])
	_, err := service.Get(addressArray)
	if err == nil {
		t.Errorf("Missing affiliate should return error")
		return
	}
}

func TestSetAffiliate(t *testing.T) {
	redisClient := getRedisClient(t)
	if redisClient == nil {
		return
	}
	baseFee := config.NewBaseFee(redisClient)
	if err := baseFee.Set(big.NewInt(10000)); err != nil {
		t.Errorf(err.Error())
		return
	}
	service := affiliates.NewRedisAffiliateService(redisClient)
	affiliate := affiliates.NewAffiliate(new(big.Int), 100)
	address, _ := hex.DecodeString("0000000000000000000000000000000000000000")
	addressArray := &types.Address{}
	copy(addressArray[:], address[:])
	err := service.Set(addressArray, affiliate)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	lookedUpAffiliate, err := service.Get(addressArray)
	fee, _ := baseFee.Get()
	if lookedUpAffiliate.Fee().Cmp(fee) != 0 {
		t.Errorf("Fee should be equal to base fee")
	}
	listedAffiliates, err := service.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(listedAffiliates) != 1 {
		t.Fatalf("Expected exactly one affiliate")
	}
	if !bytes.Equal(listedAffiliates[0][:], addressArray[:]) {
		t.Errorf("Expected '%#x' to equal '%#x'", listedAffiliates[0][:], addressArray[:])
	}
}
