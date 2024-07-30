import { vultisig as KeysignMessage } from "../protos/vultisig/keysign/v1/keysign_message";
import { vultisig as BlockchainSpecific } from "../protos/vultisig/keysign/v1/blockchain_specific";
import { vultisig as CoinType } from "../protos/vultisig/keysign/v1/coin";

export type KeysignPayloadType = KeysignMessage.keysign.v1.KeysignPayload;

export const { KeysignPayload } = KeysignMessage.keysign.v1;
export const { THORChainSpecific } = BlockchainSpecific.keysign.v1;
export const { Coin } = CoinType.keysign.v1;

export interface KeysignResponse {
  msg: string;
  r: string;
  s: string;
  der_signature: string;
  recovery_id: string;
}
