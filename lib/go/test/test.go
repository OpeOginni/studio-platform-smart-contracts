package test

import (
	"io/ioutil"
	"testing"

	"github.com/onflow/cadence"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-emulator/types"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	sdktemplates "github.com/onflow/flow-go-sdk/templates"
	"github.com/onflow/flow-go-sdk/test"
	nftcontracts "github.com/onflow/flow-nft/lib/go/contracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/onflow/flow-go-sdk"
)

const (
	flowTokenName         = "FlowToken"
	nonFungibleTokenName  = "NonFungibleToken"
	defaultAccountFunding = "1000.0"
)

var (
	ftAddress        = flow.HexToAddress("ee82856bf20e2aa6")
	flowTokenAddress = flow.HexToAddress("0ae53cb6e3f42a79")
)

type Contracts struct {
	NFTAddress          flow.Address
	MetadataViewAddress flow.Address
	DapperSportAddress  flow.Address
	DapperSportSigner   crypto.Signer
}

func deployNFTContract(t *testing.T, b *emulator.Blockchain) flow.Address {
	nftCode := nftcontracts.NonFungibleToken()
	nftAddress, err := b.CreateAccount(nil,
		[]sdktemplates.Contract{
			{
				Name:   nonFungibleTokenName,
				Source: string(nftCode),
			},
		},
	)
	require.NoError(t, err)

	_, err = b.CommitBlock()
	require.NoError(t, err)

	return nftAddress
}

func DapperSportDeployContracts(t *testing.T, b *emulator.Blockchain) Contracts {
	accountKeys := test.AccountKeyGenerator()

	nftAddress := deployNFTContract(t, b)
	metadataCode := LoadMetadataViews(ftAddress, nftAddress)
	metadataViewsAddr, err := b.CreateAccount(nil, []sdktemplates.Contract{
		{
			Name:   "MetadataViews",
			Source: string(metadataCode),
		},
	})
	if !assert.NoError(t, err) {
		t.Log(err.Error())
	}
	_, err = b.CommitBlock()
	assert.NoError(t, err)

	DapperSportAccountKey, DapperSportSigner := accountKeys.NewWithSigner()
	DapperSportCode := LoadDapperSport(nftAddress, metadataViewsAddr)

	DapperSportAddress, err := b.CreateAccount(
		[]*flow.AccountKey{DapperSportAccountKey},
		nil,
	)
	require.NoError(t, err)

	fundAccount(t, b, DapperSportAddress, defaultAccountFunding)

	tx1 := sdktemplates.AddAccountContract(
		DapperSportAddress,
		sdktemplates.Contract{
			Name:   "DapperSport",
			Source: string(DapperSportCode),
		},
	)

	tx1.
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address)

	signer, err := b.ServiceKey().Signer()
	assert.NoError(t, err)
	
	signAndSubmit(
		t, b, tx1,
		[]flow.Address{b.ServiceKey().Address, DapperSportAddress},
		[]crypto.Signer{signer, DapperSportSigner},
		false,
	)

	_, err = b.CommitBlock()
	require.NoError(t, err)

	return Contracts{
		nftAddress,
		metadataViewsAddr,
		DapperSportAddress,
		DapperSportSigner,
	}
}

// newEmulator returns a emulator object for testing
func newEmulator() *emulator.Blockchain {
	b, err := emulator.NewBlockchain()
	if err != nil {
		panic(err)
	}
	return b
}

// signAndSubmit signs a transaction with an array of signers and adds their signatures to the transaction
// Then submits the transaction to the emulator. If the private keys don't match up with the addresses,
// the transaction will not succeed.
// shouldRevert parameter indicates whether the transaction should fail or not
// This function asserts the correct result and commits the block if it passed
func signAndSubmit(
	t *testing.T,
	b *emulator.Blockchain,
	tx *flow.Transaction,
	signerAddresses []flow.Address,
	signers []crypto.Signer,
	shouldRevert bool,
) *types.TransactionResult{
	// sign transaction with each signer
	for i := len(signerAddresses) - 1; i >= 0; i-- {
		signerAddress := signerAddresses[i]
		signer := signers[i]

		if i == 0 {
			err := tx.SignEnvelope(signerAddress, 0, signer)
			assert.NoError(t, err)
		} else {
			err := tx.SignPayload(signerAddress, 0, signer)
			assert.NoError(t, err)
		}
	}

	return submit(t, b, tx, shouldRevert)
}

// submit submits a transaction and checks
// if it fails or not
func submit(
	t *testing.T,
	b *emulator.Blockchain,
	tx *flow.Transaction,
	shouldRevert bool,
) *types.TransactionResult{
	// submit the signed transaction
	err := b.AddTransaction(*tx)
	require.NoError(t, err)

	result, err := b.ExecuteNextTransaction()
	require.NoError(t, err)

	if shouldRevert {
		assert.True(t, result.Reverted())
	} else {
		if !assert.True(t, result.Succeeded()) {
			t.Log(result.Error.Error())
		}
	}

	_, err = b.CommitBlock()
	assert.NoError(t, err)
	return result
}

// executeScriptAndCheck executes a script and checks to make sure that it succeeded.
func executeScriptAndCheck(t *testing.T, b *emulator.Blockchain, script []byte, arguments [][]byte) cadence.Value {
	result, err := b.ExecuteScript(script, arguments)
	require.NoError(t, err)
	if !assert.True(t, result.Succeeded()) {
		t.Log(result.Error.Error())
	}
	return result.Value
}

// readFile reads a file from the file system
// and returns its contents
func readFile(path string) []byte {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return contents
}

// cadenceUFix64 returns a UFix64 value
func cadenceUFix64(value string) cadence.Value {
	newValue, err := cadence.NewUFix64(value)
	if err != nil {
		panic(err)
	}

	return newValue
}

// Simple error-handling wrapper for Flow account creation.
func createAccount(t *testing.T, b *emulator.Blockchain) (sdk.Address, crypto.Signer) {
	accountKeys := test.AccountKeyGenerator()
	accountKey, signer := accountKeys.NewWithSigner()

	address, err := b.CreateAccount([]*sdk.AccountKey{accountKey}, nil)
	require.NoError(t, err)

	return address, signer
}

func setupDapperSport(
	t *testing.T,
	b *emulator.Blockchain,
	userAddress sdk.Address,
	userSigner crypto.Signer,
	contracts Contracts,
) {
	tx := flow.NewTransaction().
		SetScript(loadDapperSportSetupAccountTransaction(contracts)).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)

	signer, err := b.ServiceKey().Signer()
	assert.NoError(t, err)

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{signer, userSigner},
		false,
	)
}

func setupAccount(
	t *testing.T,
	b *emulator.Blockchain,
	address flow.Address,
	signer crypto.Signer,
	contracts Contracts,
) (sdk.Address, crypto.Signer) {
	setupDapperSport(t, b, address, signer, contracts)
	fundAccount(t, b, address, defaultAccountFunding)

	return address, signer
}

func metadataDict(dict map[string]string) cadence.Dictionary {
	pairs := []cadence.KeyValuePair{}

	for key, value := range dict {
		cadenceKey, _ := cadence.NewString(key)
		cadenceValue, _ := cadence.NewString(value)
		pairs = append(pairs, cadence.KeyValuePair{Key: cadenceKey, Value: cadenceValue})
	}

	return cadence.NewDictionary(pairs)
}