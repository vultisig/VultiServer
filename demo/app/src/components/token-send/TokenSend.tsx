import React, { useState } from "react";
import { initWasm, TW, WalletCore } from "@trustwallet/wallet-core";
import { v4 as uuidv4 } from "uuid";
import {
  Coin,
  KeysignMessage,
  KeysignPayload,
  KeysignPayloadType,
  KeysignResponse,
  THORChainSpecific,
} from "../../utils/types";
import { createHash } from "crypto-browserify";
import { Buffer } from "buffer";
import {
  generateRandomHex,
  getSignatureWithRecoveryID,
  lzmaCompressData,
} from "../../utils/utils";
import StepOne from "./StepOne";
import StepTwo from "./StepTwo";
import StepThree from "./StepThree";
import {
  broadcastSignedTransaction,
  getAccountInfo,
} from "../../api/thorchain";
import { getDerivedPublicKey } from "../../api/utils/utils";
import { getSignResult, signMessages } from "../../api/vault/vault";

const TokenSend: React.FC = () => {
  const [currentStep, setCurrentStep] = useState<number>(1);
  const [sessionId, setSessionId] = useState<string>("");
  const [qrString, setQrString] = useState<string>("");
  const [uniqueStrings, setUniqueStrings] = useState<string[]>([]);
  const [status, setStatus] = useState<string>("pending");

  const goToStep = (step: number) => {
    setCurrentStep(step);
  };

  function getPreSignedInputData(
    walletCore: WalletCore,
    keysignPayload: KeysignPayloadType
  ): Uint8Array {
    const { AnyAddress } = walletCore;
    const fromAddr = AnyAddress.createWithString(
      keysignPayload.coin.address,
      walletCore.CoinType.thorchain
    );
    const { account_number, sequence } = keysignPayload.thorchain_specific;
    const pubKeyData = Buffer.from(keysignPayload.coin.hex_public_key, "hex");
    const pubKey = walletCore.PublicKey.createWithData(
      pubKeyData,
      walletCore.PublicKeyType.secp256k1
    );
    const toAddress = AnyAddress.createWithString(
      keysignPayload.to_address,
      walletCore.CoinType.thorchain
    );
    const messages = [
      TW.Cosmos.Proto.Message.create({
        thorchainSendMessage: TW.Cosmos.Proto.Message.THORChainSend.create({
          fromAddress: fromAddr.data(),
          amounts: [
            TW.Cosmos.Proto.Amount.create({
              denom: keysignPayload.coin.ticker.toLowerCase(),
              amount: keysignPayload.to_amount.toString(),
            }),
          ],
          toAddress: toAddress.data(),
        }),
      }),
    ];
    const input = TW.Cosmos.Proto.SigningInput.create({
      publicKey: pubKey.data(),
      signingMode: TW.Cosmos.Proto.SigningMode.Protobuf,
      chainId: walletCore.CoinTypeExt.chainId(walletCore.CoinType.thorchain),
      accountNumber: account_number,
      sequence: sequence,
      mode: TW.Cosmos.Proto.BroadcastMode.SYNC,
      memo: keysignPayload.memo,
      messages: messages,
      fee: TW.Cosmos.Proto.Fee.create({
        gas: 20000000,
      }),
    });

    return TW.Cosmos.Proto.SigningInput.encode(input).finish();
  }

  const getSignedTransaction = async (
    walletCore: WalletCore,
    vaultHexPubKey: string,
    vaultHexChainCode: string,
    keysignPayload: KeysignPayloadType,
    signatures: { [key: string]: KeysignResponse }
  ) => {
    const inputData = getPreSignedInputData(walletCore, keysignPayload);
    const thorPublicKey = await getDerivedPublicKey(
      vaultHexPubKey,
      vaultHexChainCode,
      walletCore.CoinTypeExt.derivationPath(walletCore.CoinType.thorchain),
      false
    );
    const pubkeyData = Buffer.from(thorPublicKey, "hex");
    const publicKey = walletCore.PublicKey.createWithData(
      pubkeyData,
      walletCore.PublicKeyType.secp256k1
    );

    try {
      const hashes = walletCore.TransactionCompiler.preImageHashes(
        walletCore.CoinType.thorchain,
        inputData
      );
      const preSigningOutput =
        TW.TxCompiler.Proto.PreSigningOutput.decode(hashes);
      const allSignatures = walletCore.DataVector.create();
      const allPublicKeys = walletCore.DataVector.create();
      const signature = getSignatureWithRecoveryID(
        signatures[Buffer.from(preSigningOutput.dataHash).toString("base64")]
      );
      if (!publicKey.verify(signature, preSigningOutput.dataHash)) {
        console.error("fail to verify signature");
        return;
      }

      allSignatures.add(signature);
      allPublicKeys.add(publicKey.data());

      const compileWithSignature =
        walletCore.TransactionCompiler.compileWithSignatures(
          walletCore.CoinType.thorchain,
          inputData,
          allSignatures,
          allPublicKeys
        );
      const output = TW.Cosmos.Proto.SigningOutput.decode(compileWithSignature);
      const buffer = Buffer.from(
        JSON.parse(output.serialized).tx_bytes,
        "base64"
      );
      const hash = createHash("sha256").update(buffer).digest("hex");
      console.log("txHash:", hash);

      return output.serialized; // raw transaction
    } catch (error: any) {
      console.error(`fail to get signed transaction, error: ${error.message}`);
      return;
    }
  };

  const sendTransaction = async (
    vaultPublicKeyEcdsa: string,
    vaultLocalPartyId: string,
    vaultHexChainCode: string,
    fromPublicKey: string,
    fromAddress: string,
    toAddress: string,
    amount: string,
    passwd: string
  ) => {
    if (
      !vaultPublicKeyEcdsa ||
      !vaultLocalPartyId ||
      !vaultHexChainCode ||
      !fromAddress ||
      !toAddress ||
      !amount
    )
      return;

    try {
      const walletCore = await initWasm();
      const accountInfo = await (await getAccountInfo(fromAddress)).json();
      const accountNumber = accountInfo.result.value.account_number;
      const sequence = accountInfo.result.value.sequence;
      const thorchainspecific = new THORChainSpecific({
        account_number: accountNumber,
        sequence: sequence,
        fee: 2000000,
      });
      const coin = new Coin({
        chain: "THORChain",
        ticker: "RUNE",
        decimals: 8,
        is_native_token: true,
        hex_public_key: fromPublicKey,
        address: fromAddress,
      });
      const payload = new KeysignPayload({
        coin: coin,
        to_address: toAddress,
        to_amount: amount,
        thorchain_specific: thorchainspecific,
        vault_public_key_ecdsa: vaultPublicKeyEcdsa,
        vault_local_party_id: vaultLocalPartyId,
      });
      const txInputData = getPreSignedInputData(walletCore, payload);
      const preImageHashes = walletCore.TransactionCompiler.preImageHashes(
        walletCore.CoinType.thorchain,
        txInputData
      );
      const preSigningOutput =
        TW.TxCompiler.Proto.PreSigningOutput.decode(preImageHashes);
      const message = Buffer.from(preSigningOutput.dataHash).toString("base64");

      const sessionId = uuidv4();
      setSessionId(sessionId);
      const hexEncryptionKey = generateRandomHex(32);

      const keysignMsg = new KeysignMessage({
        session_id: sessionId,
        service_name: vaultLocalPartyId,
        keysign_payload: payload,
        encryption_key_hex: hexEncryptionKey,
        use_vultisig_relay: true,
      });

      const compressedData = await lzmaCompressData(keysignMsg.serialize());
      setQrString(
        `vultisig://vultisig.com?type=SignTransaction&vault=${vaultPublicKeyEcdsa}&jsonData=${Buffer.from(
          compressedData
        ).toString("base64")}`
      );

      const taskId = await (
        await signMessages(
          passwd,
          JSON.stringify({
            public_key: vaultPublicKeyEcdsa,
            messages: [message],
            session: sessionId,
            hex_encryption_key: hexEncryptionKey,
            derive_path: walletCore.CoinTypeExt.derivationPath(
              walletCore.CoinType.thorchain
            ),
            is_ecdsa: true,
            vault_password: "",
          })
        )
      ).json();
      goToStep(2);
      setTimeout(() => {
        signTransaction(
          taskId,
          walletCore,
          vaultPublicKeyEcdsa,
          vaultHexChainCode,
          payload
        );
      }, 20000);
    } catch (error) {
      console.error("Error sending transaction:", error);
    }
  };

  const signTransaction = async (
    taskId: string,
    walletCore: WalletCore,
    vaultPublicKeyEcdsa: string,
    vaultHexChainCode: string,
    payload: KeysignPayloadType
  ) => {
    try {
      const resp = await getSignResult(taskId);
      if (resp.status !== 200) {
        console.error("Invalid Task Id");
        setTimeout(() => {
          signTransaction(
            taskId,
            walletCore,
            vaultPublicKeyEcdsa,
            vaultHexChainCode,
            payload
          );
        }, 5000);
        return;
      }
      const result = await resp.json();
      if (
        result === "Task is still in progress" ||
        result === "task state is invalid"
      ) {
        setTimeout(() => {
          signTransaction(
            taskId,
            walletCore,
            vaultPublicKeyEcdsa,
            vaultHexChainCode,
            payload
          );
        }, 5000);
        return;
      }
      const signatures = JSON.parse(Buffer.from(result, "base64").toString());
      const signedTx = await getSignedTransaction(
        walletCore,
        vaultPublicKeyEcdsa,
        vaultHexChainCode,
        payload,
        signatures
      );
      if (signedTx) {
        broadcastSignedTransaction(signedTx);
        setStatus("done");
      }
    } catch (error) {
      console.error("Error signing transaction:", error);
    }
  };

  return (
    <div className="my-16 flex flex-col items-center justify-center">
      {currentStep === 1 && <StepOne sendTransaction={sendTransaction} />}
      {currentStep === 2 && (
        <StepTwo
          uniqueStrings={uniqueStrings}
          setUniqueStrings={setUniqueStrings}
          goToStep={goToStep}
          qrCodeString={qrString}
          session_id={sessionId}
        />
      )}
      {currentStep === 3 && <StepThree status={status} />}
    </div>
  );
};

export default TokenSend;
