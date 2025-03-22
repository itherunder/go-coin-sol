package jupiter

import (
	"encoding/base64"
	"strconv"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	type_ "github.com/itherunder/go-coin-sol/type"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
	"github.com/pkg/errors"
)

type SwapInfoType struct {
	AmmKey     string `json:"ammKey"`
	Label      string `json:"label"`
	InputMint  string `json:"inputMint"`
	OutputMint string `json:"outputMint"`
	InAmount   string `json:"inAmount"`
	OutAmount  string `json:"outAmount"`
	FeeAmount  string `json:"feeAmount"`
	FeeMint    string `json:"feeMint"`
}

type RoutePlanType struct {
	SwapInfo SwapInfoType `json:"swapInfo"`
	Percent  uint64       `json:"percent"`
}

type QuoteType struct {
	InputMint            string `json:"inputMint"`
	InAmount             string `json:"inAmount"`
	OutputMint           string `json:"outputMint"`
	OutAmount            string `json:"outAmount"`
	OtherAmountThreshold string `json:"otherAmountThreshold"`
	SwapMode             string `json:"swapMode"`
	SlippageBps          uint64 `json:"slippageBps"`
	// PlatformFee          string `json:"platformFee,omitempty"`
	PriceImpactPct string          `json:"priceImpactPct"`
	RoutePlan      []RoutePlanType `json:"routePlan"`
	// ContextSlot      uint64  `json:"contextSlot"`
	// TimeTaken        float64 `json:"timeTaken"`
	// SwapUsdValue     string  `json:"swapUsdValue"`
	// SimplerRouteUsed bool    `json:"simplerRouteUsed"`
}

type SwapInstructionsParamsType struct {
	QuoteResponse    *QuoteType `json:"quoteResponse"`
	UserPublicKey    string     `json:"userPublicKey"`
	WrapAndUnwrapSol bool       `json:"wrapAndUnwrapSol"`
}

type InstructionType struct {
	ProgramId string `json:"programId"`
	Accounts_ []struct {
		Pubkey     string `json:"pubkey"`
		IsSigner   bool   `json:"isSigner"`
		IsWritable bool   `json:"isWritable"`
	} `json:"accounts"`
	Data_ string `json:"data"`
}

func (t *InstructionType) ProgramID() solana.PublicKey {
	return solana.MustPublicKeyFromBase58(t.ProgramId)
}
func (t *InstructionType) Accounts() []*solana.AccountMeta {
	accounts := make([]*solana.AccountMeta, 0)
	for _, account := range t.Accounts_ {
		accounts = append(accounts, solana.NewAccountMeta(
			solana.MustPublicKeyFromBase58(account.Pubkey),
			account.IsWritable,
			account.IsSigner,
		))
	}
	return accounts
}
func (t *InstructionType) Data() ([]byte, error) {
	return base64.StdEncoding.DecodeString(t.Data_)
}

type SwapInstructionsResultType struct {
	ComputeBudgetInstructions   []*InstructionType `json:"computeBudgetInstructions"`
	SetupInstructions           []*InstructionType `json:"setupInstructions"`
	SwapInstruction             *InstructionType   `json:"swapInstruction"`
	CleanupInstruction          *InstructionType   `json:"cleanupInstruction"`
	OtherInstructions           []*InstructionType `json:"otherInstructions"`
	AddressLookupTableAddresses []string           `json:"addressLookupTableAddresses"`
	PrioritizationFeeLamports   uint64             `json:"prioritizationFeeLamports"`
	ComputeUnitLimit            uint64             `json:"computeUnitLimit"`
	PrioritizationType          struct {
		ComputeBudget struct {
			MicroLamports          uint64 `json:"microLamports"`
			EstimatedMicroLamports uint64 `json:"estimatedMicroLamports"`
		} `json:"computeBudget"`
	} `json:"prioritizationType"`
}

func GetQuote(
	logger i_logger.ILogger,
	swapType type_.SwapType,
	tokenAddress solana.PublicKey,
	tokenAmountWithDecimals uint64,
	slippage uint64,
) (*QuoteType, error) {
	quoteQueries := map[string]string{
		"amount":      strconv.FormatUint(tokenAmountWithDecimals, 10),
		"slippageBps": strconv.FormatUint(slippage, 10),
	}
	if swapType == type_.SwapType_Buy {
		quoteQueries["inputMint"] = solana.SolMint.String()
		quoteQueries["outputMint"] = tokenAddress.String()
		quoteQueries["swapMode"] = "ExactOut"
	} else {
		quoteQueries["inputMint"] = tokenAddress.String()
		quoteQueries["outputMint"] = solana.SolMint.String()
		quoteQueries["swapMode"] = "ExactIn"
	}
	var quoteResponse QuoteType
	_, _, err := go_http.NewHttpRequester(go_http.WithLogger(logger)).GetForStruct(
		&go_http.RequestParams{
			Url:     "https://quote-api.jup.ag/v6/quote",
			Queries: quoteQueries,
		},
		&quoteResponse,
	)
	if err != nil {
		return nil, err
	}

	return &quoteResponse, nil
}

func GetSwapInstructions(
	logger i_logger.ILogger,
	userAddress solana.PublicKey,
	tokenAddress solana.PublicKey,
	quoteResponse *QuoteType,
	isClose bool,
) ([]solana.Instruction, error) {
	instructions := make([]solana.Instruction, 0)

	var swapInstructionsResult SwapInstructionsResultType
	_, _, err := go_http.NewHttpRequester(go_http.WithLogger(logger)).PostForStruct(
		&go_http.RequestParams{
			Url: "https://quote-api.jup.ag/v6/swap-instructions",
			Params: SwapInstructionsParamsType{
				QuoteResponse:    quoteResponse,
				UserPublicKey:    userAddress.String(),
				WrapAndUnwrapSol: true,
			},
		},
		&swapInstructionsResult,
	)
	if err != nil {
		return nil, err
	}

	for _, inst := range swapInstructionsResult.SetupInstructions {
		instructions = append(instructions, inst)
	}

	instructions = append(instructions, swapInstructionsResult.SwapInstruction, swapInstructionsResult.CleanupInstruction)

	if isClose {
		userTokenAssociatedAccount, _, err := solana.FindAssociatedTokenAddress(
			userAddress,
			tokenAddress,
		)
		if err != nil {
			return nil, errors.Wrapf(err, "<userAddress: %s> <tokenAddress: %s>", userAddress, tokenAddress)
		}

		instructions = append(
			instructions,
			token.NewCloseAccountInstruction(
				userTokenAssociatedAccount,
				userAddress,
				userAddress,
				nil,
			).Build(),
		)
	}

	return instructions, nil
}
