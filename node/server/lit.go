package server

import (
	"blocksui-node/abi"
	"blocksui-node/account"
	"blocksui-node/config"
	"blocksui-node/contracts"
	"blocksui-node/ipfs"
	"blocksui-node/lit"
	"bytes"
	"fmt"

	"github.com/gin-gonic/gin"
	goIpfs "github.com/ipfs/go-ipfs-api"
)

func LitEncrypt(c *config.Config, a *account.Account) gin.HandlerFunc {
	return func(r *gin.Context) {
		ipfsClient := r.MustGet("ipfs").(*goIpfs.Shell)
		plaintext := r.MustGet("block").([]byte)
		metadata := r.MustGet("metadata").(*BlockMeta)

		symmetricKey := lit.Prng(32)
		ciphertext := lit.AesEncrypt(symmetricKey, plaintext)

		cid, err := ipfsClient.Add(bytes.NewBuffer(ciphertext))
		if err != nil {
			r.AbortWithError(500, err)
			return
		}

		b32Cid := ipfs.CidToBytes32(cid)

		contract, ok := contracts.GetContract("BUIBlockNFT")
		if !ok {
			r.AbortWithError(500, fmt.Errorf("Contract not found"))
			return
		}

		method := contract.Abi.GetMethod("verifyOwner")
		if method == nil {
			r.AbortWithError(404, fmt.Errorf("Method not found"))
			return
		}

		authConditions := []lit.EvmContractCondition{
			lit.EvmContractCondition{
				ContractAddress: contract.Address.String(),
				Chain:           c.Chain(), // TODO: make this configurable
				FunctionAbi:     abi.MethodToMember(method),
				FunctionName:    "verifyOwner",
				FunctionParams: []string{
					b32Cid,
					":userAddress",
				},
				ReturnValueTest: lit.ReturnValueTest{
					Key:        "",
					Comparator: "=",
					Value:      "true",
				},
			},
		}

		// TODO: need a chain <> name map
		authSig, err := a.Siwe("80001", "")
		if err != nil {
			r.AbortWithError(500, err)
			return
		}

		litClient := lit.New(c)

		encryptedKey, err := litClient.SaveEncryptionKey(
			symmetricKey,
			*authSig,
			authConditions,
			c.Chain(),
		)
		if err != nil {
			r.AbortWithError(500, err)
			return
		}

		metadata.BUIProps = BUIProps{
			Cid:          cid,
			EncryptedKey: encryptedKey,
		}

		r.Set("metadata", metadata)
		r.Set("cid", b32Cid)

		r.Next()
	}
}
