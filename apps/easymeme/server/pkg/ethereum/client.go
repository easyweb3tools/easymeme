package ethereum

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	PancakeFactoryV2 = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"
	PancakeRouterV2  = "0x10ED43C718714eb63d5aA57B78B54704E256024E"
	WBNB             = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"
)

var PairCreatedTopic = common.HexToHash("0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9")

type Client struct {
	http *ethclient.Client
	ws   *ethclient.Client
}

func NewClient(httpURL, wsURL string) (*Client, error) {
	httpClient, err := ethclient.Dial(httpURL)
	if err != nil {
		return nil, err
	}

	var wsClient *ethclient.Client
	if wsURL != "" {
		wsClient, err = ethclient.Dial(wsURL)
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		http: httpClient,
		ws:   wsClient,
	}, nil
}

func (c *Client) SubscribePairCreated(ctx context.Context) (chan types.Log, ethereum.Subscription, error) {
	if c.ws == nil {
		return nil, nil, ethereum.NotFound
	}
	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(PancakeFactoryV2)},
		Topics:    [][]common.Hash{{PairCreatedTopic}},
	}

	logs := make(chan types.Log)
	sub, err := c.ws.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return nil, nil, err
	}

	return logs, sub, nil
}

func (c *Client) LatestBlockNumber(ctx context.Context) (uint64, error) {
	return c.http.BlockNumber(ctx)
}

func (c *Client) GetPairCreatedLogs(ctx context.Context, fromBlock, toBlock uint64) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{common.HexToAddress(PancakeFactoryV2)},
		Topics:    [][]common.Hash{{PairCreatedTopic}},
	}

	return c.http.FilterLogs(ctx, query)
}

func (c *Client) GetTokenInfo(ctx context.Context, tokenAddr common.Address) (name, symbol string, decimals uint8, err error) {
	nameData, err := c.http.CallContract(ctx, ethereum.CallMsg{
		To:   &tokenAddr,
		Data: common.Hex2Bytes("06fdde03"),
	}, nil)
	if err == nil && len(nameData) > 0 {
		name = parseString(nameData)
	}

	symbolData, err := c.http.CallContract(ctx, ethereum.CallMsg{
		To:   &tokenAddr,
		Data: common.Hex2Bytes("95d89b41"),
	}, nil)
	if err == nil && len(symbolData) > 0 {
		symbol = parseString(symbolData)
	}

	decimalsData, err := c.http.CallContract(ctx, ethereum.CallMsg{
		To:   &tokenAddr,
		Data: common.Hex2Bytes("313ce567"),
	}, nil)
	if err == nil && len(decimalsData) > 0 {
		decimals = uint8(new(big.Int).SetBytes(decimalsData).Uint64())
	} else {
		decimals = 18
	}

	return name, symbol, decimals, nil
}

func (c *Client) GetPairReserves(ctx context.Context, pairAddr common.Address) (reserve0, reserve1 *big.Int, err error) {
	data, err := c.http.CallContract(ctx, ethereum.CallMsg{
		To:   &pairAddr,
		Data: common.Hex2Bytes("0902f1ac"),
	}, nil)
	if err != nil {
		return nil, nil, err
	}

	if len(data) >= 64 {
		reserve0 = new(big.Int).SetBytes(data[0:32])
		reserve1 = new(big.Int).SetBytes(data[32:64])
	}
	return reserve0, reserve1, nil
}

func (c *Client) SimulateSell(ctx context.Context, tokenAddr common.Address, amount *big.Int) error {
	router := common.HexToAddress(PancakeRouterV2)
	wbnb := common.HexToAddress(WBNB)

	routerABI := `[{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForETHSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
	parsed, err := abi.JSON(strings.NewReader(routerABI))
	if err != nil {
		return err
	}

	deadline := big.NewInt(time.Now().Add(2 * time.Minute).Unix())
	to := common.HexToAddress("0x000000000000000000000000000000000000dEaD")

	data, err := parsed.Pack(
		"swapExactTokensForETHSupportingFeeOnTransferTokens",
		amount,
		big.NewInt(0),
		[]common.Address{tokenAddr, wbnb},
		to,
		deadline,
	)
	if err != nil {
		return err
	}

	_, err = c.http.CallContract(ctx, ethereum.CallMsg{
		To:   &router,
		Data: data,
	}, nil)

	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "insufficient allowance") ||
			strings.Contains(msg, "transfer amount exceeds balance") ||
			strings.Contains(msg, "insufficient balance") ||
			strings.Contains(msg, "erc20: transfer amount exceeds balance") {
			return nil
		}
	}

	return err
}

func (c *Client) Close() {
	c.http.Close()
	if c.ws != nil {
		c.ws.Close()
	}
}

func (c *Client) GetBalance(ctx context.Context, addr common.Address) (*big.Int, error) {
	return c.http.BalanceAt(ctx, addr, nil)
}

func (c *Client) Receipt(ctx context.Context, hash common.Hash) (*types.Receipt, error) {
	return c.http.TransactionReceipt(ctx, hash)
}

func (c *Client) TokenBalance(ctx context.Context, tokenAddr, owner common.Address) (*big.Int, error) {
	erc20ABI := `[{"constant":true,"inputs":[{"name":"owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, err
	}
	data, err := parsed.Pack("balanceOf", owner)
	if err != nil {
		return nil, err
	}
	res, err := c.http.CallContract(ctx, ethereum.CallMsg{
		To:   &tokenAddr,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(res), nil
}

func (c *Client) ApproveToken(ctx context.Context, pk *ecdsa.PrivateKey, tokenAddr, spender common.Address, amount *big.Int) (common.Hash, error) {
	erc20ABI := `[{"constant":false,"inputs":[{"name":"spender","type":"address"},{"name":"value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]`
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return common.Hash{}, err
	}
	data, err := parsed.Pack("approve", spender, amount)
	if err != nil {
		return common.Hash{}, err
	}
	return c.sendTx(ctx, pk, tokenAddr, big.NewInt(0), data)
}

func (c *Client) SwapExactETHForTokens(ctx context.Context, pk *ecdsa.PrivateKey, tokenAddr common.Address, amountInWei, amountOutMin *big.Int) (common.Hash, error) {
	router := common.HexToAddress(PancakeRouterV2)
	wbnb := common.HexToAddress(WBNB)
	routerABI := `[{"inputs":[{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactETHForTokensSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"payable","type":"function"}]`
	parsed, err := abi.JSON(strings.NewReader(routerABI))
	if err != nil {
		return common.Hash{}, err
	}
	deadline := big.NewInt(time.Now().Add(2 * time.Minute).Unix())
	to := cryptoPubkeyAddress(pk)
	data, err := parsed.Pack(
		"swapExactETHForTokensSupportingFeeOnTransferTokens",
		amountOutMin,
		[]common.Address{wbnb, tokenAddr},
		to,
		deadline,
	)
	if err != nil {
		return common.Hash{}, err
	}
	return c.sendTx(ctx, pk, router, amountInWei, data)
}

func (c *Client) SwapExactTokensForETH(ctx context.Context, pk *ecdsa.PrivateKey, tokenAddr common.Address, amountIn, amountOutMin *big.Int) (common.Hash, error) {
	router := common.HexToAddress(PancakeRouterV2)
	wbnb := common.HexToAddress(WBNB)
	routerABI := `[{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForETHSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
	parsed, err := abi.JSON(strings.NewReader(routerABI))
	if err != nil {
		return common.Hash{}, err
	}
	deadline := big.NewInt(time.Now().Add(2 * time.Minute).Unix())
	to := cryptoPubkeyAddress(pk)
	data, err := parsed.Pack(
		"swapExactTokensForETHSupportingFeeOnTransferTokens",
		amountIn,
		amountOutMin,
		[]common.Address{tokenAddr, wbnb},
		to,
		deadline,
	)
	if err != nil {
		return common.Hash{}, err
	}
	return c.sendTx(ctx, pk, router, big.NewInt(0), data)
}

func (c *Client) sendTx(ctx context.Context, pk *ecdsa.PrivateKey, to common.Address, value *big.Int, data []byte) (common.Hash, error) {
	from := cryptoPubkeyAddress(pk)
	nonce, err := c.http.PendingNonceAt(ctx, from)
	if err != nil {
		return common.Hash{}, err
	}
	gasPrice, err := c.http.SuggestGasPrice(ctx)
	if err != nil {
		return common.Hash{}, err
	}
	callMsg := ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: value,
		Data:  data,
	}
	gasLimit, err := c.http.EstimateGas(ctx, callMsg)
	if err != nil {
		gasLimit = 400000
	}
	tx := types.NewTransaction(nonce, to, value, gasLimit, gasPrice, data)
	chainID, err := c.http.NetworkID(ctx)
	if err != nil {
		return common.Hash{}, err
	}
	signed, err := types.SignTx(tx, types.NewEIP155Signer(chainID), pk)
	if err != nil {
		return common.Hash{}, err
	}
	if err := c.http.SendTransaction(ctx, signed); err != nil {
		return common.Hash{}, err
	}
	return signed.Hash(), nil
}

func cryptoPubkeyAddress(pk *ecdsa.PrivateKey) common.Address {
	return crypto.PubkeyToAddress(pk.PublicKey)
}

func parseString(data []byte) string {
	if len(data) < 64 {
		return ""
	}
	offset := new(big.Int).SetBytes(data[0:32]).Uint64()
	if offset+32 > uint64(len(data)) {
		return ""
	}
	length := new(big.Int).SetBytes(data[offset : offset+32]).Uint64()
	if offset+32+length > uint64(len(data)) {
		return ""
	}
	return string(data[offset+32 : offset+32+length])
}
