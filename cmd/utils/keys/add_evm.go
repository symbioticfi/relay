package keys

//
//import (
//	"errors"
//	"github.com/spf13/cobra"
//	"middleware-offchain/core/entity"
//)
//
//var addEvmKeyCmd = &cobra.Command{
//	Use:   "add-evm",
//	Short: "Add EVM key",
//	RunE: func(cmd *cobra.Command, args []string) error {
//		if addEvmFlags.PrivateKey == "" && !addEvmFlags.Generate {
//			return errors.New("add --generate if private key omitted")
//		}
// TODO add functions to add with diff namespace
//		kt := entity.KeyTag(uint8(entity.KeyTypeEcdsaSecp256k1)<<4 | (uint8(addEvmFlags.ChainId) & 0x0F))
//		return addKey(kt, addEvmFlags.Generate, addEvmFlags.Force, addEvmFlags.PrivateKey)
//	},
//}
